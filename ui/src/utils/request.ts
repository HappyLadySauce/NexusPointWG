import { message } from 'antd';
import axios from 'axios';

// Create axios instance
const request = axios.create({
    baseURL: '/api/v1', // API base URL, aligned with vite proxy
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json',
    },
});

// Request interceptor
request.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// Response interceptor
request.interceptors.response.use(
    (response) => {
        return response.data;
    },
    (error) => {
        if (error.response) {
            const { status, data } = error.response;

            switch (status) {
                case 401:
                    // Unauthorized - clear token and redirect to login
                    localStorage.removeItem('token');
                    // Only redirect if not already on login/register/console page to avoid loops
                    const isAuthPage = window.location.pathname.includes('/login') || 
                                      window.location.pathname.includes('/register') || 
                                      window.location.pathname.includes('/console');
                    if (!isAuthPage) {
                        window.location.href = '/login';
                        message.error('登录已过期，请重新登录');
                    } else {
                        // On auth pages, show the error message from server
                        const errorMsg = data?.message || '用户名或密码错误';
                        message.error(errorMsg);
                    }
                    break;
                case 403:
                    message.error('没有权限执行此操作');
                    break;
                case 404:
                    message.error('请求的资源不存在');
                    break;
                case 500:
                    message.error('服务器错误，请稍后重试');
                    break;
                default:
                    message.error(data?.message || '发生未知错误');
            }
        } else if (error.request) {
            message.error('网络连接失败，请检查您的网络');
        } else {
            message.error('请求配置错误');
        }
        return Promise.reject(error);
    }
);

export default request;

