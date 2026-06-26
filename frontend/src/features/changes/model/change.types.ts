import type { Requirement } from '@/features/requirements/model/requirement.types';

export interface ReferenceOption {
  slug: string;
  priority: number;
}

export interface Change {
  id: number;
  version: number;
  project_id: number;
  epic_id?: number | null;
  change_phase: string;
  change_types: string[];
  title: string;
  body: string;
  body_html: string;
  closed: boolean;
  done_req: number;
  total_req: number;
  completed: number;
  created: string;
  modified: string;
}

export interface ChangeDetail {
  change: Change;
  requirements: Requirement[];
}

export interface ChangeReferences {
  phases: ReferenceOption[];
  types: ReferenceOption[];
}

export interface ChangeRenderedBody {
  id: number;
  body_html: string;
}

export interface ChangeRenderedBodiesResponse {
  bodies: ChangeRenderedBody[];
}

export interface ChangeCreateInput {
  project_id: number;
  epic_id?: number | null;
  title: string;
  body?: string;
  change_phase: string;
  change_types: string[];
}

export interface ChangeUpdateInput {
  id: number;
  title: string;
  body?: string;
  change_types: string[];
}

export interface SelectOption {
  label: string;
  value: string;
}

export interface Epic {
  id: number;
  version: number;
  project_id: number;
  name: string;
  done_req: number;
  total_req: number;
  completed: number;
  created: string;
  modified: string;
}

export interface EpicCreateInput {
  project_id: number;
  name: string;
}

export interface EpicUpdateInput {
  id: number;
  name: string;
}
