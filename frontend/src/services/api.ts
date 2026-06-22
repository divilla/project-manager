const API_BASE_URL = import.meta.env.QCLI_API_BASE_URL || 'http://localhost:8080';

async function post<T>(path: string, body: unknown = {}): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const payload = (await response.json()) as T & { message?: string; error?: string };

  if (!response.ok) {
    throw new Error(payload.error || payload.message || 'Backend request failed.');
  }

  return payload;
}

export interface HealthResponse {
  status: string;
  api: string;
  database: string;
  error?: string;
}

export async function getHealth(): Promise<HealthResponse> {
  const response = await fetch(`${API_BASE_URL}/api/v1/health`);
  const payload = (await response.json()) as HealthResponse;

  if (!response.ok) {
    throw new Error(payload.error || 'Backend health check failed.');
  }

  return payload;
}

export interface Project {
  id: string;
  name: string;
}

export interface ReferenceOption {
  slug: string;
  priority: number;
}

export interface Task {
  id: string;
  project_id: string;
  parent_id?: string;
  phase: string;
  type: string;
  name: string;
  description: string;
  difficulty: number;
  complete: number;
  priority: number;
  depth: number;
  created: string;
  modified: string;
}

export interface Requirement {
  id: string;
  task_id: string;
  definition: string;
  done: boolean;
  created: string;
  modified: string;
}

export interface TaskDetail {
  task: Task;
  requirements: Requirement[];
}

export interface TaskReferences {
  phases: ReferenceOption[];
  types: ReferenceOption[];
}

export function listProjects(): Promise<Project[]> {
  return post<Project[]>('/api/project/list');
}

export function createProject(name: string): Promise<Project> {
  return post<Project>('/api/project/create', { name });
}

export function updateProject(id: string, name: string): Promise<Project> {
  return post<Project>('/api/project/update', { id, name });
}

export function deleteProject(id: string): Promise<void> {
  return post<void>('/api/project/delete', { id });
}

export function getTaskReferences(): Promise<TaskReferences> {
  return post<TaskReferences>('/api/task/reference');
}

export function listTasks(projectId: string): Promise<Task[]> {
  return post<Task[]>('/api/task/list', { project_id: projectId });
}

export function getTask(id: string): Promise<TaskDetail> {
  return post<TaskDetail>('/api/task/get', { id });
}

export function createTask(input: {
  project_id: string;
  name: string;
  description?: string;
  phase?: string;
  type?: string;
  difficulty?: number;
  priority?: number;
}): Promise<Task> {
  return post<Task>('/api/task/create', input);
}

export function updateTask(input: {
  id: string;
  name: string;
  description?: string;
  type?: string;
  difficulty?: number;
  priority?: number;
}): Promise<Task> {
  return post<Task>('/api/task/update', input);
}

export function changeTaskPhase(id: string, phase: string): Promise<Task> {
  return post<Task>('/api/task/phase', { id, phase });
}

export function deleteTask(id: string): Promise<void> {
  return post<void>('/api/task/delete', { id });
}
