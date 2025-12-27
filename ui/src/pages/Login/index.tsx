import { authApi } from '@/api';
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { Button, Card, Checkbox, Form, Input, Typography, message } from 'antd';
import { jwtDecode } from "jwt-decode";
import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

const { Title, Text } = Typography;

const Login: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();

    const onFinish = async (values: any) => {
        setLoading(true);
        try {
            const res = await authApi.login(values);
            // Assuming res contains token directly or in data structure based on request.ts interceptor
            // The interceptor returns response.data
            const token = res.token;

            if (token) {
                localStorage.setItem('token', token);
                message.success('登录成功');

                // Decode token to get role
                try {
                    const decoded: any = jwtDecode(token);
                    // Assuming token payload has 'role' field. Adjust if needed.
                    const role = decoded.role || 'user';

                    if (role === 'admin') {
                        navigate('/admin/dashboard');
                    } else {
                        navigate('/user/dashboard');
                    }
                } catch (e) {
                    // Fallback if decode fails or simple logic
                    // For now, let's just default to user dashboard if decode fails, or maybe admin for testing if username is admin
                    if (values.username === 'admin') {
                        navigate('/admin/dashboard');
                    } else {
                        navigate('/user/dashboard');
                    }
                }
            } else {
                message.error('登录失败：未获取到 Token');
            }
        } catch (error) {
            console.error('Login error:', error);
            // Error message is handled by interceptor globally, but we can add specific handling here if needed
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
            backgroundImage: 'url("https://gw.alipayobjects.com/zos/rmsportal/TVYTbAXWheQpRcWDaDMu.svg")', // Ant Design Pro background pattern
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
                    <Text type="secondary">WireGuard 管理平台</Text>
                </div>

                <Form
                    name="normal_login"
                    className="login-form"
                    initialValues={{ remember: true }}
                    onFinish={onFinish}
                    size="large"
                >
                    <Form.Item
                        name="username"
                        rules={[{ required: true, message: '请输入用户名!' }]}
                    >
                        <Input prefix={<UserOutlined className="site-form-item-icon" />} placeholder="用户名 / 邮箱" />
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

                        <a className="login-form-forgot" href="" style={{ float: 'right' }}>
                            忘记密码
                        </a>
                    </Form.Item>

                    <Form.Item>
                        <Button type="primary" htmlType="submit" className="login-form-button" block loading={loading}>
                            登录
                        </Button>
                        <div style={{ marginTop: 16, textAlign: 'center' }}>
                            还没有账号? <Link to="/register">立即注册</Link>
                        </div>
                    </Form.Item>
                </Form>
            </Card>
        </div>
    );
};

export default Login;
