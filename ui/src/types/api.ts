// Common response types
export interface ApiResponse<T = any> {
    code?: number;
    message?: string;
    data?: T;
    // Some APIs might return data directly or wrapped in data property
    // Adjust based on actual backend response structure
}

// User types
export interface User {
    id?: string;
    username: string;
    nickname: string;
    email: string;
    avatar?: string;
    role: 'admin' | 'user';
    status: 'active' | 'inactive' | 'deleted';
    created_at?: string;
    last_login?: string;
}

export interface LoginResponse {
    token: string;
    // user: User; // Backend might return user info with token
}

export interface Peer {
    id: string;
    name: string;
    user_id: string;
    username?: string; // For display
    public_key: string;
    allowed_ips: string;
    endpoint: string;
    last_handshake?: string;
    rx_bytes?: number;
    tx_bytes?: number;
    status: 'active' | 'inactive';
    created_at: string;
}

export interface SystemStats {
    total_users: number;
    active_peers: number;
    total_peers: number;
    total_traffic_rx: number;
    total_traffic_tx: number;
    cpu_usage: number; // percentage
    memory_usage: number; // percentage
}
