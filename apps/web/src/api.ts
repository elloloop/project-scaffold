export type User = {
  id: string;
  email: string;
  tenant_id: string;
  role: string;
};

export type Session = {
  token: string;
  user: User;
  expires_at: string;
};

export type Task = {
  id: string;
  title: string;
  completed: boolean;
  created_by: string;
  created_at: string;
  updated_at: string;
};

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly code: string,
  ) {
    super(message);
  }
}

const API_BASE = (import.meta.env.VITE_API_BASE_URL ?? "").replace(/\/$/, "");

async function request<T>(
  path: string,
  init: RequestInit = {},
  token?: string,
): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE}${path}`, { ...init, headers });
  if (!response.ok) {
    let code = "http_error";
    let message = `Request failed with status ${response.status}`;
    try {
      const body = (await response.json()) as {
        code?: string;
        message?: string;
      };
      code = body.code ?? code;
      message = body.message ?? message;
    } catch {
      // Keep the status-derived message.
    }
    throw new ApiError(message, response.status, code);
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

export const api = {
  login(email: string, password: string): Promise<Session> {
    return request<Session>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  },

  session(token: string): Promise<{ user: User }> {
    return request<{ user: User }>("/api/session", {}, token);
  },

  listTasks(token: string): Promise<{ items: Task[] }> {
    return request<{ items: Task[] }>("/api/tasks", {}, token);
  },

  createTask(token: string, title: string): Promise<Task> {
    return request<Task>(
      "/api/tasks",
      {
        method: "POST",
        body: JSON.stringify({ title }),
      },
      token,
    );
  },

  updateTask(
    token: string,
    id: string,
    patch: { title?: string; completed?: boolean },
  ): Promise<Task> {
    return request<Task>(
      `/api/tasks/${encodeURIComponent(id)}`,
      {
        method: "PATCH",
        body: JSON.stringify(patch),
      },
      token,
    );
  },

  deleteTask(token: string, id: string): Promise<void> {
    return request<void>(
      `/api/tasks/${encodeURIComponent(id)}`,
      {
        method: "DELETE",
      },
      token,
    );
  },
};
