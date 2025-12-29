import { authApi } from '@/api';
import LanguageSelect from '@/components/LanguageSelect';
import { useAuthStore } from '@/store/auth';
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { Button, Card, Checkbox, Form, Input, Typography, message } from 'antd';
import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate } from 'react-router-dom';

const { Title, Text } = Typography;

const Login: React.FC = () => {
    const { t } = useTranslation();
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
                // Use useAuthStore to update state (this also handles localStorage and token decoding)
                useAuthStore.getState().setToken(token);
                message.success(t('auth.login.success'));

                // Get role from store after token is set
                const userInfo = useAuthStore.getState().userInfo;
                const role = userInfo?.role || 'user';

                if (role === 'admin') {
                    navigate('/admin/dashboard');
                } else {
                    navigate('/user/dashboard');
                }
            } else {
                message.error(t('auth.login.failed') + '：' + t('common.unknownError'));
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
            backgroundImage: 'url("https://gw.alipayobjects.com/zos/rmsportal/TVYTbAXWheQpRcWDaDMu.svg")', // Ant Design Pro background pattern
            backgroundRepeat: 'no-repeat',
            backgroundPosition: 'center 110px',
            backgroundSize: '100%',
        }}>
            <div style={{ position: 'absolute', top: 20, right: 20 }}>
                <LanguageSelect />
            </div>
            <Card
                style={{ width: 400, boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}
                bodyStyle={{ padding: '40px 32px' }}
            >
                <div style={{ textAlign: 'center', marginBottom: 32 }}>
                    <img src="/vite.svg" alt="logo" style={{ height: 44, marginBottom: 16 }} />
                    <Title level={3} style={{ margin: 0 }}>{t('auth.login.title')}</Title>
                    <Text type="secondary">{t('auth.login.subtitle')}</Text>
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
                        rules={[{ required: true, message: t('auth.fields.username.required') }]}
                    >
                        <Input prefix={<UserOutlined className="site-form-item-icon" />} placeholder={t('auth.fields.username.placeholder')} />
                    </Form.Item>
                    <Form.Item
                        name="password"
                        rules={[{ required: true, message: t('auth.fields.password.required') }]}
                    >
                        <Input
                            prefix={<LockOutlined className="site-form-item-icon" />}
                            type="password"
                            placeholder={t('auth.fields.password.placeholder')}
                        />
                    </Form.Item>
                    <Form.Item>
                        <Form.Item name="remember" valuePropName="checked" noStyle>
                            <Checkbox>{t('auth.login.rememberMe')}</Checkbox>
                        </Form.Item>

                        <a className="login-form-forgot" href="" style={{ float: 'right' }}>
                            {t('auth.login.forgotPassword')}
                        </a>
                    </Form.Item>

                    <Form.Item>
                        <Button type="primary" htmlType="submit" className="login-form-button" block loading={loading}>
                            {t('auth.login.submit')}
                        </Button>
                        <div style={{ marginTop: 16, textAlign: 'center' }}>
                            {t('auth.login.noAccount')} <Link to="/register">{t('auth.login.registerNow')}</Link>
                        </div>
                    </Form.Item>
                </Form>
            </Card>
        </div>
    );
};

export default Login;
