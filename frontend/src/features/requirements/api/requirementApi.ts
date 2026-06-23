import { post } from '@/shared/api/httpClient';
import type {
  Requirement,
  RequirementMutation,
  RequirementUpdateInput,
} from '../model/requirement.types';

export function listRequirements(taskId: number): Promise<Requirement[]> {
  return post<Requirement[]>('/api/v1/requirement/list', { task_id: taskId });
}

export function createRequirement(
  taskId: number,
  definition: string,
): Promise<RequirementMutation> {
  return post<RequirementMutation>('/api/v1/requirement/create', { task_id: taskId, definition });
}

export function updateRequirement(input: RequirementUpdateInput): Promise<RequirementMutation> {
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
