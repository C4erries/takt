import React, { useState } from 'react';
import {
  Inbox,
  Zap,
  FolderClosed,
  Plus,
  LogOut,
  LogIn,
  SlidersHorizontal,
  Layers,
  X,
} from 'lucide-react';
import { Button, Badge, Input } from '@/shared/components/ui';
import type { User, Workspace } from '@/shared/api';
import { useTaskStore, useInboxTasks } from '@/features/task-navigation';
import { useProjectsQuery, useCreateProjectMutation } from '@/entities/project';

export interface SidebarProps {
  user: User | null;
  workspace: Workspace | null;
  inboxCount?: number;
  activeNav?: string;
  onNavChange?: (nav: string) => void;
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
  const {
    setQuickCaptureOpen,
    activeNav: storeActiveNav,
    setActiveNav: storeSetActiveNav,
    setSelectedProjectId,
  } = useTaskStore();
  const { data: inboxTasks } = useInboxTasks();
  const { data: projects = [] } = useProjectsQuery(workspace?.id);
  const createProjectMutation = useCreateProjectMutation(workspace?.id);

  const [isCreatingProject, setIsCreatingProject] = useState(false);
  const [newProjectName, setNewProjectName] = useState('');

  const currentInboxCount =
    inboxCount !== undefined
      ? inboxCount
      : (inboxTasks?.filter((t) => !t.isCompleted).length ?? 0);
  const currentNav = activeNav ?? storeActiveNav;
  const handleNavChange = onNavChange ?? storeSetActiveNav;
  const handleOpenQuickCapture = onOpenQuickCapture ?? (() => setQuickCaptureOpen(true));

  const handleSelectProject = (projectId: string) => {
    setSelectedProjectId(projectId);
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
    <aside className="w-60 shrink-0 h-full flex flex-col justify-between py-6 px-4 select-none">
      {/* Top section: Workspace, Quick Action, Navigation */}
      <div className="space-y-6 overflow-y-auto pr-0.5">
        {/* Workspace Brand */}
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

        {/* Horizons Navigation */}
        <div className="space-y-1">
          <div className="px-2 pb-1 text-[10px] font-mono uppercase tracking-wider text-muted">
            Горизонты
          </div>

          {/* Inbox */}
          <button
            type="button"
            onClick={() => {
              setSelectedProjectId(null);
              handleNavChange('inbox');
            }}
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
            onClick={() => {
              setSelectedProjectId(null);
              handleNavChange('today');
            }}
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
