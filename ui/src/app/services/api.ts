
// Toggle this to switch between real backend and mock mode
// Set to false to connect to the real Go backend
const USE_MOCK = false;
const API_BASE = "/api/v1";

// ============================================================================
// Type Definitions (matching backend types)
// ============================================================================

// User types
export interface UserResponse {
  username: string;
  nickname: string;
  email: string;
  peer_count?: number;
}

export interface User {
  id: string;
  username: string;
  role: "admin" | "user";
  nickname?: string;
  email?: string;
}

export interface CreateUserRequest {
  username: string;
  nickname?: string;
  avatar?: string;
  email: string;
  password: string;
  role?: "user" | "admin";
  status?: "active" | "inactive" | "deleted";
}

export interface UpdateUserRequest {
  username?: string;
  nickname?: string;
  avatar?: string;
  email?: string;
  password?: string;
  status?: "active" | "inactive" | "deleted";
  role?: "user" | "admin";
}

export interface ChangePasswordRequest {
  oldPassword: string;
  newPassword: string;
}

// WireGuard Peer types
export interface WGPeerResponse {
  id: string;
  user_id: string;
  username?: string;
  device_name: string;
  client_public_key: string;
  client_private_key?: string; // Optional, sensitive information
  client_ip: string;
  allowed_ips: string; // Comma-separated CIDRs, not array
  dns?: string;
  endpoint?: string;
  persistent_keepalive: number;
  status: string;
  ip_pool_id?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateWGPeerRequest {
  username?: string; // Admin can specify username, regular user uses their own
  user_id?: string; // Deprecated, use username instead
  device_name: string;
  client_ip?: string;
  ip_pool_id?: string;
  allowed_ips?: string;
  dns?: string;
  endpoint?: string;
  persistent_keepalive?: number;
  client_private_key?: string;
}

export interface UpdateWGPeerRequest {
  device_name?: string;
  client_ip?: string;
  ip_pool_id?: string;
  client_private_key?: string;
  allowed_ips?: string;
  dns?: string;
  endpoint?: string;
  persistent_keepalive?: number;
  status?: "active" | "disabled";
}

// IP Pool types
export interface IPPoolResponse {
  id: string;
  name: string;
  cidr: string;
  routes?: string;
  dns?: string;
  endpoint?: string;
  description?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface CreateIPPoolRequest {
  name: string;
  cidr: string;
  routes?: string;
  dns?: string;
  endpoint?: string;
  description?: string;
}

export interface UpdateIPPoolRequest {
  name?: string;
  routes?: string;
  dns?: string;
  endpoint?: string;
  description?: string;
  status?: "active" | "disabled";
}

export interface AvailableIPsResponse {
  ip_pool_id: string;
  cidr: string;
  ips: string[];
  total: number;
}

// Server Configuration types
export interface GetServerConfigResponse {
  address: string;      // Server tunnel IP, e.g., "100.100.100.1/24"
  listen_port: number;  // Listening port, e.g., 51820
  private_key: string;  // Server private key (sensitive)
  mtu: number;          // MTU, e.g., 1420
  post_up: string;      // PostUp command
  post_down: string;    // PostDown command
  public_key: string;   // Server public key (calculated from private key)
  server_ip: string;   // Server public IP for client endpoint (optional, auto-detected if empty)
  dns?: string;         // DNS server for client configs (optional, comma-separated IP addresses)
}

export interface UpdateServerConfigRequest {
  address?: string;
  listen_port?: number;
  private_key?: string;
  mtu?: number;
  post_up?: string;
  post_down?: string;
  server_ip?: string;   // Server public IP for client endpoint (optional, auto-detected if empty)
  dns?: string;         // DNS server for client configs (optional, comma-separated IP addresses)
}

// Pagination types
export interface ListResponse<T> {
  total: number;
  items: T[];
}

// List filter options
export interface PeerListOptions {
  offset?: number;
  limit?: number;
  user_id?: string;
  status?: string;
  ip_pool_id?: string;
  device_name?: string;
}

export interface IPPoolListOptions {
  offset?: number;
  limit?: number;
  status?: string;
}

export interface UserListOptions {
  offset?: number;
  limit?: number;
  username?: string;
  email?: string;
  role?: string;
  status?: string;
}

// ============================================================================
// Mock Data
// ============================================================================

let mockUsers: User[] = [
  { id: "1", username: "admin", role: "admin", nickname: "Admin", email: "admin@example.com" },
  { id: "2", username: "user", role: "user", nickname: "User", email: "user@example.com" },
];

let mockPeers: WGPeerResponse[] = [
  {
    id: "peer-1",
    user_id: "1",
    username: "admin",
    device_name: "iPhone",
    client_public_key: "Jw8.../2A=",
    client_ip: "10.0.0.2",
    allowed_ips: "10.0.0.2/32",
    dns: "1.1.1.1",
    endpoint: "vpn.example.com:51820",
    persistent_keepalive: 25,
    status: "active",
    ip_pool_id: "pool-1",
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "peer-2",
    user_id: "2",
    username: "user",
    device_name: "MacBook Pro",
    client_public_key: "K2d.../9C=",
    client_ip: "10.0.0.3",
    allowed_ips: "10.0.0.3/32",
    dns: "1.1.1.1",
    endpoint: "vpn.example.com:51820",
    persistent_keepalive: 25,
    status: "disabled",
    ip_pool_id: "pool-1",
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
];

let mockIPPools: IPPoolResponse[] = [
  {
    id: "pool-1",
    name: "Main Pool",
    cidr: "10.0.0.0/24",
    routes: "10.0.0.0/24, 192.168.1.0/24",
    dns: "1.1.1.1, 223.5.5.5",
    endpoint: "118.24.41.142:51820",
    description: "Main IP pool",
    status: "active",
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
];

// ============================================================================
// Helper Functions
// ============================================================================

// Helper to simulate delay
const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

// Helper for Auth Headers
const getHeaders = () => {
    const token = localStorage.getItem("token");
    return {
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : "",
    };
};

// Helper for handling responses
const handleResponse = async (res: Response) => {
    if (res.status === 401) {
        localStorage.removeItem("token");
        window.location.href = "/"; // Force login
        throw new Error("Unauthorized");
    }
    if (!res.ok) {
        const text = await res.text();
        try {
            const json = JSON.parse(text);
            throw new Error(json.error || json.message || res.statusText);
        } catch {
            throw new Error(text || res.statusText);
        }
    }
    return res.json();
};

// Helper to build query string from options
const buildQueryString = (options: Record<string, any>): string => {
  const params = new URLSearchParams();
  Object.entries(options).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      params.append(key, String(value));
    }
  });
  const query = params.toString();
  return query ? `?${query}` : "";
};

// Helper to parse JWT token (basic parsing without verification)
const parseJWT = (token: string): { user_id?: string; username?: string; role?: string } => {
  try {
    const base64Url = token.split(".")[1];
    const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split("")
        .map((c) => "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2))
        .join("")
    );
    return JSON.parse(jsonPayload);
  } catch {
    return {};
  }
};

// ============================================================================
// API Implementation
// ============================================================================

export const api = {
  // ==========================================================================
  // Authentication
  // ==========================================================================
  auth: {
    login: async (username: string, password: string): Promise<{ token: string; user: User }> => {
      if (USE_MOCK) {
        await delay(500);
        if (username === "admin" && password === "admin") {
          const user: User = { id: "1", username: "admin", role: "admin" };
          const token = "mock-jwt-token-admin";
          localStorage.setItem("token", token);
          return { token, user };
        }
        if (username === "user" && password === "user") {
          const user: User = { id: "2", username: "user", role: "user" };
          const token = "mock-jwt-token-user";
          localStorage.setItem("token", token);
          return { token, user };
        }
        throw new Error("Invalid credentials");
      }

      const res = await fetch(`${API_BASE}/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
      });

      const data = await handleResponse(res);
      if (!data.token) {
        throw new Error("No token in response");
      }

      // Save token
      localStorage.setItem("token", data.token);

      // Parse user info from token
      const claims = parseJWT(data.token);
      const user: User = {
        id: claims.user_id || "",
        username: claims.username || username,
        role: (claims.role as "admin" | "user") || "user",
      };

      return { token: data.token, user };
    },
  },

  // ==========================================================================
  // WireGuard Peers
  // ==========================================================================
  wg: {
    listPeers: async (options?: PeerListOptions): Promise<ListResponse<WGPeerResponse>> => {
      if (USE_MOCK) {
        await delay(300);
        let filtered = [...mockPeers];

        if (options?.user_id) {
          filtered = filtered.filter((p) => p.user_id === options.user_id);
      }
        if (options?.status) {
          filtered = filtered.filter((p) => p.status === options.status);
        }
        if (options?.device_name) {
          filtered = filtered.filter((p) => p.device_name.includes(options.device_name!));
        }

        const offset = options?.offset || 0;
        const limit = options?.limit || 20;
        const total = filtered.length;
        const items = filtered.slice(offset, offset + limit);

        return { total, items };
      }

      const query = buildQueryString(options || {});
      const res = await fetch(`${API_BASE}/wg/peers${query}`, {
        headers: getHeaders(),
      });
      return handleResponse(res);
    },

    getPeer: async (id: string): Promise<WGPeerResponse> => {
      if (USE_MOCK) {
        await delay(200);
        const peer = mockPeers.find((p) => p.id === id);
        if (!peer) throw new Error("Peer not found");
        return peer;
      }

      const res = await fetch(`${API_BASE}/wg/peers/${id}`, {
        headers: getHeaders(),
      });
      return handleResponse(res);
    },

    createPeer: async (data: CreateWGPeerRequest): Promise<WGPeerResponse> => {
      if (USE_MOCK) {
        await delay(400);
        const newPeer: WGPeerResponse = {
          id: `peer-${Date.now()}`,
          user_id: data.user_id || "1",
          device_name: data.device_name,
          client_public_key: "MOCK_KEY_" + Math.random().toString(36).substring(7),
          client_ip: data.client_ip || `10.0.0.${mockPeers.length + 2}`,
          allowed_ips: data.allowed_ips || `${data.client_ip || `10.0.0.${mockPeers.length + 2}`}/32`,
          dns: data.dns,
          endpoint: data.endpoint,
          persistent_keepalive: data.persistent_keepalive || 25,
          status: "active",
          ip_pool_id: data.ip_pool_id,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        mockPeers.push(newPeer);
        return newPeer;
      }

      const res = await fetch(`${API_BASE}/wg/peers`, {
        method: "POST",
        headers: getHeaders(),
        body: JSON.stringify(data),
      });
      return handleResponse(res);
    },

    updatePeer: async (id: string, data: UpdateWGPeerRequest): Promise<WGPeerResponse> => {
      if (USE_MOCK) {
        await delay(300);
        const peerIndex = mockPeers.findIndex((p) => p.id === id);
        if (peerIndex === -1) throw new Error("Peer not found");

        const updatedPeer = {
          ...mockPeers[peerIndex],
          ...data,
          updated_at: new Date().toISOString(),
        };
        mockPeers[peerIndex] = updatedPeer;
        return updatedPeer;
      }

      const res = await fetch(`${API_BASE}/wg/peers/${id}`, {
        method: "PUT",
        headers: getHeaders(),
        body: JSON.stringify(data),
      });
      return handleResponse(res);
    },

    deletePeer: async (id: string): Promise<void> => {
      if (USE_MOCK) {
        await delay(300);
        mockPeers = mockPeers.filter((p) => p.id !== id);
        return;
      }

       const res = await fetch(`${API_BASE}/wg/peers/${id}`, { 
           method: "DELETE",
        headers: getHeaders(),
       });
       if (!res.ok) throw new Error("Failed to delete peer");
    },

    downloadConfig: async (id: string): Promise<string> => {
        if (USE_MOCK) {
            await delay(500);
            return `[Interface]\nPrivateKey = MOCK_PRIVATE_KEY\nAddress = 10.0.0.5/32\nDNS = 1.1.1.1\n\n[Peer]\nPublicKey = SERVER_PUBLIC_KEY\nEndpoint = vpn.example.com:51820\nAllowedIPs = 0.0.0.0/0`;
        }

        const res = await fetch(`${API_BASE}/wg/peers/${id}/config`, {
        headers: getHeaders(),
        });
        if (!res.ok) throw new Error("Failed to download config");
        return res.text();
    },

    // ========================================================================
    // IP Pools
    // ========================================================================
    listIPPools: async (options?: IPPoolListOptions): Promise<ListResponse<IPPoolResponse>> => {
       if (USE_MOCK) {
        await delay(300);
        let filtered = [...mockIPPools];

        if (options?.status) {
          filtered = filtered.filter((p) => p.status === options.status);
        }

        const offset = options?.offset || 0;
        const limit = options?.limit || 20;
        const total = filtered.length;
        const items = filtered.slice(offset, offset + limit);

        return { total, items };
      }

      const query = buildQueryString(options || {});
      const res = await fetch(`${API_BASE}/wg/ip-pools${query}`, {
        headers: getHeaders(),
      });
      return handleResponse(res);
    },

    createIPPool: async (data: CreateIPPoolRequest): Promise<IPPoolResponse> => {
      if (USE_MOCK) {
        await delay(400);
        const newPool: IPPoolResponse = {
          id: `pool-${Date.now()}`,
          name: data.name,
          cidr: data.cidr,
          routes: data.routes,
          dns: data.dns,
          endpoint: data.endpoint,
          description: data.description,
          status: "active",
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        mockIPPools.push(newPool);
        return newPool;
      }

      const res = await fetch(`${API_BASE}/wg/ip-pools`, {
        method: "POST",
        headers: getHeaders(),
        body: JSON.stringify(data),
      });
      return handleResponse(res);
  },

    updateIPPool: async (poolID: string, data: UpdateIPPoolRequest): Promise<IPPoolResponse> => {
          if (USE_MOCK) {
        await delay(400);
        const poolIndex = mockIPPools.findIndex((p) => p.id === poolID);
        if (poolIndex === -1) throw new Error("IP pool not found");

        const updatedPool: IPPoolResponse = {
          ...mockIPPools[poolIndex],
          ...(data.name !== undefined && { name: data.name }),
          ...(data.routes !== undefined && { routes: data.routes }),
          ...(data.dns !== undefined && { dns: data.dns }),
          ...(data.endpoint !== undefined && { endpoint: data.endpoint }),
          ...(data.description !== undefined && { description: data.description }),
          ...(data.status !== undefined && { status: data.status }),
          updated_at: new Date().toISOString(),
        };
        mockIPPools[poolIndex] = updatedPool;
        return updatedPool;
          }

      const res = await fetch(`${API_BASE}/wg/ip-pools/${poolID}`, {
        method: "PUT",
        headers: getHeaders(),
        body: JSON.stringify(data),
          });
          return handleResponse(res);
      },

    deleteIPPool: async (poolID: string): Promise<void> => {
          if (USE_MOCK) {
        await delay(300);
        const poolIndex = mockIPPools.findIndex((p) => p.id === poolID);
        if (poolIndex === -1) throw new Error("IP pool not found");
        mockIPPools.splice(poolIndex, 1);
        return;
          }

      const res = await fetch(`${API_BASE}/wg/ip-pools/${poolID}`, {
        method: "DELETE",
              headers: getHeaders(),
      });
      if (!res.ok) {
        const text = await res.text();
        try {
          const json = JSON.parse(text);
          throw new Error(json.message || json.error || res.statusText);
        } catch {
          throw new Error(text || res.statusText);
        }
      }
    },

    getAvailableIPs: async (poolID: string, limit?: number): Promise<AvailableIPsResponse> => {
      if (USE_MOCK) {
        await delay(300);
        const pool = mockIPPools.find((p) => p.id === poolID);
        if (!pool) throw new Error("IP pool not found");

        // Mock available IPs
        const ips: string[] = [];
        const maxLimit = limit || 10;
        for (let i = 0; i < maxLimit; i++) {
          ips.push(`10.0.0.${100 + i}`);
        }

        return {
          ip_pool_id: poolID,
          cidr: pool.cidr,
          ips,
          total: ips.length,
        };
      }

      const query = limit ? `?limit=${limit}` : "";
      const res = await fetch(`${API_BASE}/wg/ip-pools/${poolID}/available-ips${query}`, {
        headers: getHeaders(),
      });
      return handleResponse(res);
    },

    // ========================================================================
    // Server Configuration
    // ========================================================================
    getServerConfig: async (): Promise<GetServerConfigResponse> => {
      if (USE_MOCK) {
        await delay(300);
        return {
          address: "100.100.100.1/24",
          listen_port: 51820,
          private_key: "MOCK_PRIVATE_KEY",
          mtu: 1420,
          post_up: "iptables -A FORWARD -i wg0 -j ACCEPT; iptables -A FORWARD -o wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE",
          post_down: "iptables -D FORWARD -i wg0 -j ACCEPT; iptables -D FORWARD -o wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE",
          public_key: "MOCK_PUBLIC_KEY",
        };
      }

      const res = await fetch(`${API_BASE}/wg/server-config`, {
        headers: getHeaders(),
          });
          return handleResponse(res);
    },

    updateServerConfig: async (data: UpdateServerConfigRequest): Promise<void> => {
      if (USE_MOCK) {
        await delay(400);
        return;
      }

      const res = await fetch(`${API_BASE}/wg/server-config`, {
        method: "PUT",
        headers: getHeaders(),
        body: JSON.stringify(data),
      });
      if (!res.ok) {
        const text = await res.text();
        try {
          const json = JSON.parse(text);
          throw new Error(json.message || json.error || res.statusText);
        } catch {
          throw new Error(text || res.statusText);
        }
      }
    },
  },

  // ==========================================================================
  // Users
  // ==========================================================================
  users: {
    list: async (options?: UserListOptions): Promise<ListResponse<UserResponse>> => {
          if (USE_MOCK) {
              await delay(300);
        let filtered = mockUsers.map((u) => ({
          username: u.username,
          nickname: u.nickname || u.username,
          email: u.email || `${u.username}@example.com`,
        }));

        if (options?.username) {
          filtered = filtered.filter((u) => u.username.includes(options.username!));
        }
        if (options?.role) {
          const roleFilter = options.role;
          filtered = filtered.filter((u) => {
            const user = mockUsers.find((mu) => mu.username === u.username);
            return user?.role === roleFilter;
          });
        }

        const offset = options?.offset || 0;
        const limit = options?.limit || 20;
        const total = filtered.length;
        const items = filtered.slice(offset, offset + limit);

        return { total, items };
      }

      const query = buildQueryString(options || {});
      const res = await fetch(`${API_BASE}/users${query}`, {
        headers: getHeaders(),
      });
      return handleResponse(res);
    },

    get: async (username: string): Promise<UserResponse> => {
      if (USE_MOCK) {
        await delay(200);
        const user = mockUsers.find((u) => u.username === username);
        if (!user) throw new Error("User not found");
        return {
          username: user.username,
          nickname: user.nickname || user.username,
          email: user.email || `${user.username}@example.com`,
        };
      }

      const res = await fetch(`${API_BASE}/users/${username}`, {
        headers: getHeaders(),
      });
      return handleResponse(res);
    },

    create: async (data: CreateUserRequest): Promise<void> => {
      if (USE_MOCK) {
        await delay(500);
        const newUser: User = {
          id: String(mockUsers.length + 1),
          username: data.username,
          role: "user",
          nickname: data.nickname,
          email: data.email,
        };
        mockUsers.push(newUser);
        return;
      }

          const res = await fetch(`${API_BASE}/users`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });
      if (!res.ok) {
        const text = await res.text();
        try {
          const json = JSON.parse(text);
          throw new Error(json.error || json.message || res.statusText);
        } catch {
          throw new Error(text || res.statusText);
        }
      }
    },

    update: async (username: string, data: UpdateUserRequest): Promise<UserResponse> => {
      if (USE_MOCK) {
        await delay(300);
        const userIndex = mockUsers.findIndex((u) => u.username === username);
        if (userIndex === -1) throw new Error("User not found");

        const updatedUser = {
          ...mockUsers[userIndex],
          ...data,
        };
        mockUsers[userIndex] = updatedUser;
        return {
          username: updatedUser.username,
          nickname: updatedUser.nickname || updatedUser.username,
          email: updatedUser.email || `${updatedUser.username}@example.com`,
        };
      }

      const res = await fetch(`${API_BASE}/users/${username}`, {
        method: "PUT",
        headers: getHeaders(),
        body: JSON.stringify(data),
          });
          return handleResponse(res);
    },

    delete: async (username: string): Promise<void> => {
      if (USE_MOCK) {
        await delay(300);
        mockUsers = mockUsers.filter((u) => u.username !== username);
        return;
      }

      const res = await fetch(`${API_BASE}/users/${username}`, {
        method: "DELETE",
        headers: getHeaders(),
      });
      if (!res.ok) throw new Error("Failed to delete user");
    },

    changePassword: async (username: string, data: ChangePasswordRequest): Promise<void> => {
      if (USE_MOCK) {
        await delay(500);
        return;
      }

      const res = await fetch(`${API_BASE}/users/${username}/password`, {
        method: "POST",
        headers: getHeaders(),
        body: JSON.stringify(data),
      });
      if (!res.ok) {
        const text = await res.text();
        try {
          const json = JSON.parse(text);
          throw new Error(json.error || json.message || res.statusText);
        } catch {
          throw new Error(text || res.statusText);
        }
      }
    },
  },
};
