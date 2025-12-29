import request from '@/utils/request';

export interface WGPeer {
    id: string;
    user_id: string;
    username: string;
    device_name: string;
    client_public_key: string;
    client_ip: string;
    allowed_ips: string;
    persistent_keepalive: number;
    status: 'active' | 'revoked';
}

export interface WGPeerListResponse {
    total: number;
    items: WGPeer[];
}

export interface CreateWGPeerRequest {
    username: string;
    device_name: string;
    allowed_ips?: string;
    persistent_keepalive?: number;
}

export interface UpdateWGPeerRequest {
    allowed_ips?: string;
    persistent_keepalive?: number;
    status?: 'active' | 'revoked';
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

