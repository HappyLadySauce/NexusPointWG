import request from '@/utils/request';

export interface WGPeer {
    id: string;
    user_id: string;
    username: string;
    device_name: string;
    client_public_key: string;
    client_ip: string;
    allowed_ips: string;
    dns: string;
    persistent_keepalive: number;
    status: 'active' | 'disabled';
}

export interface WGPeerListResponse {
    total: number;
    items: WGPeer[];
}

export interface CreateWGPeerRequest {
    username: string;
    deviceName: string;
    allowedIPs?: string;
    persistentKeepalive?: number;
    endpoint?: string;
    dns?: string;
    privateKey?: string;
}

export interface UpdateWGPeerRequest {
    allowedIPs?: string;
    persistentKeepalive?: number;
    dns?: string;
    status?: 'active' | 'disabled';
    privateKey?: string;
    endpoint?: string;
    deviceName?: string;
    clientIP?: string;
}

export interface UserUpdateConfigRequest {
    allowedIPs?: string;
    persistentKeepalive?: number;
    dns?: string;
    endpoint?: string;
}

export interface WGPeerListQuery {
    user_id?: string;
    device_name?: string;
    client_ip?: string;
    status?: string;
    offset?: number;
    limit?: number;
}

export const getPeers = (params: WGPeerListQuery) => {
    return request.get<any, WGPeerListResponse>('/wg/peers', { params });
};

export const createPeer = (data: CreateWGPeerRequest) => {
    return request.post<any, WGPeer>('/wg/peers', data);
};

export const updatePeer = (id: string, data: UpdateWGPeerRequest) => {
    return request.put<any, WGPeer>(`/wg/peers/${id}`, data);
};

export const deletePeer = (id: string) => {
    return request.delete<any, void>(`/wg/peers/${id}`);
};

export const getMyConfigs = () => {
    return request.get<any, WGPeerListResponse>('/wg/configs');
};

export const downloadConfig = (id: string) => {
    return request.get(`/wg/configs/${id}/download`, { responseType: 'blob' });
};

export const rotateConfig = (id: string) => {
    return request.post<any, void>(`/wg/configs/${id}/rotate`);
};

export const revokeConfig = (id: string) => {
    return request.post<any, void>(`/wg/configs/${id}/revoke`);
};

export const updateConfig = (id: string, data: UserUpdateConfigRequest) => {
    return request.put<any, void>(`/wg/configs/${id}`, data);
};

