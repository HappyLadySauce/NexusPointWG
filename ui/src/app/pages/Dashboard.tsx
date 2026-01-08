import { Activity, Users as UsersIcon } from "lucide-react";
import React, { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { api, WGPeerResponse } from "../services/api";
import { useAuth } from "../context/AuthContext";

export function Dashboard() {
  const { user } = useAuth();
  const { t } = useTranslation('dashboard');
  const isAdmin = user?.role === "admin";
  const [peers, setPeers] = useState<WGPeerResponse[]>([]);
  const [stats, setStats] = useState({
    totalPeers: 0,
    activePeers: 0,
    totalUsers: 0,
  });

  useEffect(() => {
    const loadData = async () => {
      try {
        const response = await api.wg.listPeers();
        const peerList = response.items || [];
        setPeers(peerList);
        
        const active = peerList.filter(p => p.status === 'active').length;
        
        // Fetch users data only for admin
        let totalUsers = 0;
        if (isAdmin) {
        try {
          const usersResponse = await api.users.list();
          totalUsers = usersResponse.total || 0;
        } catch (e) {
          console.error("Failed to fetch users:", e);
          }
        }
        
        setStats({
          totalPeers: response.total || peerList.length,
          activePeers: active,
          totalUsers: totalUsers,
        });
      } catch (e) {
        console.error(e);
      }
    };
    loadData();
  }, [isAdmin]);

  const { t: tCommon } = useTranslation('common');

  return (
    <div className="space-y-6 p-8 bg-slate-50/50">
      <h1 className="text-3xl font-bold tracking-tight">{t('title')}</h1>
      
      <div className={`grid gap-4 ${isAdmin ? 'md:grid-cols-2 lg:grid-cols-3' : 'md:grid-cols-2'}`}>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('stats.totalPeers')}</CardTitle>
            <UsersIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalPeers}</div>
            <p className="text-xs text-muted-foreground">
              {stats.activePeers} {t('stats.activeNow')}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('stats.activePeers')}</CardTitle>
            <Activity className="h-4 w-4 text-emerald-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-emerald-600">{stats.activePeers}</div>
            <p className="text-xs text-muted-foreground">
              {t('stats.currentlyConnected')}
            </p>
          </CardContent>
        </Card>
        {isAdmin && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('stats.totalUsers')}</CardTitle>
            <UsersIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalUsers}</div>
            <p className="text-xs text-muted-foreground">
              {t('stats.registeredUsers')}
            </p>
          </CardContent>
        </Card>
        )}
      </div>

      <Card>
          <CardHeader>
          <CardTitle>{t('recentPeers.title')}</CardTitle>
          </CardHeader>
          <CardContent>
          <div className="space-y-4">
            {peers.slice(0, 10).map(peer => (
              <div className="flex items-center justify-between p-3 border rounded-lg" key={peer.id}>
                        <div className="space-y-1">
                  <p className="text-sm font-medium leading-none">{peer.device_name}</p>
                            <p className="text-sm text-muted-foreground">
                    {peer.client_ip} • {peer.username || tCommon('common.na')}
                            </p>
                        </div>
                        <div className="ml-auto font-medium text-sm">
                            {peer.status === 'active' ? (
                                <span className="text-emerald-600">{tCommon('status.active')}</span>
                            ) : (
                                <span className="text-slate-400">{tCommon('status.inactive')}</span>
                            )}
                        </div>
                    </div>
                ))}
            {peers.length === 0 && (
              <p className="text-center text-muted-foreground py-8">{t('recentPeers.noPeersFound')}</p>
            )}
            </div>
          </CardContent>
        </Card>
    </div>
  );
}
