import React, { useState, useEffect, useCallback } from 'react';
import { Modal, Input, Button } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import type { Priority, CreateInboxTaskRequest } from '@/shared/api';
import { useTaskStore, useCreateTaskMutation } from '@/features/task-navigation';

export interface QuickCaptureModalProps {
  isOpen?: boolean;
  onClose?: () => void;
  onSubmit?: (task: CreateInboxTaskRequest) => Promise<void>;
  enableGlobalHotkey?: boolean;
  designatedColumnName?: string | null;
  projectName?: string | null;
}

const PRIORITIES: { value: Priority; label: string }[] = [
  { value: 'none', label: 'Обычный' },
  { value: 'low', label: 'Низкий' },
  { value: 'medium', label: 'Средний' },
  { value: 'high', label: 'Высокий' },
  { value: 'pending', label: 'В ожидании' },
];

const PRIORITY_ACTIVE_STYLES: Record<string, string> = {
  pending: 'bg-[#4A6B82]/15 text-[#4A6B82] border-[#4A6B82]/40',
  urgent: 'bg-[#4A6B82]/15 text-[#4A6B82] border-[#4A6B82]/40',
  high: 'bg-[#D97746]/15 text-[#D97746] border-[#D97746]/40',
  medium: 'bg-[#B8943F]/15 text-[#B8943F] border-[#B8943F]/40',
  low: 'bg-[#566B44]/15 text-[#566B44] border-[#566B44]/40',
  none: 'bg-surface text-primary border-subtle',
};

export const QuickCaptureModal: React.FC<QuickCaptureModalProps> = ({
  isOpen: propsIsOpen,
  onClose: propsOnClose,
  onSubmit: propsOnSubmit,
  designatedColumnName,
  projectName,
}) => {
  const { isQuickCaptureOpen, setQuickCaptureOpen } = useTaskStore();
  const createTaskMutation = useCreateTaskMutation();

  const isOpen = propsIsOpen !== undefined ? propsIsOpen : isQuickCaptureOpen;
  const handleClose = propsOnClose ?? (() => setQuickCaptureOpen(false));

  const [title, setTitle] = useState('');
  const [priority, setPriority] = useState<Priority>('none');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Reset form when modal opens
  useEffect(() => {
    if (isOpen) {
      setTitle('');
      setPriority('none');
    }
  }, [isOpen]);

  const handleSubmit = useCallback(async () => {
    const trimmedTitle = title.trim();
    if (!trimmedTitle || isSubmitting) return;

    setIsSubmitting(true);
    try {
      if (propsOnSubmit) {
        await propsOnSubmit({
          title: trimmedTitle,
          priority,
        });
      } else {
        await createTaskMutation.mutateAsync({
          title: trimmedTitle,
          priority,
        });
      }
      setTitle('');
      setPriority('none');
      handleClose();
    } finally {
      setIsSubmitting(false);
    }
  }, [title, priority, isSubmitting, propsOnSubmit, createTaskMutation, handleClose]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      handleSubmit();
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Быстрый захват мысли"
      description={
        designatedColumnName
          ? `Задача будет добавлена во входящую колонку «${designatedColumnName}» (${projectName || 'проект'}). ⌘+Enter для быстрого сохранения.`
          : 'Мгновенно сбросьте задачу в Inbox. Клавиши ⌘+Enter для быстрого сохранения.'
      }
      footer={
        <div className="w-full flex items-center justify-between">
          <div className="flex items-center gap-1.5 text-xs text-muted font-mono">
            <kbd className="px-1.5 py-0.5 rounded bg-canvas border border-subtle text-[10px]">
              ⌘ + Enter
            </kbd>
            <span>сохранить</span>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" onClick={handleClose} disabled={isSubmitting}>
              Отмена
            </Button>
            <Button
              variant="solid"
              size="sm"
              onClick={handleSubmit}
              isLoading={isSubmitting}
              disabled={!title.trim()}
            >
              {designatedColumnName ? `В «${designatedColumnName}»` : 'В Inbox'}
            </Button>
          </div>
        </div>
      }
    >
      <div className="space-y-4" onKeyDown={handleKeyDown}>
        <div>
          <label className="block text-xs uppercase tracking-wider text-muted font-mono mb-1.5">
            Заголовок задачи
          </label>
          <Input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Что нужно сделать?"
            autoFocus
          />
        </div>

        <div>
          <label className="block text-xs uppercase tracking-wider text-muted font-mono mb-1.5">
            Приоритет
          </label>
          <div className="flex flex-wrap items-center gap-1.5">
            {PRIORITIES.map((p) => {
              const isSelected = priority === p.value || (p.value === 'pending' && priority === 'urgent');
              return (
                <button
                  key={p.value}
                  type="button"
                  onClick={() => setPriority(p.value)}
                  className={cn(
                    'text-xs px-2.5 py-1 rounded-full border transition-all',
                    isSelected
                      ? `${PRIORITY_ACTIVE_STYLES[p.value]} font-medium shadow-washi-sm`
                      : 'border-subtle/60 text-muted hover:text-primary hover:bg-canvas/50'
                  )}
                >
                  {p.label}
                </button>
              );
            })}
          </div>
        </div>
      </div>
    </Modal>
  );
};
