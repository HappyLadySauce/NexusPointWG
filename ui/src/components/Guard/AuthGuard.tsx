import { useAuthStore } from '@/store/auth';
import { Navigate, useLocation } from 'react-router-dom';

interface AuthGuardProps {
    children: React.ReactNode;
}

const AuthGuard: React.FC<AuthGuardProps> = ({ children }) => {
    const isAuthenticated = useAuthStore((state) => state.isAuthenticated());
    const location = useLocation();

    if (!isAuthenticated) {
        return <Navigate to="/login" state={{ from: location }} replace />;
    }

    return <>{children}</>;
};

export default AuthGuard;

