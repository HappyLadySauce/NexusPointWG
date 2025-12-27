import { Navigate, Route, Routes } from 'react-router-dom'
import MainLayout from '../components/Layout/MainLayout'
import AdminDashboard from '../pages/Admin/Dashboard'
import AdminLogin from '../pages/Admin/Login'
import Login from '../pages/Login'
import Register from '../pages/Register'
import UserDashboard from '../pages/User/Dashboard'

const AppRouter = () => {
    return (
        <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/console" element={<AdminLogin />} />

            {/* Admin Routes */}
            <Route path="/admin" element={<MainLayout />}>
                <Route path="dashboard" element={<AdminDashboard />} />
                {/* Placeholders for now */}
                <Route path="users" element={<div>用户管理</div>} />
                <Route path="peers" element={<div>Peer 管理</div>} />
                <Route path="settings" element={<div>系统设置</div>} />
                <Route index element={<Navigate to="dashboard" replace />} />
            </Route>

            {/* User Routes */}
            <Route path="/user" element={<MainLayout />}>
                <Route path="dashboard" element={<UserDashboard />} />
                <Route path="profile" element={<div>个人中心</div>} />
                <Route index element={<Navigate to="dashboard" replace />} />
            </Route>

            {/* Default redirect */}
            <Route path="/" element={<Navigate to="/login" replace />} />
            <Route path="/admin/login" element={<Navigate to="/console" replace />} />
            <Route path="*" element={<div>404 Not Found</div>} />
        </Routes>
    )
}

export default AppRouter
