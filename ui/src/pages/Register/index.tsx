import { userApi } from '@/api';
import { LockOutlined, MailOutlined, UserOutlined } from '@ant-design/icons';
import { Button, Card, Form, Input, Typography, message } from 'antd';
import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

const { Title, Text } = Typography;

const Register: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();

    const onFinish = async (values: any) => {
        setLoading(true);
        try {
            await userApi.createUser({
                username: values.username,
                nickname: values.nickname || values.username,
                email: values.email,
                password: values.password,
            });
            message.success('注册成功！请等待管理员审核激活账号');
            navigate('/login');
        } catch (error: any) {
            console.error('Register error:', error);
            
            // Handle validation errors with field-level messages
            if (error?.response?.data?.details) {
                const details = error.response.data.details;
                // Show field-specific errors
                Object.keys(details).forEach((field) => {
                    const fieldName = field === 'username' ? '用户名' : 
                                    field === 'email' ? '邮箱' : 
                                    field === 'password' ? '密码' : field;
                    message.error(`${fieldName}: ${details[field]}`);
                });
            } else {
                // For general errors, the interceptor will show the error message
                // If server didn't provide a message, interceptor shows generic "发生未知错误"
                // We can add registration-specific handling here if needed, but avoid duplicate messages
                // The interceptor's default case handles the fallback, so we don't need to show again
            }
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
                style={{ width: 450, boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}
                bodyStyle={{ padding: '40px 32px' }}
            >
                <div style={{ textAlign: 'center', marginBottom: 32 }}>
                    <img src="/vite.svg" alt="logo" style={{ height: 44, marginBottom: 16 }} />
                    <Title level={3} style={{ margin: 0 }}>NexusPoint WG</Title>
                    <Text type="secondary">用户注册</Text>
                </div>

                <Form
                    name="register"
                    className="register-form"
                    onFinish={onFinish}
                    size="large"
                    autoComplete="off"
                >
                    <Form.Item
                        name="username"
                        rules={[
                            { required: true, message: '请输入用户名!' },
                            { min: 3, message: '用户名至少3个字符!' },
                            { max: 32, message: '用户名最多32个字符!' },
                        ]}
                    >
                        <Input prefix={<UserOutlined />} placeholder="用户名 (3-32字符)" />
                    </Form.Item>

                    <Form.Item
                        name="nickname"
                        rules={[
                            { max: 32, message: '昵称最多32个字符!' },
                        ]}
                    >
                        <Input prefix={<UserOutlined />} placeholder="昵称 (可选，默认使用用户名)" />
                    </Form.Item>

                    <Form.Item
                        name="email"
                        rules={[
                            { required: true, message: '请输入邮箱!' },
                            { type: 'email', message: '请输入有效的邮箱地址!' },
                        ]}
                    >
                        <Input prefix={<MailOutlined />} placeholder="邮箱地址" />
                    </Form.Item>

                    <Form.Item
                        name="password"
                        rules={[
                            { required: true, message: '请输入密码!' },
                            { min: 8, message: '密码至少8个字符!' },
                            { max: 32, message: '密码最多32个字符!' },
                        ]}
                    >
                        <Input.Password
                            prefix={<LockOutlined />}
                            placeholder="密码 (8-32字符)"
                        />
                    </Form.Item>

                    <Form.Item
                        name="confirmPassword"
                        dependencies={['password']}
                        rules={[
                            { required: true, message: '请确认密码!' },
                            ({ getFieldValue }) => ({
                                validator(_, value) {
                                    if (!value || getFieldValue('password') === value) {
                                        return Promise.resolve();
                                    }
                                    return Promise.reject(new Error('两次输入的密码不一致!'));
                                },
                            }),
                        ]}
                    >
                        <Input.Password
                            prefix={<LockOutlined />}
                            placeholder="确认密码"
                        />
                    </Form.Item>

                    <Form.Item>
                        <Button type="primary" htmlType="submit" className="register-form-button" block loading={loading}>
                            注册
                        </Button>
                        <div style={{ marginTop: 16, textAlign: 'center' }}>
                            <Text>已有账号? </Text>
                            <Link to="/login">立即登录</Link>
                        </div>
                    </Form.Item>
                </Form>
            </Card>
        </div>
    );
};

export default Register;

