import i18n from '@/i18n/config';
import { useAuthStore } from '@/store/auth';
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
        const token = useAuthStore.getState().token;
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
    110003: 'error.userNotFound',

    // WireGuard codes
    120001: 'error.wgPeerNotFound',
    // Server config errors
    120002: 'error.wgServerConfigNotFound',
    120003: 'error.wgWriteFailed',
    120004: 'error.wgApplyFailed'
};

// Response interceptor
request.interceptors.response.use(
    (response) => {
        // Return the whole response object if data structure is not standard
        // But api-contract says { code, message, details, ... } or data.
        // Usually axios returns { data, status, headers ... }
        // We return response.data to get the payload directly.
        return response.data;
    },
    (error) => {
        if (error.response) {
            const { status, data } = error.response;

            // Try to map error code to translation key
            const errorCode = data?.code;
            let errorMessage = data?.message;

            // Priority:
            // 1. code -> i18n
            // 2. details -> i18n
            // 3. message (fallback)

            if (errorCode && errorCodeMap[errorCode]) {
                errorMessage = i18n.t(errorCodeMap[errorCode] as any);
            }

            // Handle validation details if present
            // e.g. details: { email: "validation.email|..." }
            if (data?.details && typeof data.details === 'object') {
                // If it's a 400 validation error, we might want to let the component handle it
                // But if we want global toast:
                // Construct a message or just show the first one.
                // For now, let's stick to the errorCode message or fallback.
                // Component form will handle field highlighting via catch().
            }

            switch (status) {
                case 401:
                    // Unauthorized - clear token and redirect to login
                    useAuthStore.getState().logout();
                    // Only redirect if not already on login/register/console page to avoid loops
                    const isAuthPage = window.location.pathname.includes('/login') ||
                        window.location.pathname.includes('/register');
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
                case 400:
                    // For registration page, let the component handle 400 errors (field conflicts, validation, etc.)
                    const isRegisterPage = window.location.pathname.includes('/register');
                    if (isRegisterPage) {
                        // Don't show error message here, let Register component handle it
                        break;
                    }
                    message.error(errorMessage || i18n.t('common.badRequest'));
                    break;
                default:
                    message.error(errorMessage || i18n.t('common.unknownError'));
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
