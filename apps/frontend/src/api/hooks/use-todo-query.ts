import { type TApiClient, useApiClient } from "@/api";
import { QUERY_KEYS } from "@/api/query-utils";
import type { TRequests } from "@/api/types";
import { showApiErrorToast } from "@/api/utils";
import type { apiContract } from "@taskify/openapi/contracts";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import type { ServerInferResponseBody } from "@ts-rest/core";

// Type definitions for Todo operations
export type TGetTodosQuery = TRequests["Todo"]["getTodos"]["query"];
export type TGetTodosResponse = ServerInferResponseBody<
  typeof apiContract.Todo.getTodos,
  200
>;

export type TGetTodoByIdResponse = ServerInferResponseBody<
  typeof apiContract.Todo.getTodoById,
  200
>;

export type TCreateTodoPayload = TRequests["Todo"]["createTodo"]["body"];
export type TCreateTodoResponse = ServerInferResponseBody<
  typeof apiContract.Todo.createTodo,
  201
>;

export type TUpdateTodoPayload = TRequests["Todo"]["updateTodo"]["body"];
export type TUpdateTodoResponse = ServerInferResponseBody<
  typeof apiContract.Todo.updateTodo,
  200
>;

export type TTodoStatsResponse = ServerInferResponseBody<
  typeof apiContract.Todo.getTodoStats,
  200
>;

// API functions
const fetchAllTodos = async ({
  api,
  query,
}: {
  api: TApiClient;
  query?: TGetTodosQuery;
}): Promise<TGetTodosResponse> => {
  const res = await api.Todo.getTodos({ query });
  
  if (res.status === 200) {
    return res.body;
  } else {
    throw res.body;
  }
};

const fetchTodoById = async ({
  api,
  id,
}: {
  api: TApiClient;
  id: string;
}): Promise<TGetTodoByIdResponse> => {
  const res = await api.Todo.getTodoById({ params: { id } });
  
  if (res.status === 200) {
    return res.body;
  } else {
    throw res.body;
  }
};

const createTodo = async ({
  api,
  data,
}: {
  api: TApiClient;
  data: TCreateTodoPayload;
}): Promise<TCreateTodoResponse> => {
  const res = await api.Todo.createTodo({ body: data });
  
  if (res.status === 201) {
    return res.body;
  } else {
    throw res.body;
  }
};

const updateTodo = async ({
  api,
  id,
  data,
}: {
  api: TApiClient;
  id: string;
  data: TUpdateTodoPayload;
}): Promise<TUpdateTodoResponse> => {
  const res = await api.Todo.updateTodo({ params: { id }, body: data });
  
  if (res.status === 200) {
    return res.body;
  } else {
    throw res.body;
  }
};

const deleteTodo = async ({
  api,
  id,
}: {
  api: TApiClient;
  id: string;
}): Promise<void> => {
  const res = await api.Todo.deleteTodo({ params: { id } });
  
  if (res.status !== 204) {
    throw res.body;
  }
};

const fetchTodoStats = async ({
  api,
}: {
  api: TApiClient;
}): Promise<TTodoStatsResponse> => {
  const res = await api.Todo.getTodoStats();
  
  if (res.status === 200) {
    return res.body;
  } else {
    throw res.body;
  }
};

// React Query hooks
export const useGetAllTodos = ({
  query,
}: {
  query?: TGetTodosQuery;
} = {}) => {
  const api = useApiClient();

  return useQuery({
    queryKey: [QUERY_KEYS.TODOS.ALL_TODOS, query],
    queryFn: () => fetchAllTodos({ api, query }),
    placeholderData: {
      data: [],
      total: 0,
      page: 1,
      limit: 20,
      totalPages: 0,
    },
  });
};

export const useGetTodoById = ({
  id,
  enabled = true,
}: {
  id: string;
  enabled?: boolean;
}) => {
  const api = useApiClient();

  return useQuery({
    queryKey: [QUERY_KEYS.TODOS.GET_TODO_BY_ID, id],
    queryFn: () => fetchTodoById({ api, id }),
    enabled: enabled && !!id,
  });
};

export const useCreateTodo = () => {
  const api = useApiClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ body }: { body: TCreateTodoPayload }) => createTodo({ api, data: body }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.TODOS.ALL_TODOS],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.TODOS.TODO_STATS],
      });
    },
    onError: (err) => {
      showApiErrorToast(err, "Failed to create todo");
    },
  });
};

export const useUpdateTodo = () => {
  const api = useApiClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ todoId, body }: { todoId: string; body: TUpdateTodoPayload }) =>
      updateTodo({ api, id: todoId, data: body }),
    onMutate: async ({ todoId, body }) => {
      await queryClient.cancelQueries({
        queryKey: [QUERY_KEYS.TODOS.ALL_TODOS],
      });
      await queryClient.cancelQueries({
        queryKey: [QUERY_KEYS.TODOS.GET_TODO_BY_ID, todoId],
      });

      const previousLists = queryClient.getQueriesData<TGetTodosResponse>({
        queryKey: [QUERY_KEYS.TODOS.ALL_TODOS],
      });
      const previousDetail = queryClient.getQueryData<TGetTodoByIdResponse>([
        QUERY_KEYS.TODOS.GET_TODO_BY_ID,
        todoId,
      ]);

      previousLists.forEach(([queryKey, cached]) => {
        if (!cached) return;
        queryClient.setQueryData<TGetTodosResponse>(queryKey, {
          ...cached,
          data: cached.data.map((todo) => {
            if (todo.id === todoId) return { ...todo, ...body };
            if (todo.children?.some((child) => child.id === todoId)) {
              return {
                ...todo,
                children: todo.children.map((child) =>
                  child.id === todoId ? { ...child, ...body } : child,
                ),
              };
            }
            return todo;
          }),
        });
      });

      if (previousDetail) {
        queryClient.setQueryData<TGetTodoByIdResponse>(
          [QUERY_KEYS.TODOS.GET_TODO_BY_ID, todoId],
          { ...previousDetail, ...body },
        );
      }

      return { previousLists, previousDetail, todoId };
    },
    onError: (err, _vars, context) => {
      context?.previousLists.forEach(([queryKey, data]) => {
        queryClient.setQueryData(queryKey, data);
      });
      if (context?.previousDetail) {
        queryClient.setQueryData(
          [QUERY_KEYS.TODOS.GET_TODO_BY_ID, context.todoId],
          context.previousDetail,
        );
      }
      showApiErrorToast(err, "Failed to update todo");
    },
    onSettled: (_, __, { todoId }) => {
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.TODOS.ALL_TODOS],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.TODOS.GET_TODO_BY_ID, todoId],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.TODOS.TODO_STATS],
      });
    },
  });
};

export const useDeleteTodo = () => {
  const api = useApiClient();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ todoId }: { todoId: string }) => deleteTodo({ api, id: todoId }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.TODOS.ALL_TODOS],
      });
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.TODOS.TODO_STATS],
      });
    },
    onError: (err) => {
      showApiErrorToast(err, "Failed to delete todo");
    },
  });
};

export const useGetTodoStats = () => {
  const api = useApiClient();

  return useQuery({
    queryKey: [QUERY_KEYS.TODOS.TODO_STATS],
    queryFn: () => fetchTodoStats({ api }),
    placeholderData: {
      total: 0,
      draft: 0,
      active: 0,
      completed: 0,
      archived: 0,
      overdue: 0,
    },
  });
};