package sqlite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gorm.io/gorm"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/db"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/options"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/passwd"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/snowflake"
	"github.com/HappyLadySauce/errors"
	"k8s.io/klog/v2"
)

type datastore struct {
	db *gorm.DB
}

func (ds *datastore) Users() store.UserStore {
	return newUsers(ds)
}

func (ds *datastore) WGPeers() store.WGPeerStore {
	return newWGPeers(ds)
}

func (ds *datastore) IPPools() store.IPPoolStore {
	return newIPPools(ds)
}

func (ds *datastore) IPAllocations() store.IPAllocationStore {
	return newIPAllocations(ds)
}

func (ds *datastore) Close() error {
	sqlDB, err := ds.db.DB()
	if err != nil {
		return errors.Wrap(err, "failed to get sql db")
	}

	return sqlDB.Close()
}

var (
	sqliteFactory store.Factory
	once          sync.Once
)

func GetSqliteFactoryOr(opts *options.SqliteOptions) (store.Factory, error) {
	// If opts is nil, use default options
	if opts == nil {
		opts = options.NewSqliteOptions()
	}

	var err error
	var dbIns *gorm.DB
	once.Do(func() {
		// Ensure the parent directory of the database file exists
		// SQLite will not create parent directories automatically
		dbPath := opts.DataSourceName

		// Get the absolute path to handle both relative and absolute paths
		absPath, absErr := filepath.Abs(dbPath)
		if absErr != nil {
			// If we can't get absolute path, use the original path
			absPath = dbPath
		}

		// Get the parent directory
		parentDir := filepath.Dir(absPath)

		// Create parent directory if it doesn't exist
		if err = os.MkdirAll(parentDir, 0755); err != nil {
			klog.V(1).InfoS("failed to create database parent directory", "parentDir", parentDir, "dataSource", opts.DataSourceName, "error", err)
			err = errors.Wrap(err, "failed to create database parent directory")
			return
		}

		klog.V(2).InfoS("database parent directory ensured", "parentDir", parentDir, "dataSource", opts.DataSourceName)

		dbOpts := &db.Options{
			DataSourceName: opts.DataSourceName,
		}
		dbIns, err = db.New(dbOpts)
		if err != nil {
			// Preserve the original error with full context
			klog.V(1).InfoS("failed to create sqlite database", "dataSource", opts.DataSourceName, "error", err)
			err = errors.Wrap(err, "failed to create sqlite db with data source")
			return
		}

		// Handle custom migrations for existing tables before AutoMigrate
		// This is needed because SQLite doesn't allow adding NOT NULL columns to existing tables
		if migrateErr := handleCustomMigrations(dbIns); migrateErr != nil {
			klog.V(1).InfoS("failed to run custom migrations", "dataSource", opts.DataSourceName, "error", migrateErr)
			// Log error but continue, as AutoMigrate might still work for new tables
		}

		// Auto migrate database schema
		if err = dbIns.AutoMigrate(
			&model.User{},
			&model.WGPeer{},
			&model.IPPool{},
			&model.IPAllocation{},
		); err != nil {
			klog.V(1).InfoS("failed to auto migrate database schema", "dataSource", opts.DataSourceName, "error", err)
			err = errors.Wrap(err, "failed to auto migrate database schema")
			return
		}

		klog.V(1).InfoS("database schema migrated successfully", "dataSource", opts.DataSourceName)

		// Initialize default admin user if not exists
		if initErr := initializeDefaultAdmin(dbIns); initErr != nil {
			klog.Errorf("Failed to initialize default admin user: %+v", initErr)
			// Don't fail the entire initialization, just log the error
			// Admin user can be created manually later
		}

		sqliteFactory = &datastore{dbIns}
	})

	if sqliteFactory == nil {
		if err != nil {
			// Return the wrapped error directly to preserve the full error chain
			klog.V(1).InfoS("failed to get sqlite factory", "dataSource", opts.DataSourceName, "error", err)
			return nil, errors.Wrap(err, "failed to get sqlite factory")
		}
		// If err is nil but sqliteFactory is nil, create a new error
		klog.V(1).InfoS("sqlite factory is nil but no error was returned", "dataSource", opts.DataSourceName)
		return nil, errors.New("failed to get sqlite factory: sqliteFactory is nil but no error was returned")
	}

	return sqliteFactory, nil
}

// initializeDefaultAdmin 初始化默认管理员用户
// 如果数据库中不存在admin用户，则创建一个默认的admin用户
// 使用固定默认密码并保存到 pwd.txt 文件中
func initializeDefaultAdmin(db *gorm.DB) error {
	ctx := context.Background()
	adminUsername := "admin"

	// 检查是否已存在admin用户
	var existingUser model.User
	err := db.WithContext(ctx).Where("username = ?", adminUsername).First(&existingUser).Error
	if err == nil {
		// Admin用户已存在，跳过初始化
		klog.V(1).InfoS("admin user already exists, skipping initialization", "username", adminUsername)
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 其他数据库错误
		return errors.Wrap(err, "failed to check admin user existence")
	}

	// 使用固定默认密码
	password := "NexusPointWG"

	// 生成盐值
	salt, err := passwd.GenerateSalt()
	if err != nil {
		return errors.Wrap(err, "failed to generate salt")
	}

	// 生成密码哈希
	passwordHash, err := passwd.HashPassword(password, salt)
	if err != nil {
		return errors.Wrap(err, "failed to hash password")
	}

	// 生成用户ID
	userID, err := snowflake.GenerateID()
	if err != nil {
		return errors.Wrap(err, "failed to generate user ID")
	}

	// 创建admin用户
	adminUser := &model.User{
		ID:           userID,
		Username:     adminUsername,
		Nickname:     "Administrator",
		Avatar:       model.DefaultAvatarURL,
		Email:        "admin@gmail.com", // 默认邮箱（使用允许的域名以通过验证），管理员可以后续修改
		Salt:         salt,
		PasswordHash: passwordHash,
		Status:       model.UserStatusActive,
		Role:         model.UserRoleAdmin,
	}

	// 验证用户数据
	if errs := adminUser.Validate(); len(errs) != 0 {
		return fmt.Errorf("admin user validation failed: %v", errs.ToAggregate().Error())
	}

	// 保存到数据库
	if err := db.WithContext(ctx).Create(adminUser).Error; err != nil {
		return errors.Wrap(err, "failed to create admin user")
	}

	// 保存密码到 pwd.txt 文件（当前工作目录）
	pwdFile := "pwd.txt"
	if err := savePasswordToFile(pwdFile, adminUsername, password); err != nil {
		// 即使保存文件失败，也不影响用户创建，只记录警告
		// SECURITY: Never log passwords, even in error scenarios. Log files may be stored long-term
		// and have broader access than expected.
		klog.V(1).InfoS("failed to save password to file", "file", pwdFile, "error", err)
		klog.Warningf("Admin user created but password file save failed. Username: %s. Please check the file manually or reset the password.", adminUsername)
		// Note: The password is lost if file save fails. Admin will need to reset password or check file system.
	} else {
		klog.V(1).InfoS("admin user initialized successfully", "username", adminUsername, "passwordFile", pwdFile)
		klog.Infof("Default admin user created. Username: %s, Password saved to: %s", adminUsername, pwdFile)
	}

	return nil
}

// handleCustomMigrations handles custom migrations for existing tables.
// This is needed because SQLite doesn't allow adding NOT NULL columns to existing tables
// without default values. We need to add the column as nullable first, then update it.
func handleCustomMigrations(db *gorm.DB) error {
	// Check if ip_pools table exists and if CIDR column is missing
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='ip_pools'").Scan(&count).Error
	if err != nil {
		return errors.Wrap(err, "failed to check if ip_pools table exists")
	}

	if count > 0 {
		// Table exists, check if CIDR column exists
		var columns []struct {
			Name string `gorm:"column:name"`
		}
		err = db.Raw("PRAGMA table_info(ip_pools)").Scan(&columns).Error
		if err != nil {
			return errors.Wrap(err, "failed to get ip_pools table info")
		}

		hasCIDR := false
		for _, col := range columns {
			if col.Name == "cidr" {
				hasCIDR = true
				break
			}
		}

		if !hasCIDR {
			// Add CIDR column as nullable first
			klog.V(1).InfoS("adding cidr column to ip_pools table as nullable")
			if err := db.Exec("ALTER TABLE ip_pools ADD COLUMN cidr TEXT").Error; err != nil {
				// If column already exists or other error, log and continue
				klog.V(1).InfoS("failed to add cidr column (may already exist)", "error", err)
			} else {
				// Update existing records with default CIDR value
				// Use "100.100.100.0/24" as default CIDR (common example in codebase)
				defaultCIDR := "100.100.100.0/24"
				klog.V(1).InfoS("updating existing ip_pools records with default cidr value", "defaultCIDR", defaultCIDR)
				if err := db.Exec("UPDATE ip_pools SET cidr = ? WHERE cidr IS NULL OR cidr = ''", defaultCIDR).Error; err != nil {
					klog.V(1).InfoS("failed to update existing records with default cidr", "error", err)
					// Continue anyway, AutoMigrate will handle it
				} else {
					klog.V(1).InfoS("successfully updated existing ip_pools records with default cidr")
				}
			}
		} else {
			// CIDR column exists, but may have NULL values
			// Update any NULL or empty CIDR values with default
			defaultCIDR := "100.100.100.0/24"
			klog.V(1).InfoS("checking for NULL or empty cidr values in ip_pools")
			var nullCount int64
			if err := db.Raw("SELECT COUNT(*) FROM ip_pools WHERE cidr IS NULL OR cidr = ''").Scan(&nullCount).Error; err == nil && nullCount > 0 {
				klog.V(1).InfoS("updating NULL or empty cidr values with default", "count", nullCount, "defaultCIDR", defaultCIDR)
				if err := db.Exec("UPDATE ip_pools SET cidr = ? WHERE cidr IS NULL OR cidr = ''", defaultCIDR).Error; err != nil {
					klog.V(1).InfoS("failed to update NULL cidr values", "error", err)
				} else {
					klog.V(1).InfoS("successfully updated NULL cidr values with default")
				}
			}
		}
	}

	return nil
}

// savePasswordToFile 将密码保存到文件
// 如果文件已存在，则覆盖
func savePasswordToFile(filename, username, password string) error {
	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get current working directory")
	}

	filePath := filepath.Join(workDir, filename)

	// 准备文件内容
	content := "Default Admin User Credentials\n"
	content += "==============================\n"
	content += fmt.Sprintf("Username: %s\n", username)
	content += fmt.Sprintf("Password: %s\n", password)
	content += "\n"
	content += "Please change the password after first login!\n"
	content += "Please keep this file secure and delete it after recording the password.\n"

	// 写入文件（覆盖模式）
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		return errors.Wrap(err, "failed to write password file")
	}

	return nil
}
