import React from "react";
import { Shield } from "lucide-react";
import { LanguageSwitcherCompact } from "./LanguageSwitcherCompact";
import { UserMenu } from "./UserMenu";

/**
 * TopBar 组件
 * 顶部导航栏，包含应用标题、语言切换和用户菜单
 */
export function TopBar() {
  return (
    <div className="fixed top-0 left-0 right-0 h-16 border-b bg-background z-20 flex items-center justify-between px-6">
      {/* Left: App Title */}
      <div className="flex items-center gap-2">
        <Shield className="h-6 w-6 text-primary" />
        <h1 className="text-lg font-bold tracking-tight">NexusPointWG</h1>
      </div>

      {/* Right: Language Switcher + User Menu */}
      <div className="flex items-center gap-3">
        <LanguageSwitcherCompact />
        <UserMenu />
      </div>
    </div>
  );
}

