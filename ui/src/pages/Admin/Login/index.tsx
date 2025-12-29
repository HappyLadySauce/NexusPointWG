import { authApi } from '@/api';
import { useAuthStore } from '@/store/auth';
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { Button, Card, Checkbox, Form, Input, Typography, message } from 'antd';
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

const { Title, Text } = Typography;

const AdminLogin: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();

    const onFinish = async (values: any) => {
        setLoading(true);
        try {
            const res = await authApi.login(values);
            const token = res.token;

            if (token) {
                // Use useAuthStore to update state (this also handles localStorage and token decoding)
                useAuthStore.getState().setToken(token);

                // Get role from store after token is set
                const userInfo = useAuthStore.getState().userInfo;
                const role = userInfo?.role || 'user';

                if (role === 'admin') {
                    message.success('管理员登录成功');
                    navigate('/admin/dashboard');
                } else {
                    message.warning('此账号不是管理员，请使用普通用户登录');
                    useAuthStore.getState().logout();
                }
            } else {
                message.error('登录失败：未获取到 Token');
            }
        } catch (error: any) {
            console.error('Login error:', error);
            // Error message is already shown by the request interceptor for auth pages
            // Additional error handling can be added here if needed
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            minHeight: '100vh',
            backgroundColor: '#f0f2f5',
            backgroundImage: 'url("https://gw.alipayobjects.com/zos/rmsportal/TVYTbAXWheQpRcWDaDMu.svg")',
            backgroundRepeat: 'no-repeat',
            backgroundPosition: 'center 110px',
            backgroundSize: '100%',
        }}>
            <Card
                style={{ width: 400, boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}
                bodyStyle={{ padding: '40px 32px' }}
            >
                <div style={{ textAlign: 'center', marginBottom: 32 }}>
                    <img src="/vite.svg" alt="logo" style={{ height: 44, marginBottom: 16 }} />
                    <Title level={3} style={{ margin: 0 }}>NexusPoint WG</Title>
                    <Text type="secondary">管理员登录</Text>
                </div>

                <Form
                    name="admin_login"
                    className="login-form"
                    initialValues={{ remember: true }}
                    onFinish={onFinish}
                    size="large"
                >
                    <Form.Item
                        name="username"
                        rules={[{ required: true, message: '请输入管理员用户名!' }]}
                    >
                        <Input prefix={<UserOutlined className="site-form-item-icon" />} placeholder="管理员用户名" />
                    </Form.Item>
                    <Form.Item
                        name="password"
                        rules={[{ required: true, message: '请输入密码!' }]}
                    >
                        <Input
                            prefix={<LockOutlined className="site-form-item-icon" />}
                            type="password"
                            placeholder="密码"
                        />
                    </Form.Item>
                    <Form.Item>
                        <Form.Item name="remember" valuePropName="checked" noStyle>
                            <Checkbox>记住我</Checkbox>
                        </Form.Item>
                    </Form.Item>

                    <Form.Item>
                        <Button type="primary" htmlType="submit" className="login-form-button" block loading={loading}>
                            管理员登录
                        </Button>
                        <div style={{ marginTop: 16, textAlign: 'center' }}>
                            <Text type="secondary">普通用户请前往 </Text>
                            <a href="/login">用户登录</a>
                        </div>
                    </Form.Item>
                </Form>
            </Card>
        </div>
    );
};

export default AdminLogin;

