import type { Requirement } from '@/features/requirements/model/requirement.types';

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

export interface TaskDetail {
  task: Task;
  requirements: Requirement[];
}

export interface TaskReferences {
  phases: ReferenceOption[];
  types: ReferenceOption[];
}

export interface TaskCreateInput {
  project_id: number;
  name: string;
  description?: string;
  task_phase?: string;
  task_type?: string;
  difficulty?: number;
  priority?: number;
}

export interface TaskUpdateInput {
  id: number;
  name: string;
  description?: string;
  task_type?: string;
}

export interface SelectOption {
  label: string;
  value: string;
}
