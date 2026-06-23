import { mount, flushPromises } from '@vue/test-utils';
import { defineComponent } from 'vue';
import {
  createProject,
  deleteProject,
  listProjects,
  updateProject,
} from '@/features/projects/api/projectApi';
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
import { taskDetailFixture, taskFixture, taskReferencesFixture } from '@/features/tasks/model/task.fixtures';
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

function mountProjectsPage() {
  let state: ProjectsPageState | undefined;
  const wrapper = mount(
    defineComponent({
      setup() {
        state = useProjectsPage();
        return () => null;
      },
    }),
  );
  return { wrapper, state: state as ProjectsPageState };
}

describe('useProjectsPage', () => {
  beforeEach(() => {
    vi.mocked(getTaskReferences).mockResolvedValue(taskReferencesFixture());
    vi.mocked(listProjects).mockResolvedValue([{ id: 1, name: 'Project' }]);
    vi.mocked(listTasks).mockResolvedValue([taskFixture({ id: 10, project_id: 1 })]);
    vi.mocked(createProject).mockResolvedValue({ id: 2, name: 'New project' });
    vi.mocked(updateProject).mockResolvedValue({ id: 1, name: 'Renamed' });
    vi.mocked(deleteProject).mockResolvedValue(undefined);
    vi.mocked(createTask).mockResolvedValue(taskFixture({ id: 11, project_id: 1 }));
    vi.mocked(updateTask).mockResolvedValue(taskFixture({ id: 10, project_id: 1, name: 'Updated' }));
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
  });

  it('loads references, projects, default selections, and tasks on mount', async () => {
    const { state } = mountProjectsPage();
    await flushPromises();

    expect(state.projects.value).toEqual([{ id: 1, name: 'Project' }]);
    expect(state.selectedProjectId.value).toBe(1);
    expect(state.taskType.value).toBe('task');
    expect(state.taskPhase.value).toBe('backlog');
    expect(state.tasks.value).toEqual([taskFixture({ id: 10, project_id: 1 })]);
    expect(listTasks).toHaveBeenCalledWith(1);
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
