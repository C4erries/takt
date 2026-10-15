import React, { useState, useEffect } from 'react';
import { Sparkles } from 'lucide-react';
import {
  Modal,
  Button,
  Skeleton,
  ErrorBanner,
} from '@/shared/components/ui';
import {
  useInvitePreviewQuery,
  useAcceptInviteMutation,
} from '@/entities/workspace';

export interface AcceptInviteModalProps {
  token?: string | null;
  onAccepted?: () => void;
}

export const AcceptInviteModal: React.FC<AcceptInviteModalProps> = ({
  token: propToken,
  onAccepted,
}) => {
  const [activeToken, setActiveToken] = useState<string | null>(propToken ?? null);
  const [isOpen, setIsOpen] = useState(false);
  const [acceptError, setAcceptError] = useState<string | null>(null);

  // Detect ?invite=token in URL if prop not passed
  useEffect(() => {
    if (propToken !== undefined) {
      setActiveToken(propToken);
      setIsOpen(Boolean(propToken));
      return;
    }

    const params = new URLSearchParams(window.location.search);
    const inviteParam = params.get('invite');
    if (inviteParam) {
      setActiveToken(inviteParam);
      setIsOpen(true);
    }
  }, [propToken]);

  const {
    data: preview,
    isLoading,
    isError,
  } = useInvitePreviewQuery(activeToken);

  const acceptMutation = useAcceptInviteMutation();

  const handleClose = () => {
    setIsOpen(false);
    setActiveToken(null);
    // Remove ?invite from URL without reload
    const url = new URL(window.location.href);
    url.searchParams.delete('invite');
    window.history.replaceState({}, document.title, url.pathname + (url.search ? url.search : ''));
  };

  const handleAccept = async () => {
    if (!activeToken) return;
    setAcceptError(null);
    try {
      await acceptMutation.mutateAsync(activeToken);
      handleClose();
      onAccepted?.();
    } catch {
      setAcceptError('Не удалось принять приглашение. Возможно, вы уже являетесь участником.');
    }
  };

  if (!isOpen || !activeToken) return null;

  const workspaceName = preview?.workspaceName || preview?.workspace_name || 'Воркспейс';
  const invitedBy = preview?.invitedByName || preview?.invited_by_name || 'Коллега';

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Приглашение в воркспейс"
      description="Вам открыт доступ к совместному рабочему пространству"
      className="max-w-md"
      footer={
        <div className="flex items-center justify-end gap-2.5 w-full">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleClose}
            disabled={acceptMutation.isPending}
          >
            Отклонить
          </Button>
          <Button
            variant="solid"
            size="sm"
            onClick={handleAccept}
            disabled={isLoading || isError || acceptMutation.isPending}
          >
            {acceptMutation.isPending ? 'Вход...' : 'Присоединиться'}
          </Button>
        </div>
      }
    >
      <div className="space-y-4">
        {acceptError && (
          <ErrorBanner
            message={acceptError}
          />
        )}

        {isLoading && (
          <div className="space-y-3 p-4">
            <Skeleton className="h-6 w-3/4 rounded-lg" />
            <Skeleton className="h-4 w-1/2 rounded-lg" />
          </div>
        )}

        {isError && (
          <ErrorBanner
            message="Приглашение недействительно или срок его действия истёк"
          />
        )}

        {!isLoading && !isError && preview && (
          <div className="p-5 rounded-xl border border-subtle bg-canvas/60 text-center space-y-3">
            <div className="w-12 h-12 rounded-2xl bg-accent-surface text-accent-solid mx-auto flex items-center justify-center shadow-washi-sm">
              <Sparkles className="w-6 h-6" />
            </div>

            <div className="space-y-1">
              <span className="text-xs font-mono uppercase tracking-wider text-muted block">
                Вас пригласил {invitedBy}
              </span>
              <h3 className="font-serif text-2xl font-normal text-primary">
                {workspaceName}
              </h3>
            </div>

            <p className="text-xs text-muted leading-relaxed">
              Присоединяйтесь, чтобы просматривать общие задачи, проекты и канбан-доски в едином ритме Takt.
            </p>
          </div>
        )}
      </div>
    </Modal>
  );
};
