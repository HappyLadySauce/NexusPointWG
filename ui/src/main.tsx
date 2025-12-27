import { App as AntdApp, ConfigProvider } from 'antd'
import enUS from 'antd/locale/en_US'
import zhCN from 'antd/locale/zh_CN'
import React, { useEffect, useState } from 'react'
import ReactDOM from 'react-dom/client'
import { useTranslation } from 'react-i18next'
import App from './App.tsx'
import './i18n/config'
import './styles/global.css'

// Wrapper component to handle Ant Design locale switching
const AppProvider = () => {
  const { i18n } = useTranslation()
  const [locale, setLocale] = useState(zhCN)

  useEffect(() => {
    if (i18n.language === 'en-US' || i18n.language === 'en') {
      setLocale(enUS)
    } else {
      setLocale(zhCN)
    }
  }, [i18n.language])

  return (
    <ConfigProvider locale={locale}>
      <AntdApp>
        <App />
      </AntdApp>
    </ConfigProvider>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <AppProvider />
  </React.StrictMode>,
)
