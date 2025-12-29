import type { WGPeer } from '@/api/wg';
import { downloadConfig, getMyConfigs, revokeConfig, rotateConfig } from '@/api/wg';
import { DeleteOutlined, DownloadOutlined, QrcodeOutlined, ReloadOutlined } from '@ant-design/icons';
import { Button, Card, Empty, List, message, Modal, Popconfirm, Tag, Tooltip } from 'antd';
import { QRCodeSVG } from 'qrcode.react';
import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

const UserDashboard: React.FC = () => {
    const { t } = useTranslation();
    const [loading, setLoading] = useState(false);
    const [data, setData] = useState<WGPeer[]>([]);
    const [qrModalOpen, setQrModalOpen] = useState(false);
    const [qrContent, setQrContent] = useState('');
    const [qrTitle, setQrTitle] = useState('');

    const fetchData = async () => {
        setLoading(true);
        try {
            const res = await getMyConfigs();
            setData(res.items);
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
    }, []);

    const handleDownload = async (id: string, deviceName: string) => {
        try {
            const res = await downloadConfig(id);
            // Create blob link to download
            const url = window.URL.createObjectURL(new Blob([res as any]));
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', `${deviceName}.conf`);
            document.body.appendChild(link);
            link.click();
            link.remove();
        } catch (error) {
            console.error(error);
            message.error(t('common.downloadFailed'));
        }
    };

    const handleShowQR = async (id: string, deviceName: string) => {
        try {
            const res = await downloadConfig(id);
            const text = await (res as any).text();
            setQrContent(text);
            setQrTitle(deviceName);
            setQrModalOpen(true);
        } catch (error) {
            console.error(error);
            message.error(t('common.error'));
        }
    };

    const handleRotate = async (id: string) => {
        try {
            await rotateConfig(id);
            message.success(t('common.success'));
            fetchData();
        } catch (error) {
            console.error(error);
        }
    };

    const handleRevoke = async (id: string) => {
        try {
            await revokeConfig(id);
            message.success(t('common.success'));
            fetchData();
        } catch (error) {
            console.error(error);
        }
    };

    return (
        <div>
            <div style={{ marginBottom: 16 }}>
                <h2>{t('wg.myConfigs')}</h2>
            </div>

            {loading && data.length === 0 ? (
                <div style={{ textAlign: 'center', padding: 50 }}>Loading...</div>
            ) : data.length === 0 ? (
                <Empty description={t('wg.noConfigs')} />
            ) : (
                <List
                    grid={{ gutter: 16, xs: 1, sm: 1, md: 2, lg: 3, xl: 3, xxl: 4 }}
                    dataSource={data}
                    renderItem={(item) => (
                        <List.Item>
                            <Card
                                title={item.device_name}
                                extra={
                                    <Tag color={item.status === 'active' ? 'success' : 'default'}>
                                        {item.status.toUpperCase()}
                                    </Tag>
                                }
                                actions={[
                                    <Tooltip title={t('common.download')}>
                                        <Button
                                            type="text"
                                            icon={<DownloadOutlined />}
                                            disabled={item.status !== 'active'}
                                            onClick={() => handleDownload(item.id, item.device_name)}
                                        />
                                    </Tooltip>,
                                    <Tooltip title={t('wg.showQR')}>
                                        <Button
                                            type="text"
                                            icon={<QrcodeOutlined />}
                                            disabled={item.status !== 'active'}
                                            onClick={() => handleShowQR(item.id, item.device_name)}
                                        />
                                    </Tooltip>,
                                    <Tooltip title={t('wg.rotateKey')}>
                                        <Popconfirm
                                            title={t('wg.confirmRotate')}
                                            onConfirm={() => handleRotate(item.id)}
                                        >
                                            <Button type="text" icon={<ReloadOutlined />} disabled={item.status !== 'active'} />
                                        </Popconfirm>
                                    </Tooltip>,
                                    <Tooltip title={t('wg.revoke')}>
                                        <Popconfirm
                                            title={t('wg.confirmRevoke')}
                                            onConfirm={() => handleRevoke(item.id)}
                                        >
                                            <Button type="text" danger icon={<DeleteOutlined />} disabled={item.status !== 'active'} />
                                        </Popconfirm>
                                    </Tooltip>,
                                ]}
                            >
                                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                                    <div>
                                        <div style={{ color: 'rgba(0, 0, 0, 0.45)' }}>IP Address</div>
                                        <div>{item.client_ip}</div>
                                    </div>
                                    <div>
                                        <div style={{ color: 'rgba(0, 0, 0, 0.45)' }}>Public Key</div>
                                        <div style={{ wordBreak: 'break-all', fontSize: '12px' }}>
                                            {item.client_public_key}
                                        </div>
                                    </div>
                                    <div>
                                        <div style={{ color: 'rgba(0, 0, 0, 0.45)' }}>Allowed IPs</div>
                                        <div style={{ fontSize: '12px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                                            <Tooltip title={item.allowed_ips || 'Default'}>
                                                {item.allowed_ips || 'Default'}
                                            </Tooltip>
                                        </div>
                                    </div>
                                </div>
                            </Card>
                        </List.Item>
                    )}
                />
            )}

            <Modal
                title={`QR Code - ${qrTitle}`}
                open={qrModalOpen}
                onCancel={() => setQrModalOpen(false)}
                footer={null}
                width={300}
                styles={{ body: { display: 'flex', justifyContent: 'center', padding: '20px' } }}
            >
                {qrContent && (
                    <QRCodeSVG value={qrContent} size={200} />
                )}
            </Modal>
        </div>
    );
};

export default UserDashboard;
