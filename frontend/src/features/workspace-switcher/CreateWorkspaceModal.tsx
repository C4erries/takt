import React, { useState } from 'react';
import { Modal, Button, Input, ErrorBanner } from '@/shared/components/ui';
import type { Workspace } from '@/entities/workspace';
import {
  useCreateWorkspaceMutation,
  useSwitchWorkspaceMutation,
} from '@/entities/workspace';

export interface CreateWorkspaceModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreated?: (workspace: Workspace) => void;
}

export const CreateWorkspaceModal: React.FC<CreateWorkspaceModalProps> = ({
  isOpen,
  onClose,
  onCreated,
}) => {
  const [name, setName] = useState('');
  const [error, setError] = useState<string | null>(null);

  const createMutation = useCreateWorkspaceMutation();
  const switchMutation = useSwitchWorkspaceMutation();

  const isPending = createMutation.isPending || switchMutation.isPending;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed || isPending) return;

    setError(null);
    try {
      const created = await createMutation.mutateAsync({ name: trimmed });
      await switchMutation.mutateAsync(created.id);
      setName('');
      onClose();
      onCreated?.(created);
    } catch {
      setError('Не удалось создать воркспейс. Попробуйте еще раз.');
    }
  };

  const handleClose = () => {
    setName('');
    setError(null);
    onClose();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Новый воркспейс"
      description="Создайте пространство для ваших задач, проектов и команды"
      className="max-w-md"
      footer={
        <div className="flex items-center justify-end gap-2.5 w-full">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleClose}
            disabled={isPending}
          >
            Отмена
          </Button>
          <Button
            type="submit"
            form="create-workspace-form"
            variant="solid"
            size="sm"
            disabled={!name.trim() || isPending}
          >
            {isPending ? 'Создание...' : 'Создать'}
          </Button>
        </div>
      }
    >
      <form id="create-workspace-form" onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <ErrorBanner
            message={error}
          />
        )}

        <div className="space-y-1.5">
          <label className="block text-xs font-medium text-muted">
            Название воркспейса
          </label>
          <Input
            placeholder="Например, Работа, Продукт или Личное"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus
            required
            disabled={isPending}
          />
        </div>
      </form>
    </Modal>
  );
};
