package ip

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/core/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/snowflake"
	"github.com/HappyLadySauce/errors"
	"k8s.io/klog/v2"
)

// SyncAllFromConfigFiles 从 WireGuard 根目录下的所有配置文件同步 Peer 和 IP 分配信息到数据库
func SyncAllFromConfigFiles(ctx context.Context, storeFactory store.Factory) error {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		klog.V(1).InfoS("WireGuard config not found, skipping sync")
		return nil
	}

	// 扫描所有配置文件
	configFiles, err := scanConfigFiles(cfg.WireGuard.RootDir)
	if err != nil {
		klog.V(1).InfoS("Failed to scan config files", "error", err)
		return errors.Wrap(err, "failed to scan config files")
	}

	if len(configFiles) == 0 {
		klog.V(1).InfoS("No config files found, skipping sync")
		return nil
	}

	klog.V(1).InfoS("Starting sync from config files", "files", len(configFiles))

	// 统计信息
	stats := struct {
		peersCreated   int
		peersUpdated   int
		peersDisabled  int
		ipAllocsSynced int
		poolsCreated   int
		skipped        int
	}{}

	// 获取数据库中所有活跃的 Peer（用于对比）
	dbPeers, _, err := storeFactory.WGPeers().ListPeers(ctx, store.WGPeerListOptions{
		Status: model.WGPeerStatusActive,
		Limit:  10000, // 获取所有活跃的 Peer
	})
	if err != nil {
		klog.V(1).InfoS("Failed to list peers from database", "error", err)
		return errors.Wrap(err, "failed to list peers from database")
	}

	// 构建 PublicKey 到 Peer 的映射
	dbPeerMap := make(map[string]*model.WGPeer)
	for _, peer := range dbPeers {
		dbPeerMap[peer.ClientPublicKey] = peer
	}

	// 收集所有配置文件中的 Peer
	configPeerMap := make(map[string]*wireguard.ServerPeerConfig)
	configPeerFiles := make(map[string]string) // PublicKey -> config file path

	// 遍历所有配置文件
	for _, configPath := range configFiles {
		configManager := wireguard.NewServerConfigManager(configPath, cfg.WireGuard.ApplyMethod)
		serverConfig, err := configManager.ReadServerConfig()
		if err != nil {
			klog.V(1).InfoS("Failed to read config file", "path", configPath, "error", err)
			continue
		}

		// 收集该配置文件中的所有 Peer
		for _, peer := range serverConfig.Peers {
			if peer.PublicKey == "" {
				continue
			}
			configPeerMap[peer.PublicKey] = peer
			configPeerFiles[peer.PublicKey] = configPath
		}
	}

	// 1. 处理配置文件中存在但数据库中没有的 Peer（创建新 Peer）
	for publicKey, configPeer := range configPeerMap {
		if _, exists := dbPeerMap[publicKey]; !exists {
			// 创建外部添加的 Peer
			if err := createExternalPeer(ctx, storeFactory, configPeer, configPeerFiles[publicKey], cfg); err != nil {
				klog.V(1).InfoS("Failed to create external peer", "publicKey", publicKey[:10]+"...", "error", err)
				stats.skipped++
				continue
			}
			stats.peersCreated++
		}
	}

	// 2. 处理数据库中存在但配置文件中没有的 Peer（标记为 disabled）
	for publicKey, dbPeer := range dbPeerMap {
		if _, exists := configPeerMap[publicKey]; !exists {
			// 将 Peer 状态更新为 disabled
			dbPeer.Status = model.WGPeerStatusDisabled
			if err := storeFactory.WGPeers().UpdatePeer(ctx, dbPeer); err != nil {
				klog.V(1).InfoS("Failed to disable peer", "peerID", dbPeer.ID, "error", err)
				stats.skipped++
				continue
			}
			stats.peersDisabled++
			klog.V(1).InfoS("Disabled peer not found in config files", "peerID", dbPeer.ID, "publicKey", publicKey[:10]+"...")
		}
	}

	// 3. 同步所有配置文件中的 Peer 的 IP 分配信息
	for _, configPath := range configFiles {
		configManager := wireguard.NewServerConfigManager(configPath, cfg.WireGuard.ApplyMethod)
		serverConfig, err := configManager.ReadServerConfig()
		if err != nil {
			klog.V(1).InfoS("Failed to read config file for IP sync", "path", configPath, "error", err)
			continue
		}

		for _, peer := range serverConfig.Peers {
			if peer.PublicKey == "" || peer.AllowedIPs == "" {
				continue
			}

			// 同步该 Peer 的 IP 分配信息
			if synced, err := syncPeerIPAllocation(ctx, storeFactory, peer); err != nil {
				klog.V(1).InfoS("Failed to sync IP allocation", "publicKey", peer.PublicKey[:10]+"...", "error", err)
				stats.skipped++
			} else if synced {
				stats.ipAllocsSynced++
			}

			// 更新 Peer 状态为 active（如果之前被禁用）
			dbPeer, err := storeFactory.WGPeers().GetPeerByPublicKey(ctx, peer.PublicKey)
			if err == nil && dbPeer != nil && dbPeer.Status != model.WGPeerStatusActive {
				dbPeer.Status = model.WGPeerStatusActive
				if err := storeFactory.WGPeers().UpdatePeer(ctx, dbPeer); err == nil {
					stats.peersUpdated++
				}
			}
		}
	}

	klog.V(1).InfoS("Sync completed",
		"peersCreated", stats.peersCreated,
		"peersUpdated", stats.peersUpdated,
		"peersDisabled", stats.peersDisabled,
		"ipAllocsSynced", stats.ipAllocsSynced,
		"poolsCreated", stats.poolsCreated,
		"skipped", stats.skipped)

	return nil
}

// scanConfigFiles 扫描 WireGuard 根目录下的所有 .conf 文件（排除 .backup 文件）
func scanConfigFiles(rootDir string) ([]string, error) {
	var configFiles []string

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to read root directory")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// 只处理 .conf 文件，排除 .backup 文件
		if strings.HasSuffix(name, ".conf") && !strings.HasSuffix(name, ".backup") {
			configPath := filepath.Join(rootDir, name)
			configFiles = append(configFiles, configPath)
		}
	}

	return configFiles, nil
}

// inferCIDRFromIP 根据 IP 地址推断 CIDR（默认使用 /24 子网）
// 例如: "100.100.100.2" -> "100.100.100.0/24"
func inferCIDRFromIP(ipStr string) (string, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", ipStr)
	}

	ipv4 := ip.To4()
	if ipv4 == nil {
		return "", fmt.Errorf("not an IPv4 address: %s", ipStr)
	}

	// 使用 /24 子网（最常见的 WireGuard 配置）
	// 将最后一个字节设为 0
	networkIP := make(net.IP, 4)
	copy(networkIP, ipv4)
	networkIP[3] = 0

	cidr := fmt.Sprintf("%s/24", networkIP.String())
	return cidr, nil
}

// findOrCreateIPPool 查找或创建 IP Pool
// 返回 Pool 和是否为新创建的标志
func findOrCreateIPPool(ctx context.Context, storeFactory store.Factory, ipAddr string, cfg *config.Config) (*model.IPPool, bool, error) {
	// 获取所有活跃的 IP Pools
	pools, _, err := storeFactory.IPPools().ListIPPools(ctx, store.IPPoolListOptions{
		Status: model.IPPoolStatusActive,
	})
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to list IP pools")
	}

	// 查找匹配的 Pool
	for _, pool := range pools {
		if ipInCIDR(ipAddr, pool.CIDR) {
			return pool, false, nil
		}
	}

	// 没有找到匹配的 Pool，推断 CIDR 并创建新 Pool
	inferredCIDR, err := inferCIDRFromIP(ipAddr)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to infer CIDR from IP")
	}

	// 检查是否已存在相同 CIDR 的 Pool（可能状态不是 active）
	existingPool, err := storeFactory.IPPools().GetIPPoolByCIDR(ctx, inferredCIDR)
	if err == nil && existingPool != nil {
		// 如果 Pool 存在但状态不是 active，更新为 active
		if existingPool.Status != model.IPPoolStatusActive {
			existingPool.Status = model.IPPoolStatusActive
			if err := storeFactory.IPPools().UpdateIPPool(ctx, existingPool); err != nil {
				klog.V(1).InfoS("Failed to update pool status", "poolID", existingPool.ID, "error", err)
			}
		}
		return existingPool, false, nil
	}

	// 创建新 Pool
	poolID, err := snowflake.GenerateID()
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to generate pool ID")
	}

	// 生成 Pool 名称（使用 CIDR 作为名称的一部分，确保唯一性）
	poolName := fmt.Sprintf("auto-%s", strings.ReplaceAll(inferredCIDR, "/", "-"))

	// 使用全局配置的默认值
	routes := inferredCIDR
	dns := cfg.WireGuard.DNS
	endpoint := cfg.WireGuard.Endpoint

	pool := &model.IPPool{
		ID:          poolID,
		Name:        poolName,
		CIDR:        inferredCIDR,
		Routes:      routes,
		DNS:         dns,
		Endpoint:    endpoint,
		Description: "Auto-created from config sync",
		Status:      model.IPPoolStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := storeFactory.IPPools().CreateIPPool(ctx, pool); err != nil {
		return nil, false, errors.Wrap(err, "failed to create IP pool")
	}

	klog.V(1).InfoS("Auto-created IP pool", "poolID", poolID, "cidr", inferredCIDR, "name", poolName)
	return pool, true, nil
}

// createExternalPeer 创建外部添加的 Peer（配置文件中存在但数据库中没有）
func createExternalPeer(ctx context.Context, storeFactory store.Factory, configPeer *wireguard.ServerPeerConfig, configPath string, cfg *config.Config) error {
	if configPeer.PublicKey == "" || configPeer.AllowedIPs == "" {
		return fmt.Errorf("invalid peer config: missing PublicKey or AllowedIPs")
	}

	// 提取 IP 地址
	ipAddr := extractIPFromCIDR(configPeer.AllowedIPs)
	if ipAddr == "" {
		return fmt.Errorf("failed to extract IP from AllowedIPs: %s", configPeer.AllowedIPs)
	}

	// 查找或创建 IP Pool
	pool, created, err := findOrCreateIPPool(ctx, storeFactory, ipAddr, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to find or create IP pool")
	}
	if created {
		klog.V(1).InfoS("Auto-created IP pool for external peer", "poolID", pool.ID, "cidr", pool.CIDR)
	}

	// 查找第一个管理员用户作为默认用户
	users, _, err := storeFactory.Users().ListUsers(ctx, store.UserListOptions{
		Role:   model.UserRoleAdmin,
		Status: model.UserStatusActive,
		Limit:  1,
	})
	if err != nil || len(users) == 0 {
		return fmt.Errorf("no admin user found for external peer")
	}
	defaultUserID := users[0].ID

	// 确定 DeviceName
	deviceName := configPeer.Comment
	if deviceName == "" {
		// 使用 PublicKey 的前 8 个字符作为设备名
		if len(configPeer.PublicKey) > 8 {
			deviceName = "[External] " + configPeer.PublicKey[:8]
		} else {
			deviceName = "[External] " + configPeer.PublicKey
		}
	} else {
		deviceName = "[External] " + deviceName
	}

	// 生成 Peer ID
	peerID, err := snowflake.GenerateID()
	if err != nil {
		return errors.Wrap(err, "failed to generate peer ID")
	}

	// 格式化 IP 为 CIDR
	clientIPCIDR, err := FormatIPAsCIDR(ipAddr)
	if err != nil {
		return errors.Wrap(err, "failed to format IP as CIDR")
	}

	// 创建 Peer（PrivateKey 使用占位符，因为配置文件中只有 PublicKey）
	// 注意：数据库约束要求 not null，所以使用占位符值
	peer := &model.WGPeer{
		ID:                  peerID,
		UserID:              defaultUserID,
		DeviceName:          deviceName,
		ClientPrivateKey:    "[external-managed]", // 外部管理的 Peer 没有 PrivateKey
		ClientPublicKey:     configPeer.PublicKey,
		ClientIP:            clientIPCIDR,
		AllowedIPs:          configPeer.AllowedIPs,
		DNS:                 "",
		Endpoint:            "",
		PersistentKeepalive: configPeer.PersistentKeepalive,
		Status:              model.WGPeerStatusActive,
		IPPoolID:            pool.ID,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := storeFactory.WGPeers().CreatePeer(ctx, peer); err != nil {
		return errors.Wrap(err, "failed to create peer")
	}

	// 创建 IP 分配记录
	allocationID, err := snowflake.GenerateID()
	if err != nil {
		return errors.Wrap(err, "failed to generate allocation ID")
	}

	allocation := &model.IPAllocation{
		ID:        allocationID,
		IPPoolID:  pool.ID,
		PeerID:    peerID,
		IPAddress: ipAddr,
		Status:    model.IPAllocationStatusAllocated,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := storeFactory.IPAllocations().CreateIPAllocation(ctx, allocation); err != nil {
		// 回滚：删除 Peer
		_ = storeFactory.WGPeers().DeletePeer(ctx, peerID)
		return errors.Wrap(err, "failed to create IP allocation")
	}

	klog.V(1).InfoS("Created external peer", "peerID", peerID, "publicKey", configPeer.PublicKey[:10]+"...", "ip", ipAddr, "poolID", pool.ID)
	return nil
}

// syncPeerIPAllocation 同步单个 Peer 的 IP 分配信息
func syncPeerIPAllocation(ctx context.Context, storeFactory store.Factory, configPeer *wireguard.ServerPeerConfig) (bool, error) {
	if configPeer.PublicKey == "" || configPeer.AllowedIPs == "" {
		return false, nil
	}

	// 提取 IP 地址
	ipAddr := extractIPFromCIDR(configPeer.AllowedIPs)
	if ipAddr == "" {
		return false, fmt.Errorf("failed to extract IP from AllowedIPs: %s", configPeer.AllowedIPs)
	}

	// 通过 PublicKey 查找 Peer
	dbPeer, err := storeFactory.WGPeers().GetPeerByPublicKey(ctx, configPeer.PublicKey)
	if err != nil || dbPeer == nil {
		// Peer 不存在，跳过（应该已经在 createExternalPeer 中处理）
		return false, nil
	}

	// 获取配置以查找或创建 IP Pool
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return false, fmt.Errorf("WireGuard config not found")
	}

	// 查找或创建 IP Pool
	pool, _, err := findOrCreateIPPool(ctx, storeFactory, ipAddr, cfg)
	if err != nil {
		return false, errors.Wrap(err, "failed to find or create IP pool")
	}

	// 检查 IPAllocation 是否存在
	existing, _ := storeFactory.IPAllocations().GetIPAllocationByPeerID(ctx, dbPeer.ID)
	if existing != nil {
		// 已存在，检查是否需要更新
		if existing.IPAddress != ipAddr || existing.IPPoolID != pool.ID {
			existing.IPAddress = ipAddr
			existing.IPPoolID = pool.ID
			existing.Status = model.IPAllocationStatusAllocated
			existing.UpdatedAt = time.Now()
			if err := storeFactory.IPAllocations().UpdateIPAllocation(ctx, existing); err != nil {
				return false, errors.Wrap(err, "failed to update IP allocation")
			}
			return true, nil
		}
		return false, nil
	}

	// 创建新的 IPAllocation 记录
	allocationID, err := snowflake.GenerateID()
	if err != nil {
		return false, errors.Wrap(err, "failed to generate allocation ID")
	}

	allocation := &model.IPAllocation{
		ID:        allocationID,
		IPPoolID:  pool.ID,
		PeerID:    dbPeer.ID,
		IPAddress: ipAddr,
		Status:    model.IPAllocationStatusAllocated,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := storeFactory.IPAllocations().CreateIPAllocation(ctx, allocation); err != nil {
		return false, errors.Wrap(err, "failed to create IP allocation")
	}

	// 更新 Peer 的 IPPoolID（如果不同）
	if dbPeer.IPPoolID != pool.ID {
		dbPeer.IPPoolID = pool.ID
		clientIPCIDR, err := FormatIPAsCIDR(ipAddr)
		if err == nil {
			dbPeer.ClientIP = clientIPCIDR
		}
		if err := storeFactory.WGPeers().UpdatePeer(ctx, dbPeer); err != nil {
			klog.V(1).InfoS("Failed to update peer IPPoolID", "peerID", dbPeer.ID, "error", err)
		}
	}

	return true, nil
}

// extractIPFromCIDR 从 CIDR 格式提取 IP 地址
// 例如: "100.100.100.2/32" -> "100.100.100.2"
func extractIPFromCIDR(cidr string) string {
	// 处理逗号分隔的多个 CIDR，取第一个
	parts := strings.Split(cidr, ",")
	if len(parts) > 0 {
		cidr = strings.TrimSpace(parts[0])
	}

	parts = strings.Split(cidr, "/")
	if len(parts) == 0 {
		return ""
	}
	ipStr := strings.TrimSpace(parts[0])

	// 验证 IP 格式
	if net.ParseIP(ipStr) == nil {
		return ""
	}

	return ipStr
}

// ipInCIDR 检查 IP 是否在 CIDR 范围内
func ipInCIDR(ipStr, cidr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	return ipNet.Contains(ip)
}

// SyncIPAllocationsFromConfig 保持向后兼容的旧函数名（已废弃，使用 SyncAllFromConfigFiles）
// Deprecated: Use SyncAllFromConfigFiles instead
func SyncIPAllocationsFromConfig(ctx context.Context, storeFactory store.Factory) error {
	return SyncAllFromConfigFiles(ctx, storeFactory)
}
