import request from '@/utils/request';

export interface User {
    username: string;
    nickname: string;
    email: string;
    avatar?: string;
    role?: 'user' | 'admin';
    status?: 'active' | 'inactive' | 'deleted';
}

export interface UserListResponse {
    total: number;
    items: User[];
}

export interface UserListQuery {
    username?: string;
    email?: string;
    role?: string;
    status?: string;
    offset?: number;
    limit?: number;
}

export const getUser = (username: string) => {
    return request.get<any, User>(`/users/${username}`);
};

export const listUsers = (params?: UserListQuery) => {
    return request.get<any, UserListResponse>('/users', { params });
};

