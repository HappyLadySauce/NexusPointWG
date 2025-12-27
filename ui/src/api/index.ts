import type { LoginResponse, SystemStats, User } from '@/types/api';
import request from '@/utils/request';

export const authApi = {
    login: (data: any) => {
        return request.post<any, LoginResponse>('/login', data);
    },
    logout: () => {
        // Client-side logout mainly, but can call backend if needed
        localStorage.removeItem('token');
    },
    getCurrentUser: () => {
        // Assuming there's an endpoint to get current user info
        // Adjust endpoint based on actual backend
        return request.get<any, User>('/users/me');
    }
};

export const userApi = {
    getUsers: (params?: any) => {
        return request.get<any, User[]>('/users', { params });
    },
    createUser: (data: any) => {
        return request.post<any, any>('/users', data);
    },
    updateUser: (username: string, data: any) => {
        return request.put<any, any>(`/users/${username}`, data);
    },
    deleteUser: (username: string) => {
        return request.delete<any, any>(`/users/${username}`);
    },
    changePassword: (username: string, data: any) => {
        return request.post<any, any>(`/users/${username}/password`, data);
    }
};

export const peerApi = {
    getPeers: (params?: any) => {
        return request.get<any, any>('/wg/peers', { params });
    },
    createPeer: (data: any) => {
        return request.post<any, any>('/wg/peers', data);
    },
    getPeer: (id: string) => {
        return request.get<any, any>(`/wg/peers/${id}`);
    },
    deletePeer: (id: string) => {
        return request.delete<any, any>(`/wg/peers/${id}`);
    },
    getUserConfigs: () => {
        return request.get<any, any>('/wg/configs');
    },
    downloadConfig: (id: string) => {
        return request.get(`/wg/configs/${id}/download`, {
            responseType: 'blob' // Important for file download
        });
    }
};

export const adminApi = {
    getStats: () => {
        // Mocking endpoint for now, or assume it exists
        return request.get<any, SystemStats>('/admin/stats');
    }
};
