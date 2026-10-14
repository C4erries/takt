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
import { QuickCaptureModal } from '@/features/quick-capture';
import { Sidebar } from '@/widgets/sidebar';
import { MainSheet } from '@/widgets/main-sheet';
import { TaskInspector } from '@/widgets/task-inspector';
import { MiniCalendar } from '@/widgets/mini-calendar';
import { UiCatalog } from './components/UiCatalog';
import { StatesShowcase } from './components/StatesShowcase';
import { Button } from '@/shared/components/ui';

import { useTodayTasksQuery } from '@/entities/task';
import { useProjectTasksQuery } from '@/entities/project';
import { AutomationsModal } from '@/features/automations-toggle';

export const App: React.FC = () => {
  const queryClient = useQueryClient();
  const { user, workspace, login, register, logout } = useAuth();

  const {
    selectedTaskId,
    isQuickCaptureOpen,
    setQuickCaptureOpen,
    activeNav,
    setActiveNav,
    selectedProjectId,
    selectedDateFilter,
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

  // Invalidate tasks query when user auth state changes
  useEffect(() => {
    queryClient.invalidateQueries({ queryKey: TASKS_QUERY_KEY });
  }, [user, queryClient]);

  // Global hotkey 'c' for Quick Capture
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (isQuickCaptureOpen || isAuthModalOpen || isAutomationsModalOpen || showUiCatalog) return;

      const target = e.target as HTMLElement | null;
      const isInput =
        target &&
        (target.tagName === 'INPUT' ||
          target.tagName === 'TEXTAREA' ||
          target.isContentEditable);

      if (isInput) return;
      if (e.metaKey || e.ctrlKey || e.altKey) return;

      if (e.key === 'c' || e.key === 'C') {
        e.preventDefault();
        setQuickCaptureOpen(true);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isQuickCaptureOpen, isAuthModalOpen, isAutomationsModalOpen, showUiCatalog, setQuickCaptureOpen]);

  let currentNavTasks =
    activeNav === 'today'
      ? todayTasks
      : selectedProjectId
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
  const selectedTask =
    (selectedTaskId ? allAvailableTasks.find((t) => t.id === selectedTaskId) : null) ??
    currentNavTasks[selectedIndex] ??
    null;

  return (
    <div className="h-screen w-screen overflow-hidden bg-canvas text-primary flex">
      {/* 3-Column Layout: Column 1 - Left Sidebar */}
      <Sidebar
        user={user}
        workspace={workspace}
        inboxCount={tasks.filter((t) => !t.isCompleted).length}
        activeNav={activeNav}
        onNavChange={setActiveNav}
        onOpenQuickCapture={() => setQuickCaptureOpen(true)}
        onOpenAuth={() => setIsAuthModalOpen(true)}
        onLogout={logout}
        showUiCatalog={showUiCatalog}
        onToggleUiCatalog={() => setShowUiCatalog((prev) => !prev)}
        onOpenAutomations={() => setIsAutomationsModalOpen(true)}
      />

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
        <main className="flex-1 h-full flex justify-center px-4 gap-4 overflow-hidden">
          {/* Column 2: Center Washi Sheet */}
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

          {/* Column 3: Right Bento Panel (Inspector + MiniCalendar) */}
          <div className="hidden lg:flex flex-col gap-3 py-3 h-full justify-start overflow-hidden">
            <TaskInspector
              task={selectedTask}
              workspaceId={workspace?.id}
              onToggleTask={(id) => toggleTaskMutation.mutate(id)}
              onDeleteTask={(id) => deleteTaskMutation.mutate(id)}
            />
            <MiniCalendar workspaceId={workspace?.id} />
          </div>
        </main>
      )}

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
    </div>
  );
};

export default App;
