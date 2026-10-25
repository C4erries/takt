import React, { useState } from 'react';
import {
  Inbox, Zap, FolderClosed, Plus, LogOut, LogIn,
  SlidersHorizontal, Layers, X, Sparkles, Search, PanelLeftClose, Palette, Tag, Archive,
} from 'lucide-react';
import { Button, Badge, Input } from '@/shared/components/ui';
import type { User, Workspace } from '@/shared/api';
import { WorkspaceSwitcher } from '@/features/workspace-switcher';
import { MiniCalendar } from '@/widgets/mini-calendar';
import { useTaskStore } from '@/features/task-navigation';
import { useThemeStore, getTheme } from '@/features/theme-switcher';
import { useProjectsQuery, useCreateProjectMutation } from '@/entities/project';
import { useTodayTasksQuery, useInboxTasks, useWorkspaceTagsQuery } from '@/entities/task';
import { useAutomationsQuery } from '@/entities/automation';
import { useArchivedTasksQuery } from '@/features/archive';

export interface SidebarProps {
  user: User | null;
  workspace: Workspace | null;
  inboxCount?: number;
  activeNav?: string;
  onNavChange?: (nav: string) => void;
  onOpenQuickCapture?: () => void;
  onOpenCommandPalette?: () => void;
  isCollapsed?: boolean;
  onToggleCollapse?: () => void;
  onOpenAuth: () => void;
  onLogout: () => void;
  showUiCatalog: boolean;
  onToggleUiCatalog: () => void;
  onOpenAutomations?: () => void;
  onWorkspaceChange?: (workspace: Workspace) => void;
}

export const Sidebar: React.FC<SidebarProps> = ({
  user, workspace, inboxCount, activeNav, onNavChange, onOpenQuickCapture,
  onOpenCommandPalette, isCollapsed = false, onToggleCollapse, onOpenAuth, onLogout,
  showUiCatalog, onToggleUiCatalog, onWorkspaceChange,
}) => {
  const {
    setQuickCaptureOpen, activeNav: storeActiveNav, setActiveNav: storeSetActiveNav,
    setSelectedProjectId, setActiveBoardId,
  } = useTaskStore();
  const { theme, setPickerOpen } = useThemeStore();
  const themeInfo = getTheme(theme);
  const { data: inboxTasks } = useInboxTasks();
  const { data: todayTasks = [] } = useTodayTasksQuery(workspace?.id);
  const { data: projects = [] } = useProjectsQuery(workspace?.id);
  const { data: automations = [] } = useAutomationsQuery(workspace?.id);
  const { data: archivedTasks = [] } = useArchivedTasksQuery(workspace?.id);
  const { data: workspaceTags = [] } = useWorkspaceTagsQuery(workspace?.id);
  const createProjectMutation = useCreateProjectMutation(workspace?.id);

  const [isCreatingProject, setIsCreatingProject] = useState(false);
  const [newProjectName, setNewProjectName] = useState('');

  const todayCount = todayTasks.filter((t) => !t.isCompleted).length;
  const activeAutomationsCount = automations.filter((a) => a.isEnabled).length;
  const currentInboxCount = inboxCount !== undefined ? inboxCount : (inboxTasks?.filter((t) => !t.isCompleted).length ?? 0);
  const currentNav = activeNav ?? storeActiveNav;
  const handleNavChange = onNavChange ?? storeSetActiveNav;
  const handleOpenQuickCapture = onOpenQuickCapture ?? (() => setQuickCaptureOpen(true));

  const handleSelectProject = (projectId: string) => {
    setSelectedProjectId(projectId);
    setActiveBoardId(null);
    handleNavChange(`project:${projectId}`);
  };

  const handleCreateProject = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = newProjectName.trim();
    if (!trimmed || createProjectMutation.isPending) return;
    try {
      const created = await createProjectMutation.mutateAsync({ name: trimmed });
      setNewProjectName('');
      setIsCreatingProject(false);
      handleSelectProject(created.id);
    } catch {
      // Handled by react-query
    }
  };

  return (
    <aside
      className={`shrink-0 h-full flex flex-col justify-between py-6 select-none transition-all duration-300 ease-in-out bg-surface/85 border-r border-subtle/70 ${
        isCollapsed
          ? 'w-0 -translate-x-full overflow-hidden opacity-0 p-0 pointer-events-none'
          : 'w-60 translate-x-0 opacity-100 px-4'
      }`}
    >
      {/* Top section: Workspace, Quick Action, Navigation */}
      <div className="space-y-6 overflow-y-auto pr-0.5">
        {/* Workspace Switcher & Collapse Button */}
        <div className="flex items-center justify-between gap-1">
          <div className="flex-1 min-w-0">
            <WorkspaceSwitcher
              currentWorkspace={workspace}
              currentUser={user}
              onWorkspaceChange={onWorkspaceChange}
            />
          </div>
          {onToggleCollapse && (
            <button
              type="button"
              onClick={onToggleCollapse}
              className="p-1 rounded-md text-muted hover:text-primary hover:bg-surface/80 transition-colors shrink-0"
              title="Свернуть сайдбар ([)"
            >
              <PanelLeftClose className="w-4 h-4" />
            </button>
          )}
        </div>

        {/* Search / Command Palette Button */}
        <button
          type="button"
          onClick={onOpenCommandPalette}
          className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-xs bg-surface/70 hover:bg-surface border border-subtle shadow-washi-sm text-muted hover:text-primary transition-all group"
          title="Поиск задач, проектов и команд (⌘K / Ctrl+K)"
        >
          <span className="flex items-center gap-2">
            <Search className="w-3.5 h-3.5 text-muted group-hover:text-primary transition-colors" />
            <span>Поиск / Команды</span>
          </span>
          <kbd className="font-mono text-[10px] text-muted bg-canvas px-1.5 py-0.5 rounded border border-subtle">
            ⌘K
          </kbd>
        </button>

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

        {/* Horizons Navigation */}
        <div className="space-y-1">
          <div className="px-2 pb-1 text-[10px] font-mono uppercase tracking-wider text-muted">
            Горизонты
          </div>
          {[
            { id: 'inbox', label: 'Входящие', icon: Inbox, count: currentInboxCount, badgeVariant: 'subtle' as const },
            { id: 'today', label: 'Сегодня', icon: Zap, count: todayCount, badgeVariant: 'persimmon' as const },
            { id: 'tags', label: 'Теги', icon: Tag, count: workspaceTags.length, badgeVariant: 'subtle' as const },
            { id: 'automations', label: 'Автоматизации', icon: Sparkles, count: activeAutomationsCount, badgeVariant: 'subtle' as const, iconColor: 'text-accent-solid' },
            { id: 'archive', label: 'Архив', icon: Archive, count: archivedTasks.length, badgeVariant: 'subtle' as const },
          ].map(({ id, label, icon: Icon, count, badgeVariant, iconColor }) => {
            const isSelected = currentNav === id && !showUiCatalog;
            return (
              <button
                key={id}
                type="button"
                onClick={() => { setSelectedProjectId(null); setActiveBoardId(null); handleNavChange(id); }}
                className={`w-full flex items-center justify-between px-2.5 py-2 rounded-lg text-sm transition-all duration-150 ${
                  isSelected ? 'bg-accent-surface font-medium text-primary shadow-washi-sm' : 'text-primary hover:bg-surface/60 opacity-90'
                }`}
              >
                <span className="flex items-center gap-2.5">
                  <Icon className={`w-4 h-4 stroke-[1.8] ${iconColor || 'text-muted'}`} />
                  <span>{label}</span>
                </span>
                {count > 0 && (
                  <Badge variant={badgeVariant} size="sm" className="font-mono text-[11px] px-1.5 py-0">
                    {count}
                  </Badge>
                )}
              </button>
            );
          })}
        </div>

        {/* Projects Section */}
        <div className="space-y-1 pt-1">
          <div className="flex items-center justify-between px-2 pb-1">
            <span className="text-[10px] font-mono uppercase tracking-wider text-muted">
              Проекты
            </span>
            <button
              type="button"
              onClick={() => setIsCreatingProject((prev) => !prev)}
              className="text-muted hover:text-primary p-0.5 rounded transition-colors"
              title="Новый проект"
              aria-label="Создать новый проект"
            >
              {isCreatingProject ? <X className="w-3 h-3" /> : <Plus className="w-3.5 h-3.5" />}
            </button>
          </div>

          {/* Inline Project Creation */}
          {isCreatingProject && (
            <form onSubmit={handleCreateProject} className="px-1 py-1 space-y-1.5">
              <Input
                autoFocus
                value={newProjectName}
                onChange={(e) => setNewProjectName(e.target.value)}
                placeholder="Имя проекта..."
                className="text-xs bg-surface border-subtle"
              />
              <div className="flex justify-end gap-1">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => setIsCreatingProject(false)}
                  className="text-[10px] h-6 px-1.5"
                >
                  Отмена
                </Button>
                <Button
                  type="submit"
                  variant="solid"
                  size="sm"
                  disabled={!newProjectName.trim()}
                  isLoading={createProjectMutation.isPending}
                  className="text-[10px] h-6 px-2"
                >
                  Создать
                </Button>
              </div>
            </form>
          )}

          {/* Project List */}
          <div className="space-y-0.5">
            {!isCreatingProject && projects.length === 0 && (
              <div className="px-2 py-1.5 text-[11px] text-muted italic">
                Нет проектов
              </div>
            )}
            {projects.map((proj) => {
              const isSelected = currentNav === `project:${proj.id}` && !showUiCatalog;
              return (
                <button
                  key={proj.id}
                  type="button"
                  onClick={() => handleSelectProject(proj.id)}
                  className={`w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-xs transition-all duration-150 ${
                    isSelected
                      ? 'bg-accent-surface font-medium text-primary shadow-washi-sm'
                      : 'text-primary hover:bg-surface/60'
                  }`}
                >
                  <span className="flex items-center gap-2 truncate">
                    <span className="w-2 h-2 rounded-full shrink-0 bg-accent-solid" />
                    <span className="truncate">{proj.name}</span>
                  </span>
                  <FolderClosed className="w-3 h-3 text-muted shrink-0" />
                </button>
              );
            })}
          </div>
        </div>

        {/* MiniCalendar section */}
        <div className="pt-2 border-t border-subtle/50">
          <MiniCalendar workspaceId={workspace?.id} isCompact />
        </div>

        {/* Settings & Extras */}
        <div className="pt-2 border-t border-subtle/50 space-y-0.5">
          <button
            type="button"
            onClick={() => setPickerOpen(true)}
            className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-xs text-muted hover:text-primary hover:bg-surface/60 transition-all group"
            title="Выбрать тему оформления"
          >
            <span className="flex items-center gap-2">
              <Palette className="w-3.5 h-3.5 text-muted group-hover:text-primary transition-colors" />
              <span>Оформление</span>
            </span>
            <span className="text-[10px] font-mono text-muted/70 group-hover:text-muted truncate max-w-[80px]">
              {themeInfo.name.split(' ')[0]}
            </span>
          </button>

          <button
            type="button"
            onClick={onToggleUiCatalog}
            className={`w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-xs transition-all duration-150 ${
              showUiCatalog ? 'bg-accent-surface font-medium text-primary' : 'text-muted hover:text-primary hover:bg-surface/60'
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
