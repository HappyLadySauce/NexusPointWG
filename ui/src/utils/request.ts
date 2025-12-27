import i18n from '@/i18n/config';
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
                        message.error(i18n.t('common.sessionExpired'));
                    } else {
                        // On auth pages, show the error message from server
                        const errorMsg = data?.message || i18n.t('auth.login.failed');
                        message.error(errorMsg);
                    }
                    break;
                case 403:
                    message.error(i18n.t('common.permissionDenied'));
                    break;
                case 404:
                    message.error(i18n.t('common.resourceNotFound'));
                    break;
                case 500:
                    message.error(i18n.t('common.serverError'));
                    break;
                default:
                    // Provide context-specific error messages based on the page
                    const isRegisterPage = window.location.pathname.includes('/register');
                    const defaultMessage = isRegisterPage
                        ? i18n.t('auth.register.failed')
                        : i18n.t('common.unknownError');
                    message.error(data?.message || defaultMessage);
            }
        } else if (error.request) {
            message.error(i18n.t('common.networkError'));
        } else {
            message.error(i18n.t('common.error'));
        }
        return Promise.reject(error);
    }
);

export default request;

