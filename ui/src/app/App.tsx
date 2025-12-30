import React, { useState } from 'react';
import { AuthProvider, useAuth } from './context/AuthContext';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';
import { Peers } from './pages/Peers';
import { UsersPage } from './pages/Users';
import { IPPools } from './pages/IPPools';
import { Sidebar } from './components/Sidebar';
import { Toaster } from './components/ui/sonner';
import { Settings } from './pages/Settings';

function MainApp() {
  const { user, isAuthenticated, isLoading } = useAuth();
  const [currentTab, setCurrentTab] = useState("dashboard");

  if (isLoading) {
    return <div className="flex items-center justify-center min-h-screen">Loading...</div>;
  }

  if (!isAuthenticated) {
    return <Login />;
  }

  return (
    <div className="flex">
      <Sidebar 
        className="w-64 fixed h-full z-10" 
        currentTab={currentTab} 
        setCurrentTab={setCurrentTab} 
      />
      <div className="flex-1 ml-64 min-h-screen bg-slate-50/50">
        {currentTab === 'dashboard' && <Dashboard />}
        {currentTab === 'peers' && <Peers />}
        {currentTab === 'ip-pools' && <IPPools />}
        {currentTab === 'users' && <UsersPage />}
        {currentTab === 'settings' && <Settings />}
      </div>
    </div>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <MainApp />
      <Toaster />
    </AuthProvider>
  );
}
