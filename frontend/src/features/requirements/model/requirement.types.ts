import type { Change } from '@/features/changes/model/change.types';

export interface Requirement {
  id: number;
  version: number;
  change_id: number;
  definition: string;
  done: boolean;
  created: string;
  modified: string;
}

export interface RequirementMutation {
  requirement?: Requirement;
  change: Change;
  requirements: Requirement[];
}

export interface RequirementUpdateInput {
  id: number;
  definition: string;
}
