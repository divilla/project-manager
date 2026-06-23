import type { Task, TaskDetail, TaskReferences } from './task.types';

export function taskFixture(overrides: Partial<Task> = {}): Task {
  const task: Task = {
    id: 1,
    version: 0,
    project_id: 1,
    task_phase: 'backlog',
    task_type: 'task',
    name: 'Task',
    description: '',
    description_html: '',
    difficulty: 1,
    done_req: 0,
    total_req: 0,
    completed: 0,
    priority: 0,
    created: '2026-01-01T00:00:00Z',
    modified: '2026-01-01T00:00:00Z',
  };
  return { ...task, ...overrides };
}

export function taskDetailFixture(overrides: Partial<TaskDetail> = {}): TaskDetail {
  return {
    task: taskFixture(),
    requirements: [],
    ...overrides,
  };
}

export function taskReferencesFixture(overrides: Partial<TaskReferences> = {}): TaskReferences {
  return {
    phases: [
      { slug: 'backlog', priority: 1 },
      { slug: 'review', priority: 2 },
    ],
    types: [
      { slug: 'task', priority: 1 },
      { slug: 'feature', priority: 2 },
    ],
    ...overrides,
  };
}
