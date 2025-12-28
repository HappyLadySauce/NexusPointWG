import LanguageSelect from '@/components/LanguageSelect';
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
import { useTranslation } from 'react-i18next';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';

const { Header, Sider, Content } = Layout;

const MainLayout: React.FC = () => {
    const { t } = useTranslation();
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
            label: t('menu.dashboard'),
        },
        {
            key: '/admin/users',
            icon: <TeamOutlined />,
            label: t('menu.users'),
        },
        {
            key: '/admin/peers',
            icon: <CloudServerOutlined />,
            label: t('menu.peers'),
        },
        {
            key: '/admin/settings',
            icon: <SettingOutlined />,
            label: t('menu.settings'),
        },
    ];

    const userMenuItems = [
        {
            key: '/user/dashboard',
            icon: <DownloadOutlined />,
            label: t('menu.myConfig'),
        },
        {
            key: '/user/profile',
            icon: <ProfileOutlined />,
            label: t('menu.profile'),
        },
    ];

    const menuItems = isAdmin ? adminMenuItems : userMenuItems;

    const userMenuItemsList = [
        {
            key: 'profile',
            icon: <UserOutlined />,
            label: t('menu.profile'),
            onClick: () => navigate('/user/profile')
        },
        {
            type: 'divider'
        },
        {
            key: 'logout',
            icon: <LogoutOutlined />,
            label: t('menu.logout'),
            danger: true,
            onClick: () => navigate('/login')
        }
    ];

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

                    <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                        <LanguageSelect />
                        <Dropdown menu={{ items: userMenuItemsList as any }} placement="bottomRight">
                            <span style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
                                <Avatar icon={<UserOutlined />} />
                                <span>{isAdmin ? 'Admin' : 'User'}</span>
                            </span>
                        </Dropdown>
                    </div>
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

