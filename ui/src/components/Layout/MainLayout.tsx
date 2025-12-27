import {
    CloudServerOutlined,
    DashboardOutlined,
    DownloadOutlined,
    LogoutOutlined,
    MenuFoldOutlined,
    MenuUnfoldOutlined,
    ProfileOutlined,
    SettingOutlined,
    TeamOutlined,
    UserOutlined
} from '@ant-design/icons';
import { Avatar, Button, Dropdown, Layout, Menu, theme } from 'antd';
import React, { useState } from 'react';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';

const { Header, Sider, Content } = Layout;

const MainLayout: React.FC = () => {
    const [collapsed, setCollapsed] = useState(false);
    const {
        token: { colorBgContainer, borderRadiusLG },
    } = theme.useToken();

    const navigate = useNavigate();
    const location = useLocation();

    // Mock role check - replace with actual auth context later
    const isAdmin = location.pathname.startsWith('/admin');

    const adminMenuItems = [
        {
            key: '/admin/dashboard',
            icon: <DashboardOutlined />,
            label: '仪表盘',
        },
        {
            key: '/admin/users',
            icon: <TeamOutlined />,
            label: '用户管理',
        },
        {
            key: '/admin/peers',
            icon: <CloudServerOutlined />,
            label: 'Peer 管理',
        },
        {
            key: '/admin/settings',
            icon: <SettingOutlined />,
            label: '系统设置',
        },
    ];

    const userMenuItems = [
        {
            key: '/user/dashboard',
            icon: <DownloadOutlined />,
            label: '我的配置',
        },
        {
            key: '/user/profile',
            icon: <ProfileOutlined />,
            label: '个人中心',
        },
    ];

    const menuItems = isAdmin ? adminMenuItems : userMenuItems;

    const userMenu = (
        <Menu>
            <Menu.Item key="profile" icon={<UserOutlined />} onClick={() => navigate('/user/profile')}>
                个人中心
            </Menu.Item>
            <Menu.Divider />
            <Menu.Item key="logout" icon={<LogoutOutlined />} danger onClick={() => navigate('/login')}>
                退出登录
            </Menu.Item>
        </Menu>
    );

    return (
        <Layout style={{ minHeight: '100vh' }}>
            <Sider trigger={null} collapsible collapsed={collapsed} breakpoint="lg" onBreakpoint={(broken) => {
                if (broken) setCollapsed(true);
            }}>
                <div className="demo-logo-vertical" style={{ height: 32, margin: 16, background: 'rgba(255, 255, 255, 0.2)', borderRadius: 6 }} />
                <Menu
                    theme="dark"
                    mode="inline"
                    defaultSelectedKeys={[location.pathname]}
                    items={menuItems}
                    onClick={({ key }) => navigate(key)}
                />
            </Sider>
            <Layout>
                <Header style={{ padding: 0, background: colorBgContainer, display: 'flex', justifyContent: 'space-between', alignItems: 'center', paddingRight: 24 }}>
                    <Button
                        type="text"
                        icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                        onClick={() => setCollapsed(!collapsed)}
                        style={{
                            fontSize: '16px',
                            width: 64,
                            height: 64,
                        }}
                    />

                    <Dropdown overlay={userMenu} placement="bottomRight">
                        <span style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
                            <Avatar icon={<UserOutlined />} />
                            <span>{isAdmin ? 'Admin' : 'User'}</span>
                        </span>
                    </Dropdown>
                </Header>
                <Content
                    style={{
                        margin: '24px 16px',
                        padding: 24,
                        minHeight: 280,
                        background: colorBgContainer,
                        borderRadius: borderRadiusLG,
                    }}
                >
                    <Outlet />
                </Content>
            </Layout>
        </Layout>
    );
};

export default MainLayout;

