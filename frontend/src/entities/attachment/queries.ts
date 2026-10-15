import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getAuthToken } from '@/shared/api';
import type { Attachment } from '@/shared/api';
import { attachmentsApi } from '@/shared/api';

export const ATTACHMENTS_QUERY_KEY = 'task-attachments';

export const getTaskAttachmentsQueryKey = (workspaceId?: string, taskId?: string) => [
  ATTACHMENTS_QUERY_KEY,
  workspaceId,
  taskId,
];

export const useTaskAttachmentsQuery = (workspaceId?: string, taskId?: string) => {
  return useQuery<Attachment[], Error>({
    queryKey: getTaskAttachmentsQueryKey(workspaceId, taskId),
    queryFn: async () => {
      if (!workspaceId || !taskId) return [];
      if (!getAuthToken()) {
        return [];
      }
      return attachmentsApi.listTaskAttachments(workspaceId, taskId);
    },
    enabled: Boolean(workspaceId && taskId),
  });
};

export const useUploadAttachmentMutation = (workspaceId?: string, taskId?: string) => {
  const queryClient = useQueryClient();

  return useMutation<Attachment, Error, File>({
    mutationFn: async (file: File) => {
      if (!workspaceId || !taskId) {
        throw new Error('Workspace ID and Task ID are required to upload attachments');
      }
      return attachmentsApi.uploadTaskAttachment(workspaceId, taskId, file);
    },
    onSuccess: (newAttachment) => {
      queryClient.setQueryData<Attachment[]>(
        getTaskAttachmentsQueryKey(workspaceId, taskId),
        (old = []) => [...old, newAttachment]
      );
    },
    onSettled: () => {
      queryClient.invalidateQueries({
        queryKey: getTaskAttachmentsQueryKey(workspaceId, taskId),
      });
    },
  });
};

export const useDeleteAttachmentMutation = (workspaceId?: string, taskId?: string) => {
  const queryClient = useQueryClient();

  return useMutation<void, Error, string, { previousAttachments?: Attachment[] }>({
    mutationFn: async (attachmentId: string) => {
      if (!workspaceId) {
        throw new Error('Workspace ID is required to delete attachments');
      }
      return attachmentsApi.deleteAttachment(workspaceId, attachmentId);
    },
    onMutate: async (attachmentId: string) => {
      const queryKey = getTaskAttachmentsQueryKey(workspaceId, taskId);
      await queryClient.cancelQueries({ queryKey });

      const previousAttachments = queryClient.getQueryData<Attachment[]>(queryKey);

      queryClient.setQueryData<Attachment[]>(queryKey, (old = []) =>
        old.filter((item) => item.id !== attachmentId)
      );

      return { previousAttachments };
    },
    onError: (_err, _attachmentId, context: { previousAttachments?: Attachment[] } | undefined) => {
      if (context?.previousAttachments) {
        queryClient.setQueryData(
          getTaskAttachmentsQueryKey(workspaceId, taskId),
          context.previousAttachments
        );
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({
        queryKey: getTaskAttachmentsQueryKey(workspaceId, taskId),
      });
    },
  });
};
