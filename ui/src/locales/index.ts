import enUSCommon from './en-US/common.json';
import zhCNCommon from './zh-CN/common.json';

export const resources = {
    'zh-CN': {
        translation: zhCNCommon
    },
    'en-US': {
        translation: enUSCommon
    }
} as const;

export type Resources = typeof resources['zh-CN'];

