import type { SystemStats } from '@/types/api';
import {
    ArrowDownOutlined,
    ArrowUpOutlined,
    CloudServerOutlined,
    UserOutlined
} from '@ant-design/icons';
import { Avatar, Card, Col, List, Progress, Row, Statistic, Typography } from 'antd';
import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

const { Title, Text } = Typography;

// Helper to format bytes
const formatBytes = (bytes: number, decimals = 2) => {
    if (!+bytes) return '0 B';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
};

const AdminDashboard: React.FC = () => {
    const { t } = useTranslation();
    const [stats, setStats] = useState<SystemStats | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchStats = async () => {
            try {
                // const data = await adminApi.getStats();
                // setStats(data);

                // Mock data for now until API is ready
                setStats({
                    total_users: 128,
                    active_peers: 42,
                    total_peers: 156,
                    total_traffic_rx: 1234567890,
                    total_traffic_tx: 9876543210,
                    cpu_usage: 45,
                    memory_usage: 60
                });
            } catch (error) {
                console.error('Failed to fetch stats:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchStats();
    }, []);

    // Mock activities
    const activities = [
        {
            user: 'Alice',
            action: 'created a new peer',
            time: '10 mins ago',
        },
        {
            user: 'Bob',
            action: 'updated profile',
            time: '1 hour ago',
        },
        {
            user: 'Charlie',
            action: 'downloaded config',
            time: '2 hours ago',
        },
        {
            user: 'Admin',
            action: 'deleted user Dave',
            time: '1 day ago',
        },
    ];

    return (
        <div>
            <Title level={2}>{t('menu.dashboard')}</Title>

            <Row gutter={[16, 16]}>
                <Col xs={24} sm={12} lg={6}>
                    <Card loading={loading}>
                        <Statistic
                            title={t('dashboard.totalUsers')}
                            value={stats?.total_users}
                            prefix={<UserOutlined />}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <Card loading={loading}>
                        <Statistic
                            title={t('dashboard.activePeers')}
                            value={stats?.active_peers}
                            suffix={`/ ${stats?.total_peers}`}
                            prefix={<CloudServerOutlined />}
                            valueStyle={{ color: '#3f8600' }}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <Card loading={loading}>
                        <Statistic
                            title="Total RX"
                            value={formatBytes(stats?.total_traffic_rx || 0)}
                            prefix={<ArrowDownOutlined />}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <Card loading={loading}>
                        <Statistic
                            title="Total TX"
                            value={formatBytes(stats?.total_traffic_tx || 0)}
                            prefix={<ArrowUpOutlined />}
                        />
                    </Card>
                </Col>
            </Row>

            <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
                <Col xs={24} lg={12}>
                    <Card title="System Status" loading={loading}>
                        <div style={{ marginBottom: 16 }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                                <Text>CPU</Text>
                                <Text>{stats?.cpu_usage}%</Text>
                            </div>
                            <Progress percent={stats?.cpu_usage} status={stats?.cpu_usage && stats.cpu_usage > 80 ? 'exception' : 'active'} />
                        </div>
                        <div>
                            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                                <Text>Memory</Text>
                                <Text>{stats?.memory_usage}%</Text>
                            </div>
                            <Progress percent={stats?.memory_usage} strokeColor="#faad14" />
                        </div>
                    </Card>
                </Col>

                <Col xs={24} lg={12}>
                    <Card title="Recent Activity" loading={loading}>
                        <List
                            itemLayout="horizontal"
                            dataSource={activities}
                            renderItem={(item) => (
                                <List.Item>
                                    <List.Item.Meta
                                        avatar={<Avatar style={{ backgroundColor: '#1890ff' }}>{item.user[0]}</Avatar>}
                                        title={<Text strong>{item.user}</Text>}
                                        description={
                                            <span>
                                                {item.action} - <Text type="secondary">{item.time}</Text>
                                            </span>
                                        }
                                    />
                                </List.Item>
                            )}
                        />
                    </Card>
                </Col>
            </Row>
        </div>
    );
};

export default AdminDashboard;
