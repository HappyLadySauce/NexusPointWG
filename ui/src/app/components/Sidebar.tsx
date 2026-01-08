import React from "react";
import { useAuth } from "../context/AuthContext";
import { cn } from "../lib/utils";
import { useTranslation } from "react-i18next";
import {
  LayoutDashboard,
  Network,
  Users,
  Settings,
  Database
} from "lucide-react";

interface SidebarProps extends React.HTMLAttributes<HTMLDivElement> {
    currentTab: string;
    setCurrentTab: (tab: string) => void;
}

export function Sidebar({ className, currentTab, setCurrentTab }: SidebarProps) {
  const { user } = useAuth();
  const { t } = useTranslation('common');

  const menuItems = [
    { id: "dashboard", label: t('menu.dashboard'), icon: LayoutDashboard },
    { id: "peers", label: t('menu.peers'), icon: Network },
    { id: "ip-pools", label: t('menu.ipPools'), icon: Database, adminOnly: true },
    { id: "users", label: t('menu.users'), icon: Users, adminOnly: true },
    { id: "settings", label: t('menu.settings'), icon: Settings, adminOnly: true },
  ];

  return (
    <div className={cn("border-r bg-background", className)}>
      <div className="space-y-4 py-4">
        <div className="px-3 py-2">
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
    </div>
  );
}
