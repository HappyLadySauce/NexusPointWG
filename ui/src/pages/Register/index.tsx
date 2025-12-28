import { userApi } from '@/api';
import LanguageSelect from '@/components/LanguageSelect';
import { LockOutlined, MailOutlined, UserOutlined } from '@ant-design/icons';
import { Button, Card, Form, Input, Typography, message } from 'antd';
import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate } from 'react-router-dom';

const { Title, Text } = Typography;

const Register: React.FC = () => {
    const { t } = useTranslation();
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();

    // backend error codes (internal/pkg/code/server.go)
    const ERR_USER_ALREADY_EXIST = 110001;
    const ERR_EMAIL_ALREADY_EXIST = 110002;

    const onFinish = async (values: any) => {
        setLoading(true);
        try {
            await userApi.createUser({
                username: values.username,
                nickname: values.nickname || values.username,
                email: values.email,
                password: values.password,
            });
            message.success(t('auth.register.success'));
            navigate('/login');
        } catch (error: any) {
            console.error('Register error:', error);

            const respCode: number | undefined = error?.response?.data?.code;
            if (respCode === ERR_USER_ALREADY_EXIST) {
                message.error(t('auth.fields.username.alreadyExists'));
                return;
            }
            if (respCode === ERR_EMAIL_ALREADY_EXIST) {
                message.error(t('auth.fields.email.alreadyExists'));
                return;
            }

            // Handle validation errors with field-level messages
            if (error?.response?.data?.details) {
                const details = error.response.data.details;
                // Show field-specific errors (supports backend validation token format: "validation.xxx|k=v")
                Object.keys(details).forEach((field) => {
                    const fieldName = field === 'username' ? t('auth.fields.username.label') :
                        field === 'email' ? t('auth.fields.email.label') :
                            field === 'password' ? t('auth.fields.password.label') :
                                field === 'nickname' ? t('auth.fields.nickname.label') : field;

                    const rawMsg: string = details[field];
                    if (typeof rawMsg === 'string' && rawMsg.startsWith('validation.')) {
                        const [key, ...kvPairs] = rawMsg.split('|');
                        const params: Record<string, string> = {};
                        kvPairs.forEach((pair) => {
                            const idx = pair.indexOf('=');
                            if (idx === -1) return;
                            const k = pair.slice(0, idx).trim();
                            const v = pair.slice(idx + 1).trim();
                            if (k) params[k] = v;
                        });
                        message.error(`${fieldName}: ${t(key as any, params as any)}`);
                        return;
                    }

                    // Fallback (legacy backend messages)
                    message.error(`${fieldName}: ${rawMsg}`);
                });
            } else {
                // For other errors, show server message if present, otherwise fallback.
                const serverMsg: string | undefined = error?.response?.data?.message;
                message.error(serverMsg || t('auth.register.failed'));
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
            <div style={{ position: 'absolute', top: 20, right: 20 }}>
                <LanguageSelect />
            </div>
            <Card
                style={{ width: 450, boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}
                bodyStyle={{ padding: '40px 32px' }}
            >
                <div style={{ textAlign: 'center', marginBottom: 32 }}>
                    <img src="/vite.svg" alt="logo" style={{ height: 44, marginBottom: 16 }} />
                    <Title level={3} style={{ margin: 0 }}>{t('auth.register.title')}</Title>
                    <Text type="secondary">{t('auth.register.subtitle')}</Text>
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
                            { required: true, message: t('auth.fields.username.required') },
                            { min: 3, message: t('auth.fields.username.min') },
                            { max: 32, message: t('auth.fields.username.max') },
                        ]}
                    >
                        <Input prefix={<UserOutlined />} placeholder={`${t('auth.fields.username.placeholder')} (3-32${t('auth.fields.username.label').slice(-2)})`} />
                    </Form.Item>

                    <Form.Item
                        name="nickname"
                        rules={[
                            { max: 32, message: t('auth.fields.nickname.max') },
                        ]}
                    >
                        <Input prefix={<UserOutlined />} placeholder={t('auth.fields.nickname.placeholder')} />
                    </Form.Item>

                    <Form.Item
                        name="email"
                        rules={[
                            { required: true, message: t('auth.fields.email.required') },
                            { type: 'email', message: t('auth.fields.email.invalid') },
                        ]}
                    >
                        <Input prefix={<MailOutlined />} placeholder={t('auth.fields.email.placeholder')} />
                    </Form.Item>

                    <Form.Item
                        name="password"
                        rules={[
                            { required: true, message: t('auth.fields.password.required') },
                            { min: 8, message: t('auth.fields.password.min') },
                            { max: 32, message: t('auth.fields.password.max') },
                        ]}
                    >
                        <Input.Password
                            prefix={<LockOutlined />}
                            placeholder={`${t('auth.fields.password.placeholder')} (8-32${t('auth.fields.username.label').slice(-2)})`}
                        />
                    </Form.Item>

                    <Form.Item
                        name="confirmPassword"
                        dependencies={['password']}
                        rules={[
                            { required: true, message: t('auth.fields.confirmPassword.required') },
                            ({ getFieldValue }) => ({
                                validator(_, value) {
                                    if (!value || getFieldValue('password') === value) {
                                        return Promise.resolve();
                                    }
                                    return Promise.reject(new Error(t('auth.fields.confirmPassword.mismatch')));
                                },
                            }),
                        ]}
                    >
                        <Input.Password
                            prefix={<LockOutlined />}
                            placeholder={t('auth.fields.confirmPassword.placeholder')}
                        />
                    </Form.Item>

                    <Form.Item>
                        <Button type="primary" htmlType="submit" className="register-form-button" block loading={loading}>
                            {t('auth.register.submit')}
                        </Button>
                        <div style={{ marginTop: 16, textAlign: 'center' }}>
                            <Text>{t('auth.register.hasAccount')} </Text>
                            <Link to="/login">{t('auth.register.loginNow')}</Link>
                        </div>
                    </Form.Item>
                </Form>
            </Card>
        </div>
    );
};

export default Register;

