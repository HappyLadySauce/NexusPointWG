import React, { useEffect, useState } from "react";
import { api, Peer } from "../services/api";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { Activity, ArrowDown, ArrowUp, Users as UsersIcon } from "lucide-react";
import { Bar, BarChart, ResponsiveContainer, XAxis, YAxis, Tooltip } from "recharts";

export function Dashboard() {
  const [peers, setPeers] = useState<Peer[]>([]);
  const [stats, setStats] = useState({
    totalPeers: 0,
    activePeers: 0,
    totalRx: 0,
    totalTx: 0,
  });

  useEffect(() => {
    const loadData = async () => {
      try {
        const data = await api.wg.listPeers();
        setPeers(data);
        
        const active = data.filter(p => p.status === 'active').length;
        const rx = data.reduce((acc, p) => acc + p.transfer_rx, 0);
        const tx = data.reduce((acc, p) => acc + p.transfer_tx, 0);
        
        setStats({
          totalPeers: data.length,
          activePeers: active,
          totalRx: rx,
          totalTx: tx,
        });
      } catch (e) {
        console.error(e);
      }
    };
    loadData();
  }, []);

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const chartData = peers.map(p => ({
    name: p.name,
    rx: parseFloat((p.transfer_rx / (1024 * 1024)).toFixed(2)), // MB
    tx: parseFloat((p.transfer_tx / (1024 * 1024)).toFixed(2)), // MB
  })).slice(0, 5); // Top 5

  return (
    <div className="space-y-6 p-8 bg-slate-50/50 min-h-screen">
      <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
      
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Peers</CardTitle>
            <UsersIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalPeers}</div>
            <p className="text-xs text-muted-foreground">
              {stats.activePeers} active now
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Download</CardTitle>
            <ArrowDown className="h-4 w-4 text-emerald-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatBytes(stats.totalRx)}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Upload</CardTitle>
            <ArrowUp className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatBytes(stats.totalTx)}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">System Status</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-emerald-600">Online</div>
            <p className="text-xs text-muted-foreground">
              WireGuard Service Running
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        <Card className="col-span-4">
          <CardHeader>
            <CardTitle>Top Data Usage (MB)</CardTitle>
          </CardHeader>
          <CardContent className="pl-2">
            <div className="h-[300px]">
              <ResponsiveContainer width="100%"Tb height="100%">
                <BarChart data={chartData}>
                  <XAxis dataKey="name" stroke="#888888" fontSize={12} tickLine={false} axisLine={false} />
                  <YAxis stroke="#888888" fontSize={12} tickLine={false} axisLine={false} tickFormatter={(value) => `${value}MB`} />
                  <Tooltip />
                  <Bar dataKey="rx" fill="#10b981" radius={[4, 4, 0, 0]} name="Download" />
                  <Bar dataKey="tx" fill="#3b82f6"yb radius={[4, 4, 0, 0]} name="Upload" />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
        <Card className="col-span-3">
           <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-8">
                {peers.slice(0, 5).map(peer => (
                    <div className="flex items-center" key={peer.id}>
                        <div className="space-y-1">
                            <p className="text-sm font-medium leading-none">{peer.name}</p>
                            <p className="text-sm text-muted-foreground">
                                {peer.latest_handshake ? new Date(peer.latest_handshake).toLocaleString() : 'Never connected'}
                            </p>
                        </div>
                        <div className="ml-auto font-medium text-sm">
                            {peer.status === 'active' ? (
                                <span className="text-emerald-600">Active</span>
                            ) : (
                                <span className="text-slate-400">Inactive</span>
                            )}
                        </div>
                    </div>
                ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
