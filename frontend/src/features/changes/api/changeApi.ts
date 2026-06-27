import { post } from '@/shared/api/httpClient';
import type {
  Change,
  ChangeCreateInput,
  ChangeDetail,
  ChangeReferences,
  ChangeRenderedBodiesResponse,
  ChangeUpdateInput,
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

export function updateChange(input: ChangeUpdateInput): Promise<Change> {
  return post<Change>('/api/v1/change/update', input);
}

export function updateChangeEpic(id: number, epicId: number | null): Promise<Change> {
  return post<Change>('/api/v1/change/update-epic', { id, epic_id: epicId });
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
