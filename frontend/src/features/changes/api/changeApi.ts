import { post } from '@/shared/api/httpClient';
import type {
  Change,
  ChangeCreateInput,
  ChangeDetail,
  ChangeReferences,
  ChangeRenderedBodiesResponse,
} from '../model/change.types';

export function getChangeReferences(): Promise<ChangeReferences> {
  return post<ChangeReferences>('/api/v1/change/reference');
}

export function listChanges(projectId: number): Promise<Change[]> {
  return post<Change[]>('/api/v1/change/list', { project_id: projectId });
}

export function getChange(id: number): Promise<ChangeDetail> {
  return post<ChangeDetail>('/api/v1/change/get', { id });
}

export function getRenderedChangeBodies(ids: number[]): Promise<ChangeRenderedBodiesResponse> {
  return post<ChangeRenderedBodiesResponse>('/api/v1/change/rendered-bodies', { ids });
}

export function createChange(input: ChangeCreateInput): Promise<Change> {
  return post<Change>('/api/v1/change/create', input);
}

export function updateChangeEpic(id: number, epicId: number | null): Promise<Change> {
  return post<Change>('/api/v1/change/update-epic', { id, epic_id: epicId });
}

export function updateChangeTitle(id: number, title: string): Promise<Change> {
  return post<Change>('/api/v1/change/update-title', { id, title });
}

export function updateChangeRequirementBody(id: number, requirementBody: string): Promise<Change> {
  return post<Change>('/api/v1/change/update-requirement-body', { id, requirement_body: requirementBody });
}

export function updateChangeTypes(id: number, changeTypes: string[]): Promise<Change> {
  return post<Change>('/api/v1/change/update-change-types', { id, change_types: changeTypes });
}

export function updateChangePullRequestBody(id: number, pullRequestBody: string): Promise<Change> {
  return post<Change>('/api/v1/change/update-pull-request-body', { id, pull_request_body: pullRequestBody });
}

export function updateChangePullRequestURL(id: number, pullRequestURL: string): Promise<Change> {
  return post<Change>('/api/v1/change/update-pull-request-url', { id, pull_request_url: pullRequestURL });
}

export function updateChangePhase(id: number, changePhase: string): Promise<Change> {
  return post<Change>('/api/v1/change/update-phase', { id, change_phase: changePhase });
}

export function updateChangeClosed(id: number, closed: boolean): Promise<Change> {
  return post<Change>('/api/v1/change/update-closed', { id, closed });
}

export function deleteChange(id: number): Promise<void> {
  return post<void>('/api/v1/change/delete', { id });
}
