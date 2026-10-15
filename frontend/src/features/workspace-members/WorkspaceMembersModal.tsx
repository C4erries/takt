import React, { useState } from 'react';
import {
  Link2,
  Check,
  UserX,
  LogOut,
  Users,
} from 'lucide-react';
import {
  Modal,
  Button,
  Badge,
  Skeleton,
  ErrorBanner,
  EmptyState,
} from '@/shared/components/ui';
import type { User } from '@/shared/api';
import type { Workspace } from '@/entities/workspace';
import {
  useWorkspaceMembersQuery,
  useRemoveMemberMutation,
  useCreateInviteMutation,
} from '@/entities/workspace';

export interface WorkspaceMembersModalProps {
  isOpen: boolean;
  onClose: () => void;
  workspace: Workspace | null;
  currentUser: User | null;
}

export const WorkspaceMembersModal: React.FC<WorkspaceMembersModalProps> = ({
  isOpen,
  onClose,
  workspace,
  currentUser,
}) => {
  const [copied, setCopied] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const {
    data: members = [],
    isLoading,
    isError,
    refetch,
  } = useWorkspaceMembersQuery(workspace?.id);

  const removeMemberMutation = useRemoveMemberMutation(workspace?.id);
  const createInviteMutation = useCreateInviteMutation(workspace?.id);

  const isOwner = workspace?.role === 'owner';

  const handleCopyInvite = async () => {
    setActionError(null);
    try {
      const invite = await createInviteMutation.mutateAsync();
      const token = invite.token;
      const origin = window.location.origin;
      const pathname = window.location.pathname;
      const fullUrl = `${origin}${pathname}?invite=${encodeURIComponent(token)}`;

      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(fullUrl);
      }
      setCopied(true);
      setTimeout(() => setCopied(false), 2500);
    } catch {
      setActionError('Не удалось создать ссылку для приглашения');
    }
  };

  const handleRemoveMember = async (userId: string, isSelf: boolean) => {
    if (
      !window.confirm(
        isSelf
          ? 'Вы уверены, что хотите покинуть этот воркспейс?'
          : 'Вы уверены, что хотите исключить этого участника?'
      )
    ) {
      return;
    }
    setActionError(null);
    try {
      await removeMemberMutation.mutateAsync(userId);
      if (isSelf) {
        onClose();
      }
    } catch {
      setActionError('Ошибка при удалении участника');
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Участники и доступ"
      description={workspace?.name ? `Воркспейс «${workspace.name}»` : undefined}
      className="max-w-lg"
    >
      <div className="space-y-6">
        {actionError && (
          <ErrorBanner
            message={actionError}
          />
        )}

        {/* Invite Link Section */}
        <div className="p-4 rounded-xl border border-subtle bg-canvas/60 space-y-3">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <h4 className="text-xs font-medium text-primary flex items-center gap-1.5">
                <Link2 className="w-3.5 h-3.5 text-accent-solid" />
                Пригласить по ссылке
              </h4>
              <p className="text-[11px] text-muted">
                {isOwner
                  ? 'Отправьте ссылку коллеге для совместной работы'
                  : 'Приглашать участников может только владелец'}
              </p>
            </div>
            {isOwner && (
              <Button
                variant={copied ? 'secondary' : 'outline'}
                size="sm"
                onClick={handleCopyInvite}
                disabled={createInviteMutation.isPending}
                className="text-xs"
              >
                {copied ? (
                  <span className="flex items-center gap-1.5 text-accent-solid">
                    <Check className="w-3.5 h-3.5" />
                    Скопировано
                  </span>
                ) : (
                  <span className="flex items-center gap-1.5">
                    <Link2 className="w-3.5 h-3.5" />
                    Копировать
                  </span>
                )}
              </Button>
            )}
          </div>
        </div>

        {/* Members List Section */}
        <div className="space-y-3">
          <div className="flex items-center justify-between px-1">
            <h4 className="text-xs font-mono uppercase tracking-wider text-muted flex items-center gap-1.5">
              <Users className="w-3.5 h-3.5" />
              Участники ({members.length})
            </h4>
          </div>

          {isLoading && (
            <div className="space-y-2">
              <Skeleton className="h-12 w-full rounded-xl" />
              <Skeleton className="h-12 w-full rounded-xl" />
            </div>
          )}

          {isError && (
            <ErrorBanner
              message="Не удалось загрузить список участников"
              onRetry={() => refetch()}
            />
          )}

          {!isLoading && !isError && members.length === 0 && (
            <EmptyState
              title="Нет участников"
              description="В этом воркспейсе пока никого нет"
            />
          )}

          {!isLoading && !isError && members.length > 0 && (
            <div className="divide-y divide-subtle/50 rounded-xl border border-subtle bg-surface overflow-hidden">
              {members.map((member) => {
                const memberUserId = member.userId || member.user_id || member.id;
                const memberDisplayName = member.displayName || member.display_name || member.email;
                const isCurrent = currentUser?.id === memberUserId;
                const isMemberOwner = member.role === 'owner';

                return (
                  <div
                    key={member.id}
                    className="flex items-center justify-between p-3 transition-colors hover:bg-canvas/40"
                  >
                    <div className="flex items-center gap-3 min-w-0">
                      <div className="w-8 h-8 rounded-full bg-accent-surface text-primary flex items-center justify-center font-medium text-xs shrink-0 select-none">
                        {memberDisplayName.charAt(0).toUpperCase()}
                      </div>
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="text-xs font-medium text-primary truncate">
                            {memberDisplayName}
                          </span>
                          {isCurrent && (
                            <span className="text-[10px] text-muted font-mono">
                              (вы)
                            </span>
                          )}
                        </div>
                        <span className="text-[11px] text-muted truncate block">
                          {member.email}
                        </span>
                      </div>
                    </div>

                    <div className="flex items-center gap-2 shrink-0">
                      <Badge
                        variant={isMemberOwner ? 'matcha' : 'subtle'}
                        size="sm"
                      >
                        {isMemberOwner ? 'Владелец' : 'Участник'}
                      </Badge>

                      {isOwner && !isMemberOwner && !isCurrent && (
                        <button
                          type="button"
                          onClick={() => handleRemoveMember(memberUserId, false)}
                          className="p-1 rounded-md text-muted hover:text-tag-persimmon hover:bg-tag-persimmon/10 transition-colors"
                          title="Исключить участника"
                        >
                          <UserX className="w-4 h-4" />
                        </button>
                      )}

                      {!isOwner && isCurrent && (
                        <button
                          type="button"
                          onClick={() => handleRemoveMember(memberUserId, true)}
                          className="p-1 rounded-md text-muted hover:text-tag-persimmon hover:bg-tag-persimmon/10 transition-colors"
                          title="Покинуть воркспейс"
                        >
                          <LogOut className="w-4 h-4" />
                        </button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </Modal>
  );
};
