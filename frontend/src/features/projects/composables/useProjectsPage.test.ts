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
  createTask,
  deleteTask,
  getTask,
  getTaskReferences,
  listTasks,
  updateTask,
  updateTaskPhase,
} from '@/features/tasks/api/taskApi';
import {
  createRequirement,
  deleteRequirement,
  updateRequirement,
  updateRequirementDone,
} from '@/features/requirements/api/requirementApi';
import { requirementMutationFixture } from '@/features/requirements/model/requirement.fixtures';
import {
  taskDetailFixture,
  taskFixture,
  taskReferencesFixture,
} from '@/features/tasks/model/task.fixtures';
import { useProjectsPage } from './useProjectsPage';

vi.mock('@/features/projects/api/projectApi', () => ({
  createProject: vi.fn(),
  deleteProject: vi.fn(),
  listProjects: vi.fn(),
  updateProject: vi.fn(),
}));

vi.mock('@/features/tasks/api/taskApi', () => ({
  createTask: vi.fn(),
  deleteTask: vi.fn(),
  getTask: vi.fn(),
  getTaskReferences: vi.fn(),
  listTasks: vi.fn(),
  updateTask: vi.fn(),
  updateTaskPhase: vi.fn(),
}));

vi.mock('@/features/requirements/api/requirementApi', () => ({
  createRequirement: vi.fn(),
  deleteRequirement: vi.fn(),
  updateRequirement: vi.fn(),
  updateRequirementDone: vi.fn(),
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
    vi.mocked(createTask).mockResolvedValue(taskFixture({ id: 11, project_id: 1 }));
    vi.mocked(updateTask).mockResolvedValue(
      taskFixture({ id: 10, project_id: 1, name: 'Updated' }),
    );
    vi.mocked(updateTaskPhase).mockResolvedValue(
      taskFixture({ id: 10, project_id: 1, task_phase: 'review' }),
    );
    vi.mocked(deleteTask).mockResolvedValue(undefined);
    vi.mocked(getTask).mockResolvedValue(taskDetailFixture({ task: taskFixture({ id: 10 }) }));
    vi.mocked(createRequirement).mockResolvedValue(requirementMutationFixture());
    vi.mocked(updateRequirement).mockResolvedValue(requirementMutationFixture());
    vi.mocked(updateRequirementDone).mockResolvedValue(requirementMutationFixture());
    const deletedRequirementMutation = requirementMutationFixture();
    delete deletedRequirementMutation.requirement;
    vi.mocked(deleteRequirement).mockResolvedValue(deletedRequirementMutation);
  });

  afterEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it('loads references, projects, default selections, and tasks on mount', async () => {
    const { state } = mountProjectsPage();
    await flushPromises();

    expect(state.projects.value).toEqual([
      projectFixture({ id: 1, name: 'Project', task_count: 1 }),
    ]);
    expect(state.selectedProjectId.value).toBe(1);
    expect(state.taskType.value).toBe('task');
    expect(state.taskPhase.value).toBe('backlog');
    expect(state.tasks.value).toEqual([taskFixture({ id: 10, project_id: 1 })]);
    expect(listTasks).toHaveBeenCalledWith(1);
  });

  it('loads only projects when task-board behavior is disabled', async () => {
    const { state } = mountProjectsPage({ tasksEnabled: false });
    await flushPromises();

    expect(state.projects.value).toEqual([
      projectFixture({ id: 1, name: 'Project', task_count: 1 }),
    ]);
    expect(state.selectedProjectId.value).toBe(1);
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

    expect(state.selectedProjectId.value).toBe(2);
    expect(listTasks).not.toHaveBeenCalled();
  });

  it('selects a created project and loads its task list', async () => {
    const { state } = mountProjectsPage();
    await flushPromises();
    vi.mocked(listTasks).mockClear();

    state.projectName.value = 'New project';
    await state.createProjectFromForm();

    expect(createProject).toHaveBeenCalledWith('New project');
    expect(state.selectedProjectId.value).toBe(2);
    expect(listTasks).toHaveBeenCalledWith(2);
  });

  it('refreshes project task counts after task creation', async () => {
    const { state } = mountProjectsPage();
    await flushPromises();
    vi.mocked(listProjects).mockClear();
    vi.mocked(listProjects).mockResolvedValueOnce([
      projectFixture({ id: 1, name: 'Project', task_count: 2 }),
    ]);

    state.taskName.value = 'New task';
    await state.createTaskFromForm();

    expect(createTask).toHaveBeenCalledWith({
      project_id: 1,
      name: 'New task',
      task_phase: 'backlog',
      task_type: 'task',
    });
    expect(listProjects).toHaveBeenCalledTimes(1);
    expect(state.projects.value).toEqual([
      projectFixture({ id: 1, name: 'Project', task_count: 2 }),
    ]);
  });

  it('refreshes project task counts after task deletion', async () => {
    const { state } = mountProjectsPage();
    await flushPromises();
    vi.mocked(listProjects).mockClear();
    vi.mocked(listTasks).mockClear();
    vi.mocked(listProjects).mockResolvedValueOnce([
      projectFixture({ id: 1, name: 'Project', task_count: 0 }),
    ]);

    await state.removeTask(taskFixture({ id: 10, project_id: 1 }));

    expect(deleteTask).toHaveBeenCalledWith(10);
    expect(listTasks).toHaveBeenCalledWith(1);
    expect(listProjects).toHaveBeenCalledTimes(1);
    expect(state.projects.value).toEqual([
      projectFixture({ id: 1, name: 'Project', task_count: 0 }),
    ]);
  });

  it('refreshes selected project tasks after requirement creation', async () => {
    const { state } = mountProjectsPage();
    await flushPromises();
    vi.mocked(listTasks).mockClear();
    vi.mocked(listTasks).mockResolvedValueOnce([taskFixture({ id: 10, completed: 100 })]);

    state.taskDetail.value = taskDetailFixture({ task: taskFixture({ id: 10 }) });
    state.requirementDefinition.value = 'Add tests';
    await state.createRequirementFromForm();

    expect(createRequirement).toHaveBeenCalledWith(10, 'Add tests');
    expect(listTasks).toHaveBeenCalledWith(1);
    expect(state.tasks.value).toEqual([taskFixture({ id: 10, completed: 100 })]);
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
