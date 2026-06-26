import { post } from '@/shared/api/httpClient';
import type {
  Requirement,
  RequirementMutation,
  RequirementUpdateInput,
} from '../model/requirement.types';

export function listRequirements(changeId: number): Promise<Requirement[]> {
  return post<Requirement[]>('/api/v1/requirement/list', { change_id: changeId });
}

export function createRequirement(
  changeId: number,
  definition: string,
): Promise<RequirementMutation> {
  return post<RequirementMutation>('/api/v1/requirement/create', { change_id: changeId, definition });
}

export function updateRequirement(input: RequirementUpdateInput): Promise<RequirementMutation> {
  return post<RequirementMutation>('/api/v1/requirement/update', input);
}

export function updateRequirementDone(id: number, done: boolean): Promise<RequirementMutation> {
  return post<RequirementMutation>('/api/v1/requirement/update-done', { id, done });
}

export function updateRequirementChange(id: number, changeId: number): Promise<RequirementMutation> {
  return post<RequirementMutation>('/api/v1/requirement/update-change', { id, change_id: changeId });
}

export function deleteRequirement(id: number): Promise<RequirementMutation> {
  return post<RequirementMutation>('/api/v1/requirement/delete', { id });
}
