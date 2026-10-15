import React, { useState, useRef, useEffect } from 'react';
import { ChevronDown, Check, Plus, Users } from 'lucide-react';
import { Badge, Skeleton } from '@/shared/components/ui';
import type { User } from '@/shared/api';
import type { Workspace } from '@/entities/workspace';
import {
  useWorkspacesQuery,
  useSwitchWorkspaceMutation,
  ACTIVE_WORKSPACE_STORAGE_KEY,
} from '@/entities/workspace';
import { CreateWorkspaceModal } from './CreateWorkspaceModal';
import { WorkspaceMembersModal } from '@/features/workspace-members';

export interface WorkspaceSwitcherProps {
  currentWorkspace: Workspace | null;
  currentUser: User | null;
  onWorkspaceChange?: (workspace: Workspace) => void;
}

export const WorkspaceSwitcher: React.FC<WorkspaceSwitcherProps> = ({
  currentWorkspace,
  currentUser,
  onWorkspaceChange,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isMembersOpen, setIsMembersOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  const { data: workspaces = [], isLoading } = useWorkspacesQuery();
  const switchMutation = useSwitchWorkspaceMutation();

  // Find active workspace: currentWorkspace prop, or saved in storage, or first in list
  const activeWorkspaceId =
    currentWorkspace?.id ||
    localStorage.getItem(ACTIVE_WORKSPACE_STORAGE_KEY) ||
    workspaces[0]?.id;

  const activeWorkspace =
    workspaces.find((w) => w.id === activeWorkspaceId) ||
    currentWorkspace ||
    workspaces[0] ||
    null;

  // Close dropdown on outside click or Escape key
  useEffect(() => {
    if (!isOpen) return;

    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    window.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen]);

  const handleSwitch = async (workspace: Workspace) => {
    setIsOpen(false);
    if (workspace.id === activeWorkspace?.id) return;

    try {
      const res = await switchMutation.mutateAsync(workspace.id);
      onWorkspaceChange?.(res.workspace);
    } catch {
      // Fallback
      onWorkspaceChange?.(workspace);
    }
  };

  const displayName = activeWorkspace?.name || 'Личное';
  const roleLabel =
    activeWorkspace?.role === 'owner'
      ? 'Владелец'
      : activeWorkspace?.role === 'member'
      ? 'Участник'
      : 'Workspace';

  return (
    <>
      <div className="relative" ref={menuRef}>
        {/* Trigger Button */}
        <button
          type="button"
          onClick={() => setIsOpen((prev) => !prev)}
          className="w-full flex items-center justify-between p-1.5 -ml-1 rounded-xl transition-colors hover:bg-canvas/80 text-left group focus:outline-none focus-visible:ring-2 focus-visible:ring-accent-solid"
          aria-expanded={isOpen}
          aria-haspopup="true"
        >
          <div className="flex items-center gap-2.5 min-w-0">
            <div className="w-7 h-7 rounded-lg bg-accent-solid text-surface flex items-center justify-center font-serif text-base font-semibold shadow-washi-sm shrink-0">
              {displayName.charAt(0).toUpperCase()}
            </div>
            <div className="flex flex-col min-w-0">
              <span className="font-serif text-base font-normal tracking-tight text-primary leading-tight truncate max-w-[130px]">
                {displayName}
              </span>
              <span className="text-[10px] font-mono uppercase tracking-wider text-muted">
                {roleLabel}
              </span>
            </div>
          </div>
          <ChevronDown
            className={`w-4 h-4 text-muted transition-transform duration-150 shrink-0 ${
              isOpen ? 'rotate-180 text-primary' : 'group-hover:text-primary'
            }`}
          />
        </button>

        {/* Dropdown Menu */}
        {isOpen && (
          <div
            className="absolute top-full left-0 mt-2 w-64 bg-surface border border-subtle rounded-xl shadow-washi-modal z-50 p-1.5 animate-in fade-in-0 zoom-in-95 duration-150 ease-spring-jk"
            role="menu"
          >
            {/* Header */}
            <div className="text-[10px] font-mono uppercase tracking-wider text-muted px-2.5 py-1.5">
              Воркспейсы
            </div>

            {/* Workspaces List */}
            <div className="space-y-0.5 max-h-56 overflow-y-auto">
              {isLoading && (
                <div className="space-y-1 p-1">
                  <Skeleton className="h-8 w-full rounded-lg" />
                  <Skeleton className="h-8 w-full rounded-lg" />
                </div>
              )}

              {!isLoading &&
                workspaces.map((ws) => {
                  const isActive = ws.id === activeWorkspace?.id;
                  const isWsOwner = ws.role === 'owner';

                  return (
                    <button
                      key={ws.id}
                      type="button"
                      onClick={() => handleSwitch(ws)}
                      className={`w-full flex items-center justify-between px-2.5 py-2 rounded-lg text-left text-xs transition-colors ${
                        isActive
                          ? 'bg-accent-surface/40 text-primary font-medium'
                          : 'hover:bg-canvas text-primary'
                      }`}
                      role="menuitem"
                    >
                      <div className="flex items-center gap-2 min-w-0">
                        <div className="w-5 h-5 rounded-md bg-canvas border border-subtle flex items-center justify-center font-serif text-xs shrink-0">
                          {ws.name.charAt(0).toUpperCase()}
                        </div>
                        <span className="truncate max-w-[110px]">{ws.name}</span>
                        <Badge
                          variant={isWsOwner ? 'matcha' : 'subtle'}
                          size="sm"
                        >
                          {isWsOwner ? 'Владелец' : 'Участник'}
                        </Badge>
                      </div>

                      {isActive && (
                        <Check className="w-4 h-4 text-accent-solid shrink-0" />
                      )}
                    </button>
                  );
                })}
            </div>

            {/* Tatami Divider */}
            <div className="my-1.5 border-t border-subtle" />

            {/* Action Buttons */}
            <div className="space-y-0.5">
              <button
                type="button"
                onClick={() => {
                  setIsOpen(false);
                  setIsCreateOpen(true);
                }}
                className="w-full flex items-center gap-2 px-2.5 py-2 rounded-lg text-left text-xs text-primary font-medium hover:bg-canvas transition-colors"
                role="menuitem"
              >
                <Plus className="w-4 h-4 text-accent-solid" />
                <span>Новый воркспейс</span>
              </button>

              <button
                type="button"
                onClick={() => {
                  setIsOpen(false);
                  setIsMembersOpen(true);
                }}
                className="w-full flex items-center gap-2 px-2.5 py-2 rounded-lg text-left text-xs text-primary font-medium hover:bg-canvas transition-colors"
                role="menuitem"
              >
                <Users className="w-4 h-4 text-muted" />
                <span>Участники и доступ</span>
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Modals */}
      <CreateWorkspaceModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        onCreated={(created) => onWorkspaceChange?.(created)}
      />

      <WorkspaceMembersModal
        isOpen={isMembersOpen}
        onClose={() => setIsMembersOpen(false)}
        workspace={activeWorkspace}
        currentUser={currentUser}
      />
    </>
  );
};
