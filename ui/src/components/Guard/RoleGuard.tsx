import { useAuthStore } from '@/store/auth';
import { Navigate, useLocation } from 'react-router-dom';

interface RoleGuardProps {
    children: React.ReactNode;
    roles: ('admin' | 'user')[];
}

const RoleGuard: React.FC<RoleGuardProps> = ({ children, roles }) => {
    const userInfo = useAuthStore((state) => state.userInfo);
    const location = useLocation();

    if (!userInfo || !roles.includes(userInfo.role)) {
        // Redirect to appropriate dashboard based on role, or login if no role
        if (userInfo?.role === 'admin') {
            return <Navigate to="/admin/dashboard" replace />;
        } else if (userInfo?.role === 'user') {
            return <Navigate to="/user/dashboard" replace />;
        } else {
            return <Navigate to="/login" state={{ from: location }} replace />;
        }
    }

    return <>{children}</>;
};

export default RoleGuard;

