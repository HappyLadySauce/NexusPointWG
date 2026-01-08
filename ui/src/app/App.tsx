import React, { useState } from 'react';
import { AuthProvider, useAuth } from './context/AuthContext';
import { I18nProvider } from './context/I18nContext';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';
import { Peers } from './pages/Peers';
import { UsersPage } from './pages/Users';
import { IPPools } from './pages/IPPools';
import { Sidebar } from './components/Sidebar';
import { TopBar } from './components/TopBar';
import { Toaster } from './components/ui/sonner';
import { Settings } from './pages/Settings';
import { useTranslation } from 'react-i18next';
import { cn } from './lib/utils';

function MainApp() {
  const { user, isAuthenticated, isLoading } = useAuth();
  const { t } = useTranslation('common');
  const [currentTab, setCurrentTab] = useState("dashboard");

  if (isLoading) {
    return <div className="flex items-center justify-center min-h-screen">{t('status.loading')}</div>;
  }

  if (!isAuthenticated) {
    return <Login />;
  }

  // Only Settings page needs scrolling
  const needsScrolling = currentTab === 'settings';

  return (
    <div className="flex flex-col h-screen overflow-hidden">
      <TopBar />
      <div className="flex flex-1 pt-16 overflow-hidden">
        <Sidebar 
          className="w-64 fixed top-16 h-[calc(100vh-4rem)] z-10" 
          currentTab={currentTab} 
          setCurrentTab={setCurrentTab} 
        />
        <div className={cn(
          "flex-1 ml-64 h-[calc(100vh-4rem)] bg-slate-50/50",
          needsScrolling ? "overflow-y-auto" : "overflow-hidden"
        )}>
          {currentTab === 'dashboard' && <Dashboard />}
          {currentTab === 'peers' && <Peers />}
          {currentTab === 'ip-pools' && <IPPools />}
          {currentTab === 'users' && <UsersPage />}
          {currentTab === 'settings' && <Settings />}
        </div>
      </div>
    </div>
  );
}

export default function App() {
  return (
    <I18nProvider>
      <AuthProvider>
        <MainApp />
        <Toaster />
      </AuthProvider>
    </I18nProvider>
  );
}
