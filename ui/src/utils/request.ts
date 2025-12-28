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

// Error code to translation key mapping
const errorCodeMap: Record<number, string> = {
    100201: 'error.encrypt',
    // 100202: ErrSignatureInvalid
    100202: 'error.tokenInvalid',
    // 100203: ErrExpired
    100203: 'error.tokenExpired',
    // 100204/100205: invalid/missing auth header
    100204: 'error.tokenInvalid',
    100205: 'error.tokenInvalid',
    // 100206: ErrPasswordIncorrect -> security policy: show generic auth failed
    100206: 'error.authFailed',
    // 100207: ErrPermissionDenied
    100207: 'error.permissionDenied',

    // Server codes
    110004: 'error.userNotActive',
    110001: 'error.userAlreadyExist',
    110002: 'error.emailAlreadyExist',
    110003: 'error.userNotFound'
};

// Response interceptor
request.interceptors.response.use(
    (response) => {
        return response.data;
    },
    (error) => {
        if (error.response) {
            const { status, data } = error.response;

            // Try to map error code to translation key
            const errorCode = data?.code;
            let errorMessage = data?.message;

            if (errorCode && errorCodeMap[errorCode]) {
                errorMessage = i18n.t(errorCodeMap[errorCode] as any);
            }

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
                        message.error(i18n.t('error.tokenExpired'));
                    } else {
                        // On auth pages, show the error message from server or translation
                        message.error(errorMessage || i18n.t('auth.login.failed'));
                    }
                    break;
                case 403:
                    message.error(errorMessage || i18n.t('error.permissionDenied'));
                    break;
                case 404:
                    message.error(errorMessage || i18n.t('common.resourceNotFound'));
                    break;
                case 500:
                    message.error(errorMessage || i18n.t('common.serverError'));
                    break;
                default:
                    // For registration page, let the component handle 400 errors (field conflicts, validation, etc.)
                    const isRegisterPage = window.location.pathname.includes('/register');
                    if (isRegisterPage && status === 400) {
                        // Don't show error message here, let Register component handle it
                        break;
                    }
                    // Provide context-specific error messages based on the page
                    const defaultMessage = isRegisterPage
                        ? i18n.t('auth.register.failed')
                        : i18n.t('common.unknownError');
                    message.error(errorMessage || defaultMessage);
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

