import React, { useState, useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useAuth, AuthModal } from '@/features/auth';
import {
  useTaskNavigation,
  useInboxTasks,
  useCreateTaskMutation,
  useToggleTaskMutation,
  useDeleteTaskMutation,
  useTaskStore,
  TASKS_QUERY_KEY,
} from '@/features/task-navigation';
import { PanelLeftOpen } from 'lucide-react';
import { QuickCaptureModal } from '@/features/quick-capture';
import { Sidebar } from '@/widgets/sidebar';
import { BoardsRail } from '@/widgets/boards-panel';
import { MainSheet } from '@/widgets/main-sheet';
import { TaskDrawer } from '@/widgets/task-drawer';
import { UiCatalog } from './components/UiCatalog';
import { StatesShowcase } from './components/StatesShowcase';
import { Button } from '@/shared/components/ui';

import { useTodayTasksQuery } from '@/entities/task';
import { useProjectTasksQuery } from '@/entities/project';
import { AutomationsModal } from '@/features/automations-toggle';
import { AcceptInviteModal } from '@/features/workspace-members';
import { CommandPaletteModal } from '@/features/command-palette';

export const App: React.FC = () => {
  const queryClient = useQueryClient();
  const { user, workspace, setWorkspace, login, register, logout, refresh } = useAuth();

  const {
    selectedTaskId,
    setSelectedTaskId,
    isQuickCaptureOpen,
    setQuickCaptureOpen,
    activeNav,
    setActiveNav,
    selectedProjectId,
    selectedDateFilter,
    activeBoardId,
    setActiveBoardId,
    isBoardsPanelCollapsed,
    toggleBoardsPanel,
  } = useTaskStore();

  const {
    data: tasks = [],
    isLoading: isLoadingTasks,
    error: tasksError,
    refetch: refetchTasks,
  } = useInboxTasks();

  const { data: todayTasks = [] } = useTodayTasksQuery(workspace?.id);
  const { data: projectTasks = [] } = useProjectTasksQuery(
    workspace?.id,
    selectedProjectId ?? undefined
  );

  const createTaskMutation = useCreateTaskMutation();
  const toggleTaskMutation = useToggleTaskMutation();
  const deleteTaskMutation = useDeleteTaskMutation();

  const [showUiCatalog, setShowUiCatalog] = useState<boolean>(false);
  const [catalogTab, setCatalogTab] = useState<'ui-kit' | 'states'>('ui-kit');
  const [isAuthModalOpen, setIsAuthModalOpen] = useState<boolean>(false);
  const [isAutomationsModalOpen, setIsAutomationsModalOpen] = useState<boolean>(false);
  const [isCommandPaletteOpen, setIsCommandPaletteOpen] = useState<boolean>(false);

  const [isSidebarOpen, setIsSidebarOpen] = useState<boolean>(() => {
    const saved = localStorage.getItem('takt_sidebar_open');
    if (saved !== null) return saved === 'true';
    return typeof window !== 'undefined' ? window.innerWidth >= 1024 : true;
  });

  const toggleSidebar = () => {
    setIsSidebarOpen((prev) => {
      const next = !prev;
      localStorage.setItem('takt_sidebar_open', String(next));
      return next;
    });
  };

  // Global hotkeys: Cmd+K (Palette), '[' (Sidebar), 'c' (Quick Capture)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && (e.key === 'k' || e.key === 'K')) {
        e.preventDefault();
        setIsCommandPaletteOpen((prev) => !prev);
        return;
      }

      const isInput =
        e.target instanceof HTMLElement &&
        (e.target.tagName === 'INPUT' ||
          e.target.tagName === 'TEXTAREA' ||
          e.target.isContentEditable);

      if (
        isInput ||
        isQuickCaptureOpen ||
        isAuthModalOpen ||
        isAutomationsModalOpen ||
        isCommandPaletteOpen ||
        showUiCatalog ||
        e.metaKey ||
        e.ctrlKey ||
        e.altKey
      ) {
        return;
      }

      if (e.key === '[') {
        e.preventDefault();
        toggleSidebar();
      } else if (e.key === 'c' || e.key === 'C') {
        e.preventDefault();
        setQuickCaptureOpen(true);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [
    isQuickCaptureOpen,
    isAuthModalOpen,
    isAutomationsModalOpen,
    isCommandPaletteOpen,
    showUiCatalog,
    setQuickCaptureOpen,
  ]);

  // Invalidate tasks query when user auth state changes
  useEffect(() => {
    queryClient.invalidateQueries({ queryKey: TASKS_QUERY_KEY });
  }, [user, queryClient]);

  const isProjectNav = activeNav.startsWith('project:') || Boolean(selectedProjectId);
  const activeProjectId =
    selectedProjectId ||
    (activeNav.startsWith('project:') ? activeNav.replace('project:', '') : null);

  let currentNavTasks =
    activeNav === 'today'
      ? todayTasks
      : isProjectNav
      ? projectTasks
      : tasks;

  if (selectedDateFilter) {
    currentNavTasks = currentNavTasks.filter((t) => t.dueDate === selectedDateFilter);
  }

  // J/K Task Navigation
  const { selectedIndex, setSelectedIndex } = useTaskNavigation({
    tasks: currentNavTasks,
    onToggleTask: (id) => toggleTaskMutation.mutate(id),
    isEnabled: !isQuickCaptureOpen && !isAuthModalOpen && !isAutomationsModalOpen && !showUiCatalog,
  });

  const allAvailableTasks = [...tasks, ...todayTasks, ...projectTasks];
  const selectedTask = selectedTaskId
    ? allAvailableTasks.find((t) => t.id === selectedTaskId) ?? null
    : null;

  return (
    <div className="h-screen w-screen overflow-hidden bg-canvas text-primary flex">
      {/* Sidebar with collapse support */}
      <Sidebar
        user={user}
        workspace={workspace}
        inboxCount={tasks.filter((t) => !t.isCompleted).length}
        activeNav={activeNav}
        onNavChange={setActiveNav}
        onOpenQuickCapture={() => setQuickCaptureOpen(true)}
        onOpenCommandPalette={() => setIsCommandPaletteOpen(true)}
        isCollapsed={!isSidebarOpen}
        onToggleCollapse={toggleSidebar}
        onOpenAuth={() => setIsAuthModalOpen(true)}
        onLogout={logout}
        showUiCatalog={showUiCatalog}
        onToggleUiCatalog={() => setShowUiCatalog((prev) => !prev)}
        onOpenAutomations={() => setIsAutomationsModalOpen(true)}
        onWorkspaceChange={setWorkspace}
      />

      {/* Floating button when sidebar is collapsed */}
      {!isSidebarOpen && (
        <button
          type="button"
          onClick={toggleSidebar}
          className="fixed top-4 left-4 z-30 p-2 rounded-xl bg-surface/90 hover:bg-surface border border-subtle shadow-washi text-muted hover:text-primary transition-all duration-200 group flex items-center gap-1.5"
          title="Развернуть сайдбар ([)"
        >
          <PanelLeftOpen className="w-4 h-4 text-muted group-hover:text-primary transition-colors" />
          <span className="text-[10px] font-mono text-muted group-hover:text-primary hidden sm:inline">
            [
          </span>
        </button>
      )}

      {/* Boards Sub-Navigation Rail (WEEEK-style, 220px) */}
      {!showUiCatalog && isProjectNav && activeProjectId && (
        <BoardsRail
          workspaceId={workspace?.id}
          projectId={activeProjectId}
          activeBoardId={activeBoardId}
          onSelectBoard={setActiveBoardId}
          isCollapsed={isBoardsPanelCollapsed}
          onToggleCollapse={toggleBoardsPanel}
        />
      )}

      {/* Main Content Area */}
      {showUiCatalog ? (
        <main className="flex-1 h-full overflow-y-auto p-6 md:p-8 flex flex-col items-center">
          <div className="w-full max-w-4xl bg-surface border border-subtle rounded-sheet shadow-washi p-6 md:p-10 space-y-6">
            <div className="flex items-center justify-between border-b border-subtle/60 pb-4">
              <div>
                <h2 className="font-serif text-2xl font-normal text-primary">
                  Витрина компонентов и состояний
                </h2>
                <p className="text-xs text-muted">
                  Атомарный дизайн-паспорт «Soft Editorial GTD»
                </p>
              </div>
              <div className="flex items-center gap-2">
                <Button
                  size="sm"
                  variant={catalogTab === 'ui-kit' ? 'solid' : 'ghost'}
                  onClick={() => setCatalogTab('ui-kit')}
                >
                  Компоненты UI
                </Button>
                <Button
                  size="sm"
                  variant={catalogTab === 'states' ? 'solid' : 'ghost'}
                  onClick={() => setCatalogTab('states')}
                >
                  4 Состояния экрана
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => setShowUiCatalog(false)}
                >
                  Вернуться к приложению
                </Button>
              </div>
            </div>

            {catalogTab === 'ui-kit' ? (
              <UiCatalog onOpenModal={() => setQuickCaptureOpen(true)} />
            ) : (
              <StatesShowcase />
            )}
          </div>
        </main>
      ) : (
        <main className="flex-1 h-full flex overflow-hidden w-full max-w-full px-2 md:px-4">
          {/* Center Washi Sheet (100% available canvas) */}
          <MainSheet
            tasks={tasks}
            isLoading={isLoadingTasks}
            error={tasksError ? tasksError.message : null}
            onRetry={() => refetchTasks()}
            selectedIndex={selectedIndex}
            onSelectTask={setSelectedIndex}
            onToggleTask={(id) => toggleTaskMutation.mutate(id)}
            onDeleteTask={(id) => deleteTaskMutation.mutate(id)}
            onCreateTask={async (req) => {
              await createTaskMutation.mutateAsync(req);
            }}
            onOpenQuickCapture={() => setQuickCaptureOpen(true)}
            workspaceId={workspace?.id}
          />
        </main>
      )}

      {/* Slide-out Task Drawer (560px, non-modal on desktop) */}
      <TaskDrawer
        task={selectedTask}
        workspaceId={workspace?.id}
        onClose={() => setSelectedTaskId(null)}
        onToggleTask={(id) => toggleTaskMutation.mutate(id)}
        onDeleteTask={(id) => deleteTaskMutation.mutate(id)}
      />

      {/* Global Modals */}
      <QuickCaptureModal
        isOpen={isQuickCaptureOpen}
        onClose={() => setQuickCaptureOpen(false)}
        onSubmit={async (req) => {
          await createTaskMutation.mutateAsync(req);
        }}
      />

      <AuthModal
        isOpen={isAuthModalOpen}
        onClose={() => setIsAuthModalOpen(false)}
        onLogin={login}
        onRegister={register}
      />

      <AutomationsModal
        isOpen={isAutomationsModalOpen}
        onClose={() => setIsAutomationsModalOpen(false)}
        workspaceId={workspace?.id}
      />

      <AcceptInviteModal
        onAccepted={() => refresh()}
      />

      <CommandPaletteModal
        isOpen={isCommandPaletteOpen}
        onClose={() => setIsCommandPaletteOpen(false)}
        workspaceId={workspace?.id}
        onOpenQuickCapture={() => setQuickCaptureOpen(true)}
      />
    </div>
  );
};

export default App;
