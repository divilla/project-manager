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
  id: number;
  name: string;
}

export interface ReferenceOption {
  slug: string;
  priority: number;
}

export interface Task {
  id: number;
  version: number;
  project_id: number;
  parent_id?: number;
  task_phase: string;
  task_type: string;
  name: string;
  description: string;
  difficulty: number;
  done_req: number;
  total_req: number;
  completed: number;
  priority: number;
  created: string;
  modified: string;
}

export interface Requirement {
  id: number;
  version: number;
  task_id: number;
  definition: string;
  done: boolean;
  created: string;
  modified: string;
}

export interface TaskDetail {
  task: Task;
  requirements: Requirement[];
}

export interface RequirementMutation {
  requirement?: Requirement;
  task: Task;
  requirements: Requirement[];
}

export interface TaskReferences {
  phases: ReferenceOption[];
  types: ReferenceOption[];
}

export function listProjects(): Promise<Project[]> {
  return post<Project[]>('/api/v1/project/list');
}

export function createProject(name: string): Promise<Project> {
  return post<Project>('/api/v1/project/create', { name });
}

export function updateProject(id: number, name: string): Promise<Project> {
  return post<Project>('/api/v1/project/update', { id, name });
}

export function deleteProject(id: number): Promise<void> {
  return post<void>('/api/v1/project/delete', { id });
}

export function getTaskReferences(): Promise<TaskReferences> {
  return post<TaskReferences>('/api/v1/task/reference');
}

export function listTasks(projectId: number): Promise<Task[]> {
  return post<Task[]>('/api/v1/task/list', { project_id: projectId });
}

export function getTask(id: number): Promise<TaskDetail> {
  return post<TaskDetail>('/api/v1/task/get', { id });
}

export function createTask(input: {
  project_id: number;
  name: string;
  description?: string;
  task_phase?: string;
  task_type?: string;
  difficulty?: number;
  priority?: number;
}): Promise<Task> {
  return post<Task>('/api/v1/task/create', input);
}

export function updateTask(input: {
  id: number;
  name: string;
  description?: string;
  task_type?: string;
}): Promise<Task> {
  return post<Task>('/api/v1/task/update', input);
}

export function updateTaskDifficulty(id: number, difficulty: number): Promise<Task> {
  return post<Task>('/api/v1/task/update-difficulty', { id, difficulty });
}

export function updateTaskPriority(id: number, priority: number): Promise<Task> {
  return post<Task>('/api/v1/task/update-priority', { id, priority });
}

export function updateTaskParent(id: number, parentId: number | null): Promise<Task> {
  return post<Task>('/api/v1/task/update-parent', { id, parent_id: parentId });
}

export function updateTaskPhase(id: number, taskPhase: string): Promise<Task> {
  return post<Task>('/api/v1/task/update-phase', { id, task_phase: taskPhase });
}

export function deleteTask(id: number): Promise<void> {
  return post<void>('/api/v1/task/delete', { id });
}

export function listRequirements(taskId: number): Promise<Requirement[]> {
  return post<Requirement[]>('/api/v1/requirement/list', { task_id: taskId });
}

export function createRequirement(
  taskId: number,
  definition: string,
): Promise<RequirementMutation> {
  return post<RequirementMutation>('/api/v1/requirement/create', { task_id: taskId, definition });
}

export function updateRequirement(input: {
  id: number;
  definition: string;
}): Promise<RequirementMutation> {
  return post<RequirementMutation>('/api/v1/requirement/update', input);
}

export function updateRequirementDone(id: number, done: boolean): Promise<RequirementMutation> {
  return post<RequirementMutation>('/api/v1/requirement/update-done', { id, done });
}

export function updateRequirementTask(id: number, taskId: number): Promise<RequirementMutation> {
  return post<RequirementMutation>('/api/v1/requirement/update-task', { id, task_id: taskId });
}

export function deleteRequirement(id: number): Promise<RequirementMutation> {
  return post<RequirementMutation>('/api/v1/requirement/delete', { id });
}
