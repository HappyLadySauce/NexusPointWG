import { GlobalOutlined } from '@ant-design/icons';
import type { MenuProps } from 'antd';
import { Dropdown } from 'antd';
import React from 'react';
import { useTranslation } from 'react-i18next';

const LanguageSelect: React.FC = () => {
    const { i18n } = useTranslation();

    const items: MenuProps['items'] = [
        {
            key: 'zh-CN',
            label: '简体中文',
        },
        {
            key: 'en-US',
            label: 'English',
        },
    ];

    const onClick: MenuProps['onClick'] = ({ key }) => {
        i18n.changeLanguage(key);
    };

    return (
        <Dropdown menu={{ items, onClick, selectedKeys: [i18n.language] }} placement="bottomRight">
            <GlobalOutlined style={{ fontSize: '18px', cursor: 'pointer', color: 'rgba(0, 0, 0, 0.45)' }} />
        </Dropdown>
    );
};

export default LanguageSelect;

