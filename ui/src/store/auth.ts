import { jwtDecode } from 'jwt-decode';
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface UserInfo {
    username: string;
    role: 'user' | 'admin';
    id?: string;
    nickname?: string;
    avatar?: string;
    exp?: number;
}

interface AuthState {
    token: string | null;
    userInfo: UserInfo | null;
    setToken: (token: string) => void;
    logout: () => void;
    isAuthenticated: () => boolean;
    isAdmin: () => boolean;
}

export const useAuthStore = create<AuthState>()(
    persist(
        (set, get) => ({
            token: null,
            userInfo: null,
            setToken: (token: string) => {
                try {
                    const decoded = jwtDecode<UserInfo>(token);
                    // Extract custom claims if needed, usually 'role' and 'username' are standard or custom claims in your JWT
                    // Ensure your backend JWT payload matches this structure.
                    // Assuming backend JWT payload: { "username": "...", "role": "...", "id": "...", "exp": ... }
                    set({ token, userInfo: decoded });
                    localStorage.setItem('token', token);
                } catch (error) {
                    console.error('Invalid token:', error);
                    set({ token: null, userInfo: null });
                    localStorage.removeItem('token');
                }
            },
            logout: () => {
                set({ token: null, userInfo: null });
                localStorage.removeItem('token');
            },
            isAuthenticated: () => {
                const { token, userInfo } = get();
                if (!token || !userInfo) return false;
                // Check expiration
                const now = Date.now() / 1000;
                if (userInfo.exp && userInfo.exp < now) {
                    get().logout();
                    return false;
                }
                return true;
            },
            isAdmin: () => {
                const { userInfo } = get();
                return userInfo?.role === 'admin';
            },
        }),
        {
            name: 'auth-storage', // name of the item in the storage (must be unique)
            partialize: (state) => ({ token: state.token, userInfo: state.userInfo }), // persist only token and userInfo
        }
    )
);

