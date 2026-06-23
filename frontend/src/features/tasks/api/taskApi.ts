import { post } from '@/shared/api/httpClient';
import type {
  Task,
  TaskCreateInput,
  TaskDetail,
  TaskReferences,
  TaskRenderedDescriptionsResponse,
  TaskUpdateInput,
} from '../model/task.types';

export function getTaskReferences(): Promise<TaskReferences> {
  return post<TaskReferences>('/api/v1/task/reference');
}

export function listTasks(projectId: number): Promise<Task[]> {
  return post<Task[]>('/api/v1/task/list', { project_id: projectId });
}

export function getTask(id: number): Promise<TaskDetail> {
  return post<TaskDetail>('/api/v1/task/get', { id });
}

export function getRenderedTaskDescriptions(ids: number[]): Promise<TaskRenderedDescriptionsResponse> {
  return post<TaskRenderedDescriptionsResponse>('/api/v1/task/rendered-descriptions', { ids });
}

export function createTask(input: TaskCreateInput): Promise<Task> {
  return post<Task>('/api/v1/task/create', input);
}

export function updateTask(input: TaskUpdateInput): Promise<Task> {
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
