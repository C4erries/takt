import React, { useState, useEffect, useCallback } from 'react';
import { useAuth, AuthModal } from '@/features/auth';
import { useTaskNavigation } from '@/features/task-navigation';
import { QuickCaptureModal } from '@/features/quick-capture';
import { Sidebar } from '@/widgets/sidebar';
import { MainSheet } from '@/widgets/main-sheet';
import { TaskInspector } from '@/widgets/task-inspector';
import { MiniCalendar } from '@/widgets/mini-calendar';
import { tasksApi } from '@/shared/api';
import type { Task, CreateInboxTaskRequest } from '@/shared/api';
import { UiCatalog } from './components/UiCatalog';
import { StatesShowcase } from './components/StatesShowcase';
import { Button } from '@/shared/components/ui';

const ONBOARDING_TASKS: Task[] = [
  {
    id: 'onboarding-1',
    workspaceId: 'demo-workspace',
    title: 'Нажми J и K для клавиатурного перемещения по строкам',
    descriptionMarkdown: 'Клавиши J (вниз) и K (вверх) позволяют мгновенно перемещаться между задачами со snappy-анимацией.',
    isCompleted: false,
    priority: 'high',
    position: 0,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'onboarding-2',
    workspaceId: 'demo-workspace',
    title: 'Нажми Пробел или кликни круглый чекбокс, чтобы завершить задачу',
    descriptionMarkdown: 'Клавиша Space или X переключает статус задачи.',
    isCompleted: false,
    priority: 'medium',
    position: 1,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'onboarding-3',
    workspaceId: 'demo-workspace',
    title: 'Нажми C в любом месте экрана для мгновенного сброса мысли (Quick Capture)',
    descriptionMarkdown: 'Горячая клавиша C открывает всплывающее окно быстрого захвата.',
    isCompleted: false,
    priority: 'urgent',
    position: 2,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const App: React.FC = () => {
  const { user, workspace, login, register, logout } = useAuth();

  const [tasks, setTasks] = useState<Task[]>([]);
  const [isLoadingTasks, setIsLoadingTasks] = useState<boolean>(true);
  const [tasksError, setTasksError] = useState<string | null>(null);

  const [activeNav, setActiveNav] = useState<'inbox' | 'today' | 'projects'>('inbox');
  const [showUiCatalog, setShowUiCatalog] = useState<boolean>(false);
  const [catalogTab, setCatalogTab] = useState<'ui-kit' | 'states'>('ui-kit');
  const [isQuickCaptureOpen, setIsQuickCaptureOpen] = useState<boolean>(false);
  const [isAuthModalOpen, setIsAuthModalOpen] = useState<boolean>(false);

  // Load tasks from API or fallback to onboarding
  const fetchTasks = useCallback(async () => {
    setIsLoadingTasks(true);
    setTasksError(null);
    try {
      const data = await tasksApi.getInbox();
      setTasks(data);
    } catch {
      // If unauthorized or server offline, fall back to onboarding tasks
      setTasks(ONBOARDING_TASKS);
    } finally {
      setIsLoadingTasks(false);
    }
  }, []);

  useEffect(() => {
    fetchTasks();
  }, [user, fetchTasks]);

  // Task mutation handlers
  const handleToggleTask = useCallback(
    async (taskId: string) => {
      // Optimistic local update
      setTasks((prev) =>
        prev.map((t) => (t.id === taskId ? { ...t, isCompleted: !t.isCompleted } : t))
      );

      try {
        await tasksApi.toggle(taskId);
      } catch {
        // If demo/offline mode, keep optimistic state
      }
    },
    []
  );

  const handleDeleteTask = useCallback(async (taskId: string) => {
    setTasks((prev) => prev.filter((t) => t.id !== taskId));
    try {
      await tasksApi.delete(taskId);
    } catch {
      // Ignored for demo tasks
    }
  }, []);

  const handleCreateTask = useCallback(
    async (req: CreateInboxTaskRequest) => {
      try {
        const created = await tasksApi.createInbox(req);
        setTasks((prev) => [created, ...prev]);
      } catch {
        // Fallback for offline/guest
        const demoTask: Task = {
          id: `task-${Date.now()}`,
          workspaceId: workspace?.id || 'demo-workspace',
          title: req.title,
          descriptionMarkdown: req.descriptionMarkdown || '',
          isCompleted: false,
          priority: req.priority || 'none',
          position: 0,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        };
        setTasks((prev) => [demoTask, ...prev]);
      }
    },
    [workspace]
  );

  // Global hotkey 'c' for Quick Capture
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (isQuickCaptureOpen || isAuthModalOpen || showUiCatalog) return;

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
        setIsQuickCaptureOpen(true);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isQuickCaptureOpen, isAuthModalOpen, showUiCatalog]);

  // J/K Task Navigation
  const { selectedIndex, setSelectedIndex } = useTaskNavigation({
    tasks,
    onToggleTask: handleToggleTask,
    isEnabled: !isQuickCaptureOpen && !isAuthModalOpen && !showUiCatalog,
  });

  const selectedTask = tasks[selectedIndex] ?? null;

  return (
    <div className="h-screen w-screen overflow-hidden bg-canvas text-primary flex">
      {/* 3-Column Layout: Column 1 - Left Sidebar */}
      <Sidebar
        user={user}
        workspace={workspace}
        inboxCount={tasks.filter((t) => !t.isCompleted).length}
        activeNav={activeNav}
        onNavChange={setActiveNav}
        onOpenQuickCapture={() => setIsQuickCaptureOpen(true)}
        onOpenAuth={() => setIsAuthModalOpen(true)}
        onLogout={logout}
        showUiCatalog={showUiCatalog}
        onToggleUiCatalog={() => setShowUiCatalog((prev) => !prev)}
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
              <UiCatalog onOpenModal={() => setIsQuickCaptureOpen(true)} />
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
            error={tasksError}
            onRetry={fetchTasks}
            selectedIndex={selectedIndex}
            onSelectTask={setSelectedIndex}
            onToggleTask={handleToggleTask}
            onDeleteTask={handleDeleteTask}
            onCreateTask={handleCreateTask}
            onOpenQuickCapture={() => setIsQuickCaptureOpen(true)}
          />

          {/* Column 3: Right Bento Panel (Inspector + MiniCalendar) */}
          <div className="hidden lg:flex flex-col gap-3 py-3 h-full justify-start overflow-hidden">
            <TaskInspector
              task={selectedTask}
              onToggleTask={handleToggleTask}
              onDeleteTask={handleDeleteTask}
            />
            <MiniCalendar />
          </div>
        </main>
      )}

      {/* Global Modals */}
      <QuickCaptureModal
        isOpen={isQuickCaptureOpen}
        onClose={() => setIsQuickCaptureOpen(false)}
        onSubmit={handleCreateTask}
      />

      <AuthModal
        isOpen={isAuthModalOpen}
        onClose={() => setIsAuthModalOpen(false)}
        onLogin={login}
        onRegister={register}
      />
    </div>
  );
};

export default App;
