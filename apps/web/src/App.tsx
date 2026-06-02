import { Check, LogOut, Plus, RefreshCcw, Trash2 } from "lucide-react";
import { FormEvent, useEffect, useMemo, useState } from "react";

import { api, ApiError, Task, User } from "./api";
import { canSubmitTaskTitle, normalizeTaskTitle } from "./taskRules";

const TOKEN_KEY = "project-scaffold-token";

export function App() {
  const [token, setToken] = useState(
    () => localStorage.getItem(TOKEN_KEY) ?? "",
  );
  const [user, setUser] = useState<User | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [email, setEmail] = useState("demo@example.com");
  const [password, setPassword] = useState("demo");
  const [title, setTitle] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");

  const openTasks = useMemo(
    () => tasks.filter((task) => !task.completed).length,
    [tasks],
  );

  useEffect(() => {
    if (!token) {
      return;
    }
    void refresh(token);
  }, [token]);

  async function refresh(activeToken = token) {
    if (!activeToken) {
      return;
    }
    setLoading(true);
    setMessage("");
    try {
      const [session, list] = await Promise.all([
        api.session(activeToken),
        api.listTasks(activeToken),
      ]);
      setUser(session.user);
      setTasks(list.items);
    } catch (error) {
      handleError(error);
      logout();
    } finally {
      setLoading(false);
    }
  }

  async function login(event: FormEvent) {
    event.preventDefault();
    setLoading(true);
    setMessage("");
    try {
      const session = await api.login(email, password);
      localStorage.setItem(TOKEN_KEY, session.token);
      setToken(session.token);
      setUser(session.user);
      const list = await api.listTasks(session.token);
      setTasks(list.items);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  function logout() {
    localStorage.removeItem(TOKEN_KEY);
    setToken("");
    setUser(null);
    setTasks([]);
  }

  async function addTask(event: FormEvent) {
    event.preventDefault();
    const normalized = normalizeTaskTitle(title);
    if (!canSubmitTaskTitle(normalized)) {
      setMessage("Task title must be between 1 and 120 characters.");
      return;
    }
    setLoading(true);
    setMessage("");
    try {
      const task = await api.createTask(token, normalized);
      setTasks((current) => [task, ...current]);
      setTitle("");
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function toggleTask(task: Task) {
    setLoading(true);
    setMessage("");
    try {
      const updated = await api.updateTask(token, task.id, {
        completed: !task.completed,
      });
      setTasks((current) =>
        current.map((item) => (item.id === task.id ? updated : item)),
      );
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function deleteTask(task: Task) {
    setLoading(true);
    setMessage("");
    try {
      await api.deleteTask(token, task.id);
      setTasks((current) => current.filter((item) => item.id !== task.id));
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  function handleError(error: unknown) {
    if (error instanceof ApiError) {
      setMessage(error.message);
      return;
    }
    setMessage("The app could not reach the backend.");
  }

  if (!token || !user) {
    return (
      <main className="shell auth-shell">
        <section className="auth-panel" aria-labelledby="login-title">
          <div>
            <p className="eyebrow">Project Scaffold</p>
            <h1 id="login-title">Sign in</h1>
            <p className="muted">
              Use the local demo account to start the CRUD flow.
            </p>
          </div>
          <form className="stack" onSubmit={login}>
            <label>
              Email
              <input
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                type="email"
                autoComplete="email"
              />
            </label>
            <label>
              Password
              <input
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                type="password"
                autoComplete="current-password"
              />
            </label>
            <button className="primary" disabled={loading} type="submit">
              <Check size={18} aria-hidden="true" />
              Sign in
            </button>
          </form>
          {message ? <p className="notice">{message}</p> : null}
        </section>
      </main>
    );
  }

  return (
    <main className="shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">Project Scaffold</p>
          <h1>Tasks</h1>
        </div>
        <div className="session">
          <span>{user.email}</span>
          <button
            className="icon-button"
            title="Refresh"
            onClick={() => void refresh()}
            disabled={loading}
          >
            <RefreshCcw size={18} aria-hidden="true" />
          </button>
          <button className="icon-button" title="Sign out" onClick={logout}>
            <LogOut size={18} aria-hidden="true" />
          </button>
        </div>
      </header>

      <section className="summary-grid" aria-label="Task summary">
        <div>
          <span className="summary-label">Open</span>
          <strong>{openTasks}</strong>
        </div>
        <div>
          <span className="summary-label">Done</span>
          <strong>{tasks.length - openTasks}</strong>
        </div>
        <div>
          <span className="summary-label">Total</span>
          <strong>{tasks.length}</strong>
        </div>
      </section>

      <section className="task-panel" aria-labelledby="tasks-title">
        <div className="panel-header">
          <div>
            <h2 id="tasks-title">Backlog</h2>
            <p className="muted">
              Creating a task also enqueues a background job for the worker.
            </p>
          </div>
        </div>

        <form className="task-form" onSubmit={addTask}>
          <input
            value={title}
            onChange={(event) => setTitle(event.target.value)}
            placeholder="Add a task"
            maxLength={120}
            aria-label="Task title"
          />
          <button
            className="primary"
            disabled={loading || !canSubmitTaskTitle(title)}
            type="submit"
          >
            <Plus size={18} aria-hidden="true" />
            Add
          </button>
        </form>

        {message ? <p className="notice">{message}</p> : null}

        <ul className="task-list">
          {tasks.map((task) => (
            <li key={task.id} className={task.completed ? "task done" : "task"}>
              <button
                className="check-button"
                title="Toggle complete"
                onClick={() => void toggleTask(task)}
                disabled={loading}
              >
                <Check size={18} aria-hidden="true" />
              </button>
              <span>{task.title}</span>
              <button
                className="icon-button danger"
                title="Delete task"
                onClick={() => void deleteTask(task)}
                disabled={loading}
              >
                <Trash2 size={18} aria-hidden="true" />
              </button>
            </li>
          ))}
        </ul>
      </section>
    </main>
  );
}
