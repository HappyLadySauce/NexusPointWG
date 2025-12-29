import { listUsers, type User } from '@/api/user';
import type { WGPeer, WGPeerListQuery } from '@/api/wg';
import { createPeer, deletePeer, getPeers, updatePeer } from '@/api/wg';
import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons';
import { Button, Form, Input, message, Modal, Popconfirm, Select, Table, Tag, Tooltip } from 'antd';
import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

const Peers: React.FC = () => {
    const { t } = useTranslation();
    const [loading, setLoading] = useState(false);
    const [data, setData] = useState<WGPeer[]>([]);
    const [total, setTotal] = useState(0);
    const [params, setParams] = useState<WGPeerListQuery>({
        offset: 0,
        limit: 10,
    });

    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingPeer, setEditingPeer] = useState<WGPeer | null>(null);
    const [form] = Form.useForm();
    const [userOptions, setUserOptions] = useState<User[]>([]);
    const [userSearchLoading, setUserSearchLoading] = useState(false);

    const fetchData = async () => {
        setLoading(true);
        try {
            const res = await getPeers(params);
            setData(res.items);
            setTotal(res.total);
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
    }, [params]);

    const handleTableChange = (pagination: any) => {
        setParams({
            ...params,
            offset: (pagination.current - 1) * pagination.pageSize,
            limit: pagination.pageSize,
        });
    };

    const handleUserSearch = async (searchValue: string) => {
        if (!searchValue || searchValue.length < 2) {
            setUserOptions([]);
            return;
        }
        setUserSearchLoading(true);
        try {
            const res = await listUsers({ username: searchValue, limit: 50 });
            setUserOptions(res.items);
        } catch (error) {
            console.error(error);
        } finally {
            setUserSearchLoading(false);
        }
    };

    const handleCreate = () => {
        setEditingPeer(null);
        form.resetFields();
        setUserOptions([]);
        setIsModalOpen(true);
    };

    const handleEdit = (record: WGPeer) => {
        setEditingPeer(record);
        form.setFieldsValue({
            username: record.username,
            device_name: record.device_name,
            allowed_ips: record.allowed_ips,
            persistent_keepalive: record.persistent_keepalive,
            status: record.status,
        });
        setIsModalOpen(true);
    };

    const handleDelete = async (id: string) => {
        try {
            await deletePeer(id);
            message.success(t('common.success'));
            fetchData();
        } catch (error) {
            console.error(error);
        }
    };

    const handleOk = async () => {
        try {
            const values = await form.validateFields();
            if (editingPeer) {
                await updatePeer(editingPeer.id, values);
            } else {
                await createPeer(values);
            }
            message.success(t('common.success'));
            setIsModalOpen(false);
            fetchData();
        } catch (error) {
            console.error(error);
        }
    };

    const columns = [
        {
            title: t('wg.device'),
            dataIndex: 'device_name',
            key: 'device_name',
        },
        {
            title: t('user.username'),
            dataIndex: 'username',
            key: 'username',
        },
        {
            title: t('wg.clientIP'),
            dataIndex: 'client_ip',
            key: 'client_ip',
        },
        {
            title: t('wg.allowedIPs'),
            dataIndex: 'allowed_ips',
            key: 'allowed_ips',
            render: (text: string) => (
                <Tooltip title={text}>
                    <div style={{ maxWidth: 150, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {text || '-'}
                    </div>
                </Tooltip>
            ),
        },
        {
            title: t('wg.status'),
            dataIndex: 'status',
            key: 'status',
            render: (status: string) => (
                <Tag color={status === 'active' ? 'success' : 'error'}>
                    {status.toUpperCase()}
                </Tag>
            ),
        },
        {
            title: t('common.action'),
            key: 'action',
            render: (_: any, record: WGPeer) => (
                <div style={{ display: 'flex', gap: 8 }}>
                    <Button
                        type="text"
                        icon={<EditOutlined />}
                        onClick={() => handleEdit(record)}
                    />
                    <Popconfirm
                        title={t('common.confirmDelete')}
                        onConfirm={() => handleDelete(record.id)}
                    >
                        <Button type="text" danger icon={<DeleteOutlined />} />
                    </Popconfirm>
                </div>
            ),
        },
    ];

    return (
        <div>
            <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
                <div style={{ display: 'flex', gap: 8 }}>
                    <Input.Search
                        placeholder={t('common.search')}
                        onSearch={(value) => setParams({ ...params, device_name: value, offset: 0 })}
                        style={{ width: 200 }}
                    />
                </div>
                <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
                    {t('common.create')}
                </Button>
            </div>

            <Table
                columns={columns}
                dataSource={data}
                rowKey="id"
                pagination={{
                    current: (params.offset || 0) / (params.limit || 10) + 1,
                    pageSize: params.limit,
                    total: total,
                }}
                loading={loading}
                onChange={handleTableChange}
            />

            <Modal
                title={editingPeer ? t('wg.editPeer') : t('wg.createPeer')}
                open={isModalOpen}
                onOk={handleOk}
                onCancel={() => setIsModalOpen(false)}
            >
                <Form
                    form={form}
                    layout="vertical"
                >
                    <Form.Item
                        name="username"
                        label={t('user.username')}
                        rules={[{ required: true }]}
                    >
                        {editingPeer ? (
                            <Input disabled placeholder={editingPeer.username} />
                        ) : (
                            <Select
                                showSearch
                                placeholder={t('wg.placeholder.username')}
                                filterOption={false}
                                onSearch={handleUserSearch}
                                loading={userSearchLoading}
                                notFoundContent={userSearchLoading ? t('common.loading') : null}
                            >
                                {userOptions.map((user) => (
                                    <Select.Option key={user.username} value={user.username}>
                                        {user.username} ({user.nickname || user.email})
                                    </Select.Option>
                                ))}
                            </Select>
                        )}
                    </Form.Item>
                    <Form.Item
                        name="device_name"
                        label={t('wg.device')}
                        rules={[{ required: true }]}
                    >
                        <Input />
                    </Form.Item>
                    <Form.Item
                        name="allowed_ips"
                        label={t('wg.allowedIPs')}
                        tooltip={t('wg.tip.allowedIPs')}
                    >
                        <Input placeholder="10.10.0.0/24, ..." />
                    </Form.Item>
                    <Form.Item
                        name="persistent_keepalive"
                        label={t('wg.keepalive')}
                    >
                        <Input type="number" placeholder="25" />
                    </Form.Item>
                    {editingPeer && (
                        <Form.Item
                            name="status"
                            label={t('wg.status')}
                        >
                            <Select>
                                <Select.Option value="active">Active</Select.Option>
                                <Select.Option value="revoked">Revoked</Select.Option>
                            </Select>
                        </Form.Item>
                    )}
                </Form>
            </Modal>
        </div>
    );
};

export default Peers;
