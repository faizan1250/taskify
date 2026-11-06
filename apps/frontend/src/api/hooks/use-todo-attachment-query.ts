import { type TApiClient, useApiClient } from "@/api";
import { QUERY_KEYS } from "@/api/query-utils";
import type { TRequests } from "@/api/types";
import { showApiErrorToast } from "@/api/utils";
import { env } from "@/config/env";
import type { apiContract } from "@taskify/openapi/contracts";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { QueryClient } from "@tanstack/react-query";
import type { ServerInferResponseBody } from "@ts-rest/core";
import axios from "axios";

// Type definitions for Todo Attachment operations
export type TUploadTodoAttachmentParams =
  TRequests["Todo"]["uploadTodoAttachment"]["params"];
export type TUploadTodoAttachmentBody =
  TRequests["Todo"]["uploadTodoAttachment"]["body"];
export type TUploadTodoAttachmentResponse = ServerInferResponseBody<
  typeof apiContract.Todo.uploadTodoAttachment,
  201
>;

export type TDeleteTodoAttachmentParams =
  TRequests["Todo"]["deleteTodoAttachment"]["params"];
export type TDeleteTodoAttachmentResponse = ServerInferResponseBody<
  typeof apiContract.Todo.deleteTodoAttachment,
  204
>;

export type TGetTodoAttachmentPresignedUrlParams =
  TRequests["Todo"]["getAttachmentPresignedURL"]["params"];
export type TGetTodoAttachmentPresignedUrlResponse = ServerInferResponseBody<
  typeof apiContract.Todo.getAttachmentPresignedURL,
  200
>;

// API functions
const uploadTodoAttachment = async ({
  api,
  todoId,
  file,
}: {
  api: TApiClient;
  todoId: string;
  file: File;
}): Promise<TUploadTodoAttachmentResponse> => {
  const formData = new FormData();
  formData.append("file", file);

  const res = await api.Todo.uploadTodoAttachment({
    params: { id: todoId },
    body: formData,
    extraHeaders: {
      "Content-Type": "multipart/form-data",
    },
  });

  if (res.status === 201) {
    return res.body;
  } else {
    throw res.body;
  }
};

const deleteTodoAttachment = async ({
  api,
  todoId,
  attachmentId,
}: {
  api: TApiClient;
  todoId: string;
  attachmentId: string;
}): Promise<void> => {
  const res = await api.Todo.deleteTodoAttachment({
    params: { id: todoId, attachmentId },
  });

  if (res.status !== 204) {
    throw res.body;
  }
};

const getTodoAttachmentPresignedUrl = async ({
  api,
  todoId,
  attachmentId,
}: {
  api: TApiClient;
  todoId: string;
  attachmentId: string;
}): Promise<TGetTodoAttachmentPresignedUrlResponse> => {
  const res = await api.Todo.getAttachmentPresignedURL({
    params: { id: todoId, attachmentId },
  });

  if (res.status === 200) {
    return res.body;
  } else {
    throw res.body;
  }
};

// React Query hooks
/** Shared invalidation after any attachment write (upload/delete). */
export const invalidateTodoAttachmentQueries = (
  queryClient: QueryClient,
  todoId: string,
) => {
  queryClient.invalidateQueries({
    queryKey: [QUERY_KEYS.TODOS.GET_TODO_BY_ID, todoId],
  });
  queryClient.invalidateQueries({
    queryKey: [QUERY_KEYS.TODOS.TODO_ATTACHMENTS, todoId],
  });
  queryClient.invalidateQueries({
    queryKey: [QUERY_KEYS.TODOS.ALL_TODOS],
  });
};

export type TUploadProgress = {
  loaded: number;
  total: number;
  /** 0–100 integer. */
  percent: number;
};

/**
 * Raw axios upload with real byte progress + cancellation. The ts-rest
 * client has no `onUploadProgress`/`signal` surface, so per-file progress
 * bars go through this instead of `uploadTodoAttachment`.
 */
export const uploadTodoAttachmentRaw = async ({
  todoId,
  file,
  getToken,
  onProgress,
  signal,
}: {
  todoId: string;
  file: File;
  getToken: () => Promise<string | null>;
  onProgress?: (progress: TUploadProgress) => void;
  signal?: AbortSignal;
}): Promise<TUploadTodoAttachmentResponse> => {
  const token = await getToken();
  const formData = new FormData();
  formData.append("file", file);

  const res = await axios.post<TUploadTodoAttachmentResponse>(
    `${env.VITE_API_URL}/api/v1/todos/${todoId}/attachments`,
    formData,
    {
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      signal,
      onUploadProgress: (event) => {
        const total = event.total ?? file.size;
        const loaded = event.loaded;
        onProgress?.({
          loaded,
          total,
          percent:
            total > 0 ? Math.min(100, Math.round((loaded / total) * 100)) : 0,
        });
      },
    },
  );
  return res.data;
};

export const useUploadTodoAttachment = () => {
  const api = useApiClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ todoId, file }: { todoId: string; file: File }) =>
      uploadTodoAttachment({ api, todoId, file }),
    onSuccess: (_, { todoId }) => {
      invalidateTodoAttachmentQueries(queryClient, todoId);
    },
    onError: (err) => {
      showApiErrorToast(err, "Failed to upload attachment");
    },
  });
};

export const useDeleteTodoAttachment = () => {
  const api = useApiClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      todoId,
      attachmentId,
    }: {
      todoId: string;
      attachmentId: string;
    }) => deleteTodoAttachment({ api, todoId, attachmentId }),
    onSuccess: (_, { todoId }) => {
      invalidateTodoAttachmentQueries(queryClient, todoId);
    },
    onError: (err) => {
      showApiErrorToast(err, "Failed to delete attachment");
    },
  });
};

export const useGetTodoAttachmentPresignedUrl = () => {
  const api = useApiClient();

  return useMutation({
    mutationFn: ({
      todoId,
      attachmentId,
    }: {
      todoId: string;
      attachmentId: string;
    }) => getTodoAttachmentPresignedUrl({ api, todoId, attachmentId }),
    onError: (err) => {
      showApiErrorToast(err, "Failed to get download link");
    },
  });
};
