import type { Task } from '@/features/tasks/model/task.types';

export interface Requirement {
  id: number;
  version: number;
  task_id: number;
  definition: string;
  done: boolean;
  created: string;
  modified: string;
}

export interface RequirementMutation {
  requirement?: Requirement;
  task: Task;
  requirements: Requirement[];
}

export interface RequirementUpdateInput {
  id: number;
  definition: string;
}
