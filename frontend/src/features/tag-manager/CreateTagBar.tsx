import React, { useState } from 'react';
import { Plus, X } from 'lucide-react';
import { Button, Input } from '@/shared/components/ui';
import { TAG_PALETTE, DEFAULT_TAG_COLOR } from '@/entities/task';
import type { Project } from '@/shared/api';

export interface CreateTagBarProps {
  projects: Project[];
  defaultProjectId?: string | null;
  isPending: boolean;
  onCreateTag: (data: { name: string; colorHex: string; projectId?: string | null }) => Promise<void>;
}

export const CreateTagBar: React.FC<CreateTagBarProps> = ({
  projects,
  defaultProjectId,
  isPending,
  onCreateTag,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [name, setName] = useState('');
  const [colorHex, setColorHex] = useState(DEFAULT_TAG_COLOR);
  const [targetScope, setTargetScope] = useState<'global' | string>(defaultProjectId || 'global');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = name.trim().replace(/^#/, '');
    if (!trimmed || isPending) return;

    await onCreateTag({
      name: trimmed,
      colorHex,
      projectId: targetScope === 'global' ? null : targetScope,
    });
    setName('');
    setColorHex(DEFAULT_TAG_COLOR);
    setIsOpen(false);
  };

  if (!isOpen) {
    return (
      <Button
        type="button"
        variant="solid"
        size="sm"
        onClick={() => {
          setTargetScope(defaultProjectId || 'global');
          setIsOpen(true);
        }}
        className="text-xs h-8 shrink-0"
        leftIcon={<Plus className="w-3.5 h-3.5" />}
      >
        Создать тег
      </Button>
    );
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="p-3 rounded-xl border border-accent-solid/50 bg-surface shadow-washi-sm space-y-2.5 w-full max-w-xl animate-in fade-in zoom-in-95 duration-100"
    >
      <div className="flex items-center justify-between">
        <span className="text-xs font-semibold text-primary">Новый тег</span>
        <button
          type="button"
          onClick={() => setIsOpen(false)}
          className="text-muted hover:text-primary p-0.5 rounded transition-colors"
        >
          <X className="w-3.5 h-3.5" />
        </button>
      </div>

      <div className="flex items-center gap-2 flex-wrap sm:flex-nowrap">
        <Input
          autoFocus
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Название метки (напр. срочно)..."
          className="text-xs bg-canvas border-subtle h-8 flex-1"
        />

        {/* Scope selector */}
        <select
          value={targetScope}
          onChange={(e) => setTargetScope(e.target.value)}
          className="text-xs bg-canvas border border-subtle rounded-lg px-2 py-1.5 text-primary focus:outline-none focus:border-accent-solid h-8"
        >
          <option value="global">🌐 Глобальный</option>
          {projects.map((p) => (
            <option key={p.id} value={p.id}>
              📁 {p.name}
            </option>
          ))}
        </select>
      </div>

      {/* Color Palette */}
      <div className="flex items-center justify-between gap-2 flex-wrap pt-1">
        <div className="flex items-center gap-1.5 flex-wrap">
          {TAG_PALETTE.map((c) => {
            const isSelected = colorHex.toLowerCase() === c.hex.toLowerCase();
            return (
              <button
                key={c.hex}
                type="button"
                onClick={() => setColorHex(c.hex)}
                title={c.name}
                className={`w-4 h-4 rounded-full transition-transform hover:scale-125 focus:outline-none flex items-center justify-center shrink-0 ${
                  isSelected ? 'ring-2 ring-primary ring-offset-1 ring-offset-surface scale-110' : 'opacity-80 hover:opacity-100'
                }`}
                style={{ backgroundColor: c.hex }}
              />
            );
          })}
        </div>

        <div className="flex items-center gap-1.5 ml-auto">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => setIsOpen(false)}
            className="text-xs h-7 px-2"
          >
            Отмена
          </Button>
          <Button
            type="submit"
            variant="solid"
            size="sm"
            disabled={!name.trim()}
            isLoading={isPending}
            className="text-xs h-7 px-3"
          >
            Создать
          </Button>
        </div>
      </div>
    </form>
  );
};
