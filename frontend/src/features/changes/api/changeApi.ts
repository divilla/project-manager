import { post } from '@/shared/api/httpClient';
import type {
  Change,
  ChangeCreateInput,
  ChangeDetail,
  ChangeListItem,
  ChangePhase,
  ChangeRenderedBodiesResponse,
  ChangeType,
} from '../model/change.types';

export function getChangePhases(): Promise<ChangePhase[]> {
  return post<ChangePhase[]>('/api/v1/options/change-phases-list');
}

export function getChangeTypes(): Promise<ChangeType[]> {
  return post<ChangeType[]>('/api/v1/options/change-types-list');
}

export function listChanges(projectId: number): Promise<ChangeListItem[]> {
  return post<ChangeListItem[]>('/api/v1/change/list', { project_id: projectId });
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

export function updateChangeBody(id: number, body: string): Promise<Change> {
  return post<Change>('/api/v1/change/update-body', { id, body: body });
}

export function updateChangeTypes(id: number, changeTypes: string[]): Promise<Change> {
  return post<Change>('/api/v1/change/update-change-types', { id, change_types: changeTypes });
}

export function updateChangePRBody(id: number, prBody: string): Promise<Change> {
  return post<Change>('/api/v1/change/update-pr-body', { id, pr_body: prBody });
}

export function updateChangePRUrl(id: number, prUrl: string): Promise<Change> {
  return post<Change>('/api/v1/change/update-pr-url', { id, pr_url: prUrl });
}

export function updateChangeAgentEdit(id: number, agentEdit: boolean): Promise<Change> {
  return post<Change>('/api/v1/change/update-agent-edit', { id, agent_edit: agentEdit });
}

export function updateChangePhase(id: number, changePhase: string): Promise<Change> {
  return post<Change>('/api/v1/change/update-phase', { id, change_phase: changePhase });
}

export function updateChangeOpen(id: number, open: boolean): Promise<Change> {
  return post<Change>('/api/v1/change/update-open', { id, open });
}

export function deleteChange(id: number): Promise<void> {
  return post<void>('/api/v1/change/delete', { id });
}
