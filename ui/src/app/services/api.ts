import { toast } from "sonner";

// Toggle this to switch between real backend and mock mode
// Set to false to connect to the real Go backend
const USE_MOCK = false;
const API_BASE = "/api/v1";

export interface User {
  id: string;
  username: string;
  role: "admin" | "user";
}

export interface Peer {
  id: string;
  name: string;
  public_key: string;
  endpoint: string;
  allowed_ips: string[];
  latest_handshake: string;
  transfer_rx: number;
  transfer_tx: number;
  status: "active" | "inactive";
  created_at: string;
  updated_at: string;
}

export interface IPPool {
  id: string;
  cidr: string;
  total_ips: number;
  available_ips: number;
  used_ips: number;
}

export interface GlobalSettings {
  endpoint_address: string;
  dns_servers: string[];
  mtu: number;
  keepalive: number;
  firewall_mark: number;
}

// Mock Data
let mockUsers: User[] = [
  { id: "1", username: "admin", role: "admin" },
  { id: "2", username: "user", role: "user" },
];

let mockPeers: Peer[] = [
  {
    id: "peer-1",
    name: "iPhone",
    public_key: "Jw8.../2A=",
    endpoint: "192.168.1.5:51820",
    allowed_ips: ["10.0.0.2/32"],
    latest_handshake: newqp(),
    transfer_rx: 1024 * 1024 * 50, // 50MB
    transfer_tx: 1024 * 1024 * 12, // 12MB
    status: "active",
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "peer-2",
    name: "MacBook Pro",
    public_key: "K2d.../9C=",
    endpoint: "192.168.1.6:51820",
    allowed_ips: ["10.0.0.3/32"],
    latest_handshake: "2023-10-25T10:00:00Z",
    transfer_rx: 1024 * 1024 * 1500, // 1.5GB
    transfer_tx: 1024 * 1024 * 200, // 200MB
    status: "inactive",
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
];

let mockIPPools: IPPool[] = [
  {
    id: "pool-1",
    cidr: "10.0.0.0/24",
    total_ips: 254,
    available_ips: 252,
    used_ips: 2,
  },
];

let mockSettings: GlobalSettings = {
    endpoint_address: "vpn.nexuspoint.com",
    dns_servers: ["1.1.1.1", "8.8.8.8"],
    mtu: 1420,
    keepalive: 25,
    firewall_mark: 0
};

function newqp() {
    return new Date().toISOString();
}

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

export const api = {
  auth: {
    login: async (username: string, password: string): Promise<{ token: string; user: User }> => {
      if (USE_MOCK) {
        await delay(500);
        if (username === "admin" && password === "admin") {
          return { token: "mock-jwt-token-admin", user: { id: "1", username: "admin", role: "admin" } };
        }
        if (username === "user" && password === "user") {
          return { token: "mock-jwt-token-user", user: { id: "2", username: "user", role: "user" } };
        }
        throw new Error("Invalid credentials");
      }
      const res = await fetch(`${API_BASE}/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
      });
      const data = await handleResponse(res);
      // Ensure token is saved
      if (data.token) localStorage.setItem("token", data.token);
      return data;
    },
    register: async (username: string, password: string) => {
      if (USE_MOCK) {
          await delay(500);
          return { success: true };
      }
       const res = await fetch(`${API_BASE}/auth/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
      });
      return handleResponse(res);
    },
  },
  wg: {
    listPeers: async (): Promise<Peer[]> => {
      if (USE_MOCK) {
        await delay(300);
        return [...mockPeers];
      }
      const res = await fetch(`${API_BASE}/wg/peers`, {
          headers: getHeaders()
      });
      return handleResponse(res);
    },
    createPeer: async (data: Partial<Peer>): Promise<Peer> => {
      if (USE_MOCK) {
        await delay(400);
        const newPeer: Peer = {
          id: `peer-${Date.now()}`,
          name: data.name || "Unnamed Peer",
          public_key: "MOCK_KEY_" + Math.random().toString(36).substring(7),
          endpoint: "(none)",
          allowed_ips: ["10.0.0." + (mockPeers.length + 2) + "/32"],
          latest_handshake: "",
          transfer_rx: 0,
          transfer_tx: 0,
          status: "active",
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
    deletePeer: async (id: string): Promise<void> => {
      if (USE_MOCK) {
        await delay(300);
        mockPeers = mockPeers.filter((p) => p.id !== id);
        return;
      }
       const res = await fetch(`${API_BASE}/wg/peers/${id}`, { 
           method: "DELETE",
           headers: getHeaders()
       });
       if (!res.ok) throw new Error("Failed to delete peer");
    },
    downloadConfig: async (id: string): Promise<string> => {
        if (USE_MOCK) {
            await delay(500);
            return `[Interface]\nPrivateKey = MOCK_PRIVATE_KEY\nAddress = 10.0.0.5/32\nDNS = 1.1.1.1\n\n[Peer]\nPublicKey = SERVER_PUBLIC_KEY\nEndpoint = vpn.example.com:51820\nAllowedIPs = 0.0.0.0/0`;
        }
        const res = await fetch(`${API_BASE}/wg/peers/${id}/config`, {
            headers: getHeaders()
        });
        if (!res.ok) throw new Error("Failed to download config");
        return res.text();
    },
    listIPPools: async (): Promise<IPPool[]> => {
       if (USE_MOCK) {
        await delay(300);
        return [...mockIPPools];
      }
      const res = await fetch(`${API_BASE}/wg/ip-pools`, {
          headers: getHeaders()
      });
      return handleResponse(res);
    }
  },
  settings: {
      get: async (): Promise<GlobalSettings> => {
          if (USE_MOCK) {
              await delay(300);
              return { ...mockSettings };
          }
          const res = await fetch(`${API_BASE}/settings`, {
              headers: getHeaders()
          });
          return handleResponse(res);
      },
      update: async (settings: Partial<GlobalSettings>): Promise<GlobalSettings> => {
          if (USE_MOCK) {
              await delay(500);
              mockSettings = { ...mockSettings, ...settings };
              return { ...mockSettings };
          }
          const res = await fetch(`${API_BASE}/settings`, {
              method: "PUT",
              headers: getHeaders(),
              body: JSON.stringify(settings)
          });
          return handleResponse(res);
      }
  },
  users: {
      list: async (): Promise<User[]> => {
          if (USE_MOCK) {
              await delay(300);
              return [...mockUsers];
          }
          const res = await fetch(`${API_BASE}/users`, {
              headers: getHeaders()
          });
          return handleResponse(res);
      }
  }
};
