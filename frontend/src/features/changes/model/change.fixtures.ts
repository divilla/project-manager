import type { Change, ChangeDetail, ChangeReferences, Epic } from './change.types';

export function changeFixture(overrides: Partial<Change> = {}): Change {
  const change: Change = {
    id: 1,
    version: 0,
    project_id: 1,
    epic_id: null,
    change_phase: 'backlog',
    change_types: ['feature'],
    title: 'Change',
    requirement_body: '',
    requirement_html: '',
    pull_request_body: '',
    pull_request_html: '',
    pull_request_url: '',
    closed: false,
    done_tc: 0,
    total_tc: 0,
    completed: 0,
    created: '2026-01-01T00:00:00Z',
    modified: '2026-01-01T00:00:00Z',
  };
  return { ...change, ...overrides };
}

export function changeDetailFixture(overrides: Partial<ChangeDetail> = {}): ChangeDetail {
  return {
    change: changeFixture(),
    test_cases: [],
    ...overrides,
  };
}

export function epicFixture(overrides: Partial<Epic> = {}): Epic {
  const epic: Epic = {
    id: 1,
    version: 0,
    project_id: 1,
    name: 'Epic',
    done_tc: 0,
    total_tc: 0,
    completed: 0,
    change_count: 0,
    created: '2026-01-01T00:00:00Z',
    modified: '2026-01-01T00:00:00Z',
  };
  return { ...epic, ...overrides };
}

export function changeReferencesFixture(overrides: Partial<ChangeReferences> = {}): ChangeReferences {
  return {
    phases: [
      { slug: 'backlog', priority: 1 },
      { slug: 'review', priority: 2 },
    ],
    types: [
      { slug: 'change', priority: 1 },
      { slug: 'feature', priority: 2 },
    ],
    ...overrides,
  };
}
