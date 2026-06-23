import { mount, flushPromises } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { defineComponent } from 'vue';
import {
  createProject,
  deleteProject,
  listProjects,
  updateProject,
} from '@/features/projects/api/projectApi';
import { useProjectSelectionStore } from '@/features/projects/model/projectSelection.store';
import type { Project } from '@/features/projects/model/project.types';
import {
  deleteTask,
  getTaskReferences,
  listTasks,
  updateTaskPhase,
} from '@/features/tasks/api/taskApi';
import { taskFixture, taskReferencesFixture } from '@/features/tasks/model/task.fixtures';
import { useProjectsPage } from './useProjectsPage';

vi.mock('@/features/projects/api/projectApi', () => ({
  createProject: vi.fn(),
  deleteProject: vi.fn(),
  listProjects: vi.fn(),
  updateProject: vi.fn(),
}));

vi.mock('@/features/tasks/api/taskApi', () => ({
  deleteTask: vi.fn(),
  getTaskReferences: vi.fn(),
  listTasks: vi.fn(),
  updateTaskPhase: vi.fn(),
}));

type ProjectsPageState = ReturnType<typeof useProjectsPage>;

function projectFixture(project: Pick<Project, 'id' | 'name' | 'task_count'>): Project {
  return {
    created: '2026-06-23T00:00:00Z',
    modified: '2026-06-23T00:00:00Z',
    ...project,
  };
}

function mountProjectsPage(options?: Parameters<typeof useProjectsPage>[0]) {
  let state: ProjectsPageState | undefined;
  const wrapper = mount(
    defineComponent({
      setup() {
        state = useProjectsPage(options);
        return () => null;
      },
    }),
  );
  return { wrapper, state: state as ProjectsPageState };
}

describe('useProjectsPage', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    localStorage.clear();
    vi.mocked(getTaskReferences).mockResolvedValue(taskReferencesFixture());
    vi.mocked(listProjects).mockResolvedValue([
      projectFixture({ id: 1, name: 'Project', task_count: 1 }),
    ]);
    vi.mocked(listTasks).mockResolvedValue([taskFixture({ id: 10, project_id: 1 })]);
    vi.mocked(createProject).mockResolvedValue(
      projectFixture({ id: 2, name: 'New project', task_count: 0 }),
    );
    vi.mocked(updateProject).mockResolvedValue(
      projectFixture({ id: 1, name: 'Renamed', task_count: 1 }),
    );
    vi.mocked(deleteProject).mockResolvedValue(undefined);
    vi.mocked(updateTaskPhase).mockResolvedValue(
      taskFixture({ id: 10, project_id: 1, task_phase: 'review' }),
    );
    vi.mocked(deleteTask).mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it('loads references, projects, empty search fields, and tasks on mount', async () => {
    const { state } = mountProjectsPage();
    await flushPromises();

    expect(state.projects.value).toEqual([
      projectFixture({ id: 1, name: 'Project', task_count: 1 }),
    ]);
    expect(state.currentProjectId.value).toBe(1);
    expect(state.taskType.value).toBe('');
    expect(state.taskPhase.value).toBe('');
    expect(state.tasks.value).toEqual([taskFixture({ id: 10, project_id: 1 })]);
    expect(listTasks).toHaveBeenCalledWith(1);
  });

  it('filters visible task board results by search fields', async () => {
    vi.mocked(listTasks).mockResolvedValueOnce([
      taskFixture({ id: 10, name: 'Build task search', task_type: 'task', task_phase: 'backlog' }),
      taskFixture({ id: 11, name: 'Review filters', task_type: 'bug', task_phase: 'review' }),
      taskFixture({ id: 12, name: 'Build refresh', task_type: 'task', task_phase: 'review' }),
    ]);
    const { state } = mountProjectsPage();
    await flushPromises();

    state.taskName.value = 'build';
    state.taskType.value = 'task';
    state.taskPhase.value = 'review';

    expect(state.tasksByPhase.value.backlog).toEqual([]);
    expect(state.tasksByPhase.value.review).toEqual([
      taskFixture({ id: 12, name: 'Build refresh', task_type: 'task', task_phase: 'review' }),
    ]);
  });

  it('refreshes tasks when search is submitted', async () => {
    const { state } = mountProjectsPage();
    await flushPromises();
    vi.mocked(listTasks).mockClear();
    vi.mocked(listTasks).mockResolvedValueOnce([
      taskFixture({ id: 13, name: 'Fresh result', project_id: 1 }),
    ]);

    await state.searchTasks();

    expect(listTasks).toHaveBeenCalledWith(1);
    expect(state.tasks.value).toEqual([
      taskFixture({ id: 13, name: 'Fresh result', project_id: 1 }),
    ]);
  });

  it('clears search fields and refreshes tasks', async () => {
    const { state } = mountProjectsPage();
    await flushPromises();
    vi.mocked(listTasks).mockClear();
    vi.mocked(listTasks).mockResolvedValueOnce([
      taskFixture({ id: 14, name: 'Cleared result', project_id: 1 }),
    ]);
    state.taskName.value = 'build';
    state.taskType.value = 'task';
    state.taskPhase.value = 'review';

    await state.clearTaskSearch();

    expect(state.taskName.value).toBe('');
    expect(state.taskType.value).toBe('');
    expect(state.taskPhase.value).toBe('');
    expect(listTasks).toHaveBeenCalledWith(1);
    expect(state.tasks.value).toEqual([
      taskFixture({ id: 14, name: 'Cleared result', project_id: 1 }),
    ]);
  });

  it('loads only projects when task-board behavior is disabled', async () => {
    const { state } = mountProjectsPage({ tasksEnabled: false });
    await flushPromises();

    expect(state.projects.value).toEqual([
      projectFixture({ id: 1, name: 'Project', task_count: 1 }),
    ]);
    expect(state.currentProjectId.value).toBe(1);
    expect(getTaskReferences).not.toHaveBeenCalled();
    expect(listTasks).not.toHaveBeenCalled();

    await state.selectProject(1);

    expect(listTasks).not.toHaveBeenCalled();
  });

  it('does not load tasks when the shared selector changes on the projects CRUD page', async () => {
    vi.mocked(listProjects).mockResolvedValue([
      projectFixture({ id: 1, name: 'Project', task_count: 1 }),
      projectFixture({ id: 2, name: 'Other project', task_count: 0 }),
    ]);
    const { state } = mountProjectsPage({ tasksEnabled: false });
    await flushPromises();
    vi.mocked(listTasks).mockClear();

    useProjectSelectionStore().selectProject(2);
    await flushPromises();

    expect(state.currentProjectId.value).toBe(2);
    expect(listTasks).not.toHaveBeenCalled();
  });

  it('selects a created project and loads its task list', async () => {
    const { state } = mountProjectsPage();
    await flushPromises();
    vi.mocked(listTasks).mockClear();

    state.projectName.value = 'New project';
    await state.createProjectFromForm();

    expect(createProject).toHaveBeenCalledWith('New project');
    expect(state.currentProjectId.value).toBe(2);
    expect(listTasks).toHaveBeenCalledWith(2);
  });

  it('confirms project deletion before removing it', async () => {
    const { state } = mountProjectsPage({ tasksEnabled: false });
    await flushPromises();

    state.removeProject(projectFixture({ id: 2, name: 'Empty project', task_count: 0 }));

    expect(state.confirmationDialogOpen.value).toBe(true);
    expect(deleteProject).not.toHaveBeenCalled();

    await state.confirm();

    expect(state.confirmationDialogOpen.value).toBe(false);
    expect(deleteProject).toHaveBeenCalledWith(2);
  });

  it('refreshes project task counts after task deletion', async () => {
    const { state } = mountProjectsPage();
    await flushPromises();
    vi.mocked(listProjects).mockClear();
    vi.mocked(listTasks).mockClear();
    vi.mocked(listProjects).mockResolvedValueOnce([
      projectFixture({ id: 1, name: 'Project', task_count: 0 }),
    ]);

    state.removeTask(taskFixture({ id: 10, project_id: 1 }));

    expect(state.confirmationDialogOpen.value).toBe(true);
    expect(deleteTask).not.toHaveBeenCalled();

    await state.confirm();

    expect(state.confirmationDialogOpen.value).toBe(false);
    expect(deleteTask).toHaveBeenCalledWith(10);
    expect(listTasks).toHaveBeenCalledWith(1);
    expect(listProjects).toHaveBeenCalledTimes(1);
    expect(state.projects.value).toEqual([
      projectFixture({ id: 1, name: 'Project', task_count: 0 }),
    ]);
  });

  it('refreshes selected project tasks after task phase changes', async () => {
    const { state } = mountProjectsPage();
    await flushPromises();
    vi.mocked(listTasks).mockClear();
    vi.mocked(listTasks).mockResolvedValueOnce([taskFixture({ id: 10, task_phase: 'review' })]);

    await state.moveTask(taskFixture({ id: 10 }), 'review');

    expect(updateTaskPhase).toHaveBeenCalledWith(10, 'review');
    expect(listTasks).toHaveBeenCalledWith(1);
    expect(state.tasks.value[0]?.task_phase).toBe('review');
  });

  it('sets a user-facing error when task loading fails', async () => {
    const { state } = mountProjectsPage();
    await flushPromises();
    vi.mocked(listTasks).mockRejectedValueOnce(new Error('network down'));

    await state.selectProject(1);

    expect(state.error.value).toBe('network down');
  });
});
