import React from 'react';
import {
  Inbox,
  Zap,
  FolderClosed,
  Plus,
  LogOut,
  LogIn,
  SlidersHorizontal,
  Layers,
} from 'lucide-react';
import { Button, Badge } from '@/shared/components/ui';
import type { User, Workspace } from '@/shared/api';
import { useTaskStore, useInboxTasks } from '@/features/task-navigation';

export interface SidebarProps {
  user: User | null;
  workspace: Workspace | null;
  inboxCount?: number;
  activeNav?: 'inbox' | 'today' | 'projects';
  onNavChange?: (nav: 'inbox' | 'today' | 'projects') => void;
  onOpenQuickCapture?: () => void;
  onOpenAuth: () => void;
  onLogout: () => void;
  showUiCatalog: boolean;
  onToggleUiCatalog: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({
  user,
  workspace,
  inboxCount,
  activeNav,
  onNavChange,
  onOpenQuickCapture,
  onOpenAuth,
  onLogout,
  showUiCatalog,
  onToggleUiCatalog,
}) => {
  const { setQuickCaptureOpen, activeNav: storeActiveNav, setActiveNav: storeSetActiveNav } = useTaskStore();
  const { data: inboxTasks } = useInboxTasks();

  const currentInboxCount =
    inboxCount !== undefined
      ? inboxCount
      : (inboxTasks?.filter((t) => !t.isCompleted).length ?? 0);
  const currentNav = activeNav ?? storeActiveNav;
  const handleNavChange = onNavChange ?? storeSetActiveNav;
  const handleOpenQuickCapture = onOpenQuickCapture ?? (() => setQuickCaptureOpen(true));

  return (
    <aside className="w-60 shrink-0 h-full flex flex-col justify-between py-6 px-4 select-none">
      {/* Top section: Workspace & Quick Action */}
      <div className="space-y-6">
        {/* Workspace Brand / Selector */}
        <div className="flex items-center justify-between px-2">
          <div className="flex items-center gap-2.5">
            <div className="w-7 h-7 rounded-lg bg-accent-solid text-surface flex items-center justify-center font-serif text-base font-semibold shadow-washi-sm">
              T
            </div>
            <div className="flex flex-col">
              <span className="font-serif text-lg font-normal tracking-tight text-primary leading-tight">
                {workspace?.name || 'Личное'}
              </span>
              <span className="text-[10px] font-mono uppercase tracking-wider text-muted">
                {workspace?.type === 'personal' ? 'Personal GTD' : 'Workspace'}
              </span>
            </div>
          </div>
        </div>

        {/* Quick Capture Button */}
        <Button
          variant="outline"
          size="sm"
          onClick={handleOpenQuickCapture}
          className="w-full justify-between bg-surface/70 hover:bg-surface border-subtle shadow-washi-sm text-primary group"
        >
          <span className="flex items-center gap-2 text-xs font-medium">
            <Plus className="w-3.5 h-3.5 text-accent-solid" />
            Быстрый захват
          </span>
          <kbd className="font-mono text-[10px] text-muted bg-canvas px-1.5 py-0.5 rounded border border-subtle">
            C
          </kbd>
        </Button>

        {/* Navigation Section */}
        <div className="space-y-1">
          <div className="px-2 pb-1 text-[10px] font-mono uppercase tracking-wider text-muted">
            Горизонты
          </div>

          {/* Inbox */}
          <button
            type="button"
            onClick={() => handleNavChange('inbox')}
            className={`w-full flex items-center justify-between px-2.5 py-2 rounded-lg text-sm transition-all duration-150 ${
              currentNav === 'inbox' && !showUiCatalog
                ? 'bg-accent-surface font-medium text-primary shadow-washi-sm'
                : 'text-primary hover:bg-surface/60'
            }`}
          >
            <span className="flex items-center gap-2.5">
              <Inbox className="w-4 h-4 stroke-[1.8] text-muted" />
              <span>Входящие</span>
            </span>
            {currentInboxCount > 0 && (
              <Badge variant="subtle" size="sm" className="font-mono text-[11px] px-1.5 py-0">
                {currentInboxCount}
              </Badge>
            )}
          </button>

          {/* Today */}
          <button
            type="button"
            onClick={() => handleNavChange('today')}
            className={`w-full flex items-center justify-between px-2.5 py-2 rounded-lg text-sm transition-all duration-150 ${
              currentNav === 'today' && !showUiCatalog
                ? 'bg-accent-surface font-medium text-primary shadow-washi-sm'
                : 'text-primary hover:bg-surface/60 opacity-80'
            }`}
          >
            <span className="flex items-center gap-2.5">
              <Zap className="w-4 h-4 stroke-[1.8] text-muted" />
              <span>Сегодня</span>
            </span>
            <span className="text-[10px] font-mono text-muted">Sprint 0.3</span>
          </button>

          {/* Projects */}
          <button
            type="button"
            onClick={() => handleNavChange('projects')}
            className={`w-full flex items-center justify-between px-2.5 py-2 rounded-lg text-sm transition-all duration-150 ${
              currentNav === 'projects' && !showUiCatalog
                ? 'bg-accent-surface font-medium text-primary shadow-washi-sm'
                : 'text-primary hover:bg-surface/60 opacity-80'
            }`}
          >
            <span className="flex items-center gap-2.5">
              <FolderClosed className="w-4 h-4 stroke-[1.8] text-muted" />
              <span>Проекты</span>
            </span>
            <span className="text-[10px] font-mono text-muted">Sprint 0.3</span>
          </button>
        </div>

        {/* UI-Kit Showcase Toggle */}
        <div className="pt-2 border-t border-subtle/50">
          <button
            type="button"
            onClick={onToggleUiCatalog}
            className={`w-full flex items-center justify-between px-2.5 py-2 rounded-lg text-xs transition-all duration-150 ${
              showUiCatalog
                ? 'bg-accent-surface font-medium text-primary'
                : 'text-muted hover:text-primary hover:bg-surface/60'
            }`}
          >
            <span className="flex items-center gap-2">
              <Layers className="w-3.5 h-3.5" />
              <span>Витрина UI-Kit</span>
            </span>
            <SlidersHorizontal className="w-3 h-3 text-muted" />
          </button>
        </div>
      </div>

      {/* Bottom section: User Profile / Auth */}
      <div className="pt-4 border-t border-subtle/60">
        {user ? (
          <div className="space-y-2">
            <div className="flex items-center justify-between px-2">
              <div className="flex items-center gap-2.5 min-w-0">
                <div className="w-7 h-7 rounded-full bg-accent-surface text-primary flex items-center justify-center text-xs font-medium font-serif shrink-0 border border-subtle">
                  {user.displayName.charAt(0).toUpperCase()}
                </div>
                <div className="flex flex-col min-w-0">
                  <span className="text-xs font-medium text-primary truncate leading-tight">
                    {user.displayName}
                  </span>
                  <span className="text-[10px] text-muted truncate">{user.email}</span>
                </div>
              </div>
              <button
                type="button"
                onClick={onLogout}
                title="Выйти"
                className="p-1 rounded-md text-muted hover:text-primary hover:bg-surface/60 transition-colors"
              >
                <LogOut className="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        ) : (
          <Button
            variant="outline"
            size="sm"
            onClick={onOpenAuth}
            className="w-full text-xs justify-center"
            leftIcon={<LogIn className="w-3.5 h-3.5" />}
          >
            Войти / Регистрация
          </Button>
        )}
      </div>
    </aside>
  );
};
