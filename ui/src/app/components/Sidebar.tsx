import React from "react";
import { useAuth } from "../context/AuthContext";
import { cn } from "../lib/utils";
import {
  LayoutDashboard,
  Network,
  Users,
  Settings,
  LogOut,
  Shield,
  Database
} from "lucide-react";

interface SidebarProps extends React.HTMLAttributes<HTMLDivElement> {
    currentTab: string;
    setCurrentTab: (tab: string) => void;
}

export function Sidebar({ className, currentTab, setCurrentTab }: SidebarProps) {
  const { user, logout } = useAuth();

  const menuItems = [
    { id: "dashboard", label: "Dashboard", icon: LayoutDashboard },
    { id: "peers", label: "Peers", icon: Network },
    { id: "ip-pools", label: "IP Pools", icon: Database, adminOnly: true },
    { id: "users", label: "Users", icon: Users, adminOnly: true },
    { id: "settings", label: "Settings", icon: Settings, adminOnly: true },
  ];

  return (
    <div className={cn("pb-12 min-h-screen border-r bg-background", className)}>
      <div className="space-y-4 py-4">
        <div className="px-3 py-2">
          <div className="flex items-center gap-2 px-4 mb-6">
            <Shield className="h-8 w-8 text-primary" />
            <h2 className="text-xl font-bold tracking-tight">NexusPointWG</h2>
          </div>
          <div className="space-y-1">
            {menuItems.map((item) => {
               if (item.adminOnly && user?.role !== 'admin') return null;
               const Icon = item.icon;
               return (
                  <button
                    key={item.id}
                    onClick={() => setCurrentTab(item.id)}
                    className={cn(
                      "w-full flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-all hover:text-primary",
                      currentTab === item.id 
                        ? "bg-secondary text-primary" 
                        : "text-muted-foreground hover:bg-secondary/50"
                    )}
                  >
                    <Icon className="h-4 w-4" />
                    {item.label}
                  </button>
               )
            })}
          </div>
        </div>
      </div>
      
      <div className="absolute bottom-4 px-6 w-full">
         <div className="flex items-center gap-3 mb-4 px-2">
            <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center text-primary font-bold">
                {user?.username.charAt(0).toUpperCase()}
            </div>
            <div className="flex-1 overflow-hidden">
                <p className="text-sm font-medium truncate">{user?.username}</p>
                <p className="text-xs text-muted-foreground truncate capitalize">{user?.role}</p>
            </div>
         </div>
         <button 
            onClick={logout}
            className="flex w-full items-center gap-2 rounded-lg px-2 py-2 text-sm font-medium text-destructive hover:bg-destructive/10 transition-colors"
         >
            <LogOut className="h-4 w-4" />
            Logout
         </button>
      </div>
    </div>
  );
}
