import type { TestCase } from '@/features/test-cases/model/testCase.types';

export type { Epic } from '@/features/epics/model/epic.types';

export interface ReferenceOption {
  slug: string;
  priority: number;
}

export interface Change {
  id: number;
  version: number;
  ref: number;
  slug: string;
  project_id: number;
  epic_id?: number | null;
  change_phase: string;
  change_types: string[];
  title: string;
  requirement_body: string;
  requirement_html: string;
  pull_request_body: string;
  pull_request_html: string;
  pull_request_url: string;
  closed: boolean;
  done_tc: number;
  total_tc: number;
  completed: number;
  created: string;
  modified: string;
}

export interface ChangeDetail {
  change: Change;
  test_cases: TestCase[];
}

export interface ChangeReferences {
  phases: ReferenceOption[];
  types: ReferenceOption[];
}

export interface ChangeRenderedBody {
  id: number;
  requirement_html: string;
  pull_request_html: string;
}

export interface ChangeRenderedBodiesResponse {
  bodies: ChangeRenderedBody[];
}

export interface ChangeCreateInput {
  project_id: number;
  epic_id?: number | null;
  title: string;
  requirement_body?: string;
  change_types: string[];
}

export interface SelectOption {
  label: string;
  value: string;
}
