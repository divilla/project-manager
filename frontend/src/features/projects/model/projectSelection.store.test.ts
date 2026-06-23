import { createPinia, setActivePinia } from 'pinia';
import {
  createProject,
  deleteProject,
  listProjects,
  updateProject,
} from '@/features/projects/api/projectApi';
import type { Project } from './project.types';
import { useProjectSelectionStore } from './projectSelection.store';

vi.mock('@/features/projects/api/projectApi', () => ({
  createProject: vi.fn(),
  deleteProject: vi.fn(),
  listProjects: vi.fn(),
  updateProject: vi.fn(),
}));

function projectFixture(project: Pick<Project, 'id' | 'name' | 'task_count'>): Project {
  return {
    created: '2026-06-23T00:00:00Z',
    modified: '2026-06-23T00:00:00Z',
    ...project,
  };
}

describe('projectSelection store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    localStorage.clear();
    vi.mocked(createProject).mockResolvedValue(
      projectFixture({ id: 9, name: 'Created', task_count: 0 }),
    );
    vi.mocked(updateProject).mockResolvedValue(
      projectFixture({ id: 5, name: 'Renamed', task_count: 1 }),
    );
    vi.mocked(deleteProject).mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it('selects the persisted project when it still exists', async () => {
    localStorage.setItem('aipm.activeProjectId', '5');
    vi.mocked(listProjects).mockResolvedValue([
      projectFixture({ id: 10, name: 'Later', task_count: 0 }),
      projectFixture({ id: 5, name: 'Persisted', task_count: 0 }),
    ]);

    const store = useProjectSelectionStore();
    await store.loadProjects();

    expect(store.activeProjectId).toBe(5);
    expect(store.activeProject?.name).toBe('Persisted');
  });

  it('loads the complete project list in one request', async () => {
    const project = projectFixture({ id: 1, name: 'Project', task_count: 0 });
    vi.mocked(listProjects).mockResolvedValue([project]);

    const store = useProjectSelectionStore();
    await store.loadProjects();

    expect(listProjects).toHaveBeenCalledWith();
    expect(store.projects).toEqual([project]);
  });

  it('repairs invalid persisted selections by choosing the lowest project id', async () => {
    localStorage.setItem('aipm.activeProjectId', '99');
    vi.mocked(listProjects).mockResolvedValue([
      projectFixture({ id: 10, name: 'Later', task_count: 0 }),
      projectFixture({ id: 5, name: 'First', task_count: 0 }),
    ]);

    const store = useProjectSelectionStore();
    await store.loadProjects();

    expect(store.activeProjectId).toBe(5);
    expect(localStorage.getItem('aipm.activeProjectId')).toBe('5');
  });

  it('clears selection when no projects exist', async () => {
    localStorage.setItem('aipm.activeProjectId', '5');
    vi.mocked(listProjects).mockResolvedValue([]);

    const store = useProjectSelectionStore();
    await store.loadProjects();

    expect(store.activeProjectId).toBe(0);
    expect(store.activeProject).toBeNull();
    expect(localStorage.getItem('aipm.activeProjectId')).toBeNull();
  });

  it('selects a created project', async () => {
    vi.mocked(listProjects).mockResolvedValue([]);
    const store = useProjectSelectionStore();
    await store.loadProjects();

    await store.createProject('Created');

    expect(createProject).toHaveBeenCalledWith('Created');
    expect(store.activeProjectId).toBe(9);
  });

  it('falls back to the lowest remaining id after deleting the active project', async () => {
    vi.mocked(listProjects).mockResolvedValue([
      projectFixture({ id: 8, name: 'Second', task_count: 0 }),
      projectFixture({ id: 3, name: 'First', task_count: 0 }),
    ]);

    const store = useProjectSelectionStore();
    await store.loadProjects();
    store.selectProject(8);
    await store.removeProject(projectFixture({ id: 8, name: 'Second', task_count: 0 }));

    expect(deleteProject).toHaveBeenCalledWith(8);
    expect(store.activeProjectId).toBe(3);
  });
});
