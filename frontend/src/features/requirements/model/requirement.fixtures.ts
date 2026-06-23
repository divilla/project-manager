import { taskFixture } from '@/features/tasks/model/task.fixtures';
import type { Requirement, RequirementMutation } from './requirement.types';

export function requirementFixture(overrides: Partial<Requirement> = {}): Requirement {
  return {
    id: 1,
    version: 0,
    task_id: 1,
    definition: 'Requirement',
    done: false,
    created: '2026-01-01T00:00:00Z',
    modified: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

export function requirementMutationFixture(
  overrides: Partial<RequirementMutation> = {},
): RequirementMutation {
  const requirement = requirementFixture(overrides.requirement ?? {});
  return {
    requirement,
    task: taskFixture({ id: requirement.task_id }),
    requirements: [requirement],
    ...overrides,
  };
}
