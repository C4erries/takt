import type { Task } from '@/shared/api';
import type { Project, Board, Column } from './types';

export const DEMO_PROJECT_ID = 'demo-project-1';
export const DEMO_BOARD_ID = 'demo-board-1';
export const DEMO_BOARD_2_ID = 'demo-board-2';

export const DEMO_PROJECTS: Project[] = [
  {
    id: DEMO_PROJECT_ID,
    workspaceId: 'demo-workspace',
    name: 'Запуск Takt 0.3.1',
    description: 'Внедрение проектов и Канбан-доски с тактильным Drag-and-Drop',
    color: '#2D4F3E',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const DEMO_BOARDS: Board[] = [
  {
    id: DEMO_BOARD_ID,
    projectId: DEMO_PROJECT_ID,
    name: 'Основная',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: DEMO_BOARD_2_ID,
    projectId: DEMO_PROJECT_ID,
    name: 'Мульти-доски (Sprint 0.5.2)',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const DEMO_COLUMNS: Column[] = [
  {
    id: 'demo-col-todo',
    boardId: DEMO_BOARD_ID,
    name: 'К выполнению',
    position: 1000.0,
    statusCategory: 'todo',
    isInbox: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'demo-col-in-progress',
    boardId: DEMO_BOARD_ID,
    name: 'В работе',
    position: 2000.0,
    statusCategory: 'in_progress',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'demo-col-done',
    boardId: DEMO_BOARD_ID,
    name: 'Готово',
    position: 3000.0,
    statusCategory: 'done',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'demo-col-b2-todo',
    boardId: DEMO_BOARD_2_ID,
    name: 'Запланировано',
    position: 1000.0,
    statusCategory: 'todo',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'demo-col-b2-done',
    boardId: DEMO_BOARD_2_ID,
    name: 'Готово',
    position: 2000.0,
    statusCategory: 'done',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const DEMO_PROJECT_TASKS: Task[] = [
  {
    id: 'pt-1',
    workspaceId: 'demo-workspace',
    projectId: DEMO_PROJECT_ID,
    columnId: 'demo-col-todo',
    title: 'Настроить Fractional Indexing для колонок',
    descriptionMarkdown: 'Новая позиция вычисляется как (pos_A + pos_B) / 2',
    isCompleted: false,
    priority: 'high',
    position: 1000.0,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'pt-2',
    workspaceId: 'demo-workspace',
    projectId: DEMO_PROJECT_ID,
    columnId: 'demo-col-in-progress',
    title: 'Интегрировать @dnd-kit/core с тактильным Washi-дизайном',
    descriptionMarkdown: 'Ощущение японской канцелярии Midori MD при перетаскивании карточек',
    isCompleted: false,
    priority: 'pending',
    position: 1000.0,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'pt-3',
    workspaceId: 'demo-workspace',
    projectId: DEMO_PROJECT_ID,
    columnId: 'demo-col-done',
    title: 'Разработать спецификацию OpenAPI для проектов и досок',
    descriptionMarkdown: 'Контракты согласованы и зафиксированы в api/openapi/paths/projects.yaml',
    isCompleted: true,
    completedAt: new Date().toISOString(),
    priority: 'medium',
    position: 1000.0,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'pt-4',
    workspaceId: 'demo-workspace',
    projectId: DEMO_PROJECT_ID,
    columnId: 'demo-col-b2-todo',
    title: 'Реализовать WEEEK-style Boards Sub-Navigation Rail',
    descriptionMarkdown: 'Панель 220px с кнопкой сворачивания < / > и счетчиками',
    isCompleted: false,
    priority: 'high',
    position: 1000.0,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export function updateDemoTaskPosition(taskId: string, columnId: string, position: number): void {
  const task = DEMO_PROJECT_TASKS.find((t) => t.id === taskId);
  if (task) {
    task.columnId = columnId;
    task.position = position;
    task.updatedAt = new Date().toISOString();
  }
}

export function updateDemoColumnPosition(columnId: string, position: number): void {
  const col = DEMO_COLUMNS.find((c) => c.id === columnId);
  if (col) {
    col.position = position;
    col.updatedAt = new Date().toISOString();
  }
}
