import React, { useState, useEffect, useCallback } from 'react';
import { Modal, Input, Button, Badge } from '@/shared/components/ui';
import type { Priority, CreateInboxTaskRequest } from '@/shared/api';
import { useTaskStore, useCreateTaskMutation } from '@/features/task-navigation';

export interface QuickCaptureModalProps {
  isOpen?: boolean;
  onClose?: () => void;
  onSubmit?: (task: CreateInboxTaskRequest) => Promise<void>;
  enableGlobalHotkey?: boolean;
}

const PRIORITIES: { value: Priority; label: string; badgeVariant: 'subtle' | 'straw' | 'denim' | 'persimmon' }[] = [
  { value: 'none', label: 'Обычный', badgeVariant: 'subtle' },
  { value: 'low', label: 'Низкий', badgeVariant: 'subtle' },
  { value: 'medium', label: 'Средний', badgeVariant: 'straw' },
  { value: 'high', label: 'Высокий', badgeVariant: 'denim' },
  { value: 'urgent', label: 'Срочный', badgeVariant: 'persimmon' },
];

export const QuickCaptureModal: React.FC<QuickCaptureModalProps> = ({
  isOpen: propsIsOpen,
  onClose: propsOnClose,
  onSubmit: propsOnSubmit,
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
      description="Мгновенно сбросьте задачу в Inbox. Клавиши ⌘+Enter для быстрого сохранения."
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
              В Inbox
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
          <div className="flex flex-wrap items-center gap-2">
            {PRIORITIES.map((p) => {
              const isSelected = priority === p.value;
              return (
                <button
                  key={p.value}
                  type="button"
                  onClick={() => setPriority(p.value)}
                  className={`rounded-full transition-transform ${isSelected ? 'scale-105 ring-2 ring-accent-solid ring-offset-1' : 'opacity-70 hover:opacity-100'}`}
                >
                  <Badge variant={p.badgeVariant} size="sm">
                    {p.label}
                  </Badge>
                </button>
              );
            })}
          </div>
        </div>
      </div>
    </Modal>
  );
};
