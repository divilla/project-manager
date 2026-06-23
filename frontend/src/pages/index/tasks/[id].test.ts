import { flushPromises, mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { listProjects } from '@/features/projects/api/projectApi';
import type { Project } from '@/features/projects/model/project.types';
import { useProjectSelectionStore } from '@/features/projects/model/projectSelection.store';
import { deleteRequirement } from '@/features/requirements/api/requirementApi';
import {
  requirementFixture,
  requirementMutationFixture,
} from '@/features/requirements/model/requirement.fixtures';
import {
  deleteTask,
  getRenderedTaskDescriptions,
  getTask,
  listTasks,
} from '@/features/tasks/api/taskApi';
import { taskDetailFixture, taskFixture } from '@/features/tasks/model/task.fixtures';
import TaskDetailPage from './[id].vue';

const routerMock = vi.hoisted(() => ({
  back: vi.fn(),
  push: vi.fn(),
  replace: vi.fn(),
  route: {
    fullPath: '/tasks/2',
    params: {
      id: '2',
    },
  },
}));

vi.mock('vue-router', () => ({
  useRoute: () => routerMock.route,
  useRouter: () => ({
    back: routerMock.back,
    push: routerMock.push,
    replace: routerMock.replace,
  }),
}));

vi.mock('@/features/projects/api/projectApi', () => ({
  listProjects: vi.fn(),
}));

vi.mock('@/features/tasks/api/taskApi', () => ({
  deleteTask: vi.fn(),
  getRenderedTaskDescriptions: vi.fn(),
  getTask: vi.fn(),
  listTasks: vi.fn(),
}));

vi.mock('@/features/requirements/api/requirementApi', () => ({
  createRequirement: vi.fn(),
  deleteRequirement: vi.fn(),
  updateRequirement: vi.fn(),
  updateRequirementDone: vi.fn(),
}));

function projectFixture(project: Pick<Project, 'id' | 'name' | 'task_count'>): Project {
  return {
    created: '2026-06-23T00:00:00Z',
    modified: '2026-06-23T00:00:00Z',
    ...project,
  };
}

const quasarStubs = {
  QBanner: { template: '<div><slot name="avatar" /><slot /></div>' },
  QBtn: {
    emits: ['click'],
    props: ['color', 'disable', 'flat', 'icon', 'label', 'loading'],
    template: `
      <button
        :data-color="color"
        :data-icon="icon"
        :disabled="disable || loading"
        type="button"
        @click="$emit('click', $event)"
      >
        {{ label }}<slot />
      </button>
    `,
  },
  QBtnDropdown: {
    emits: ['click'],
    props: ['dropdownIcon'],
    template: `
      <div>
        <button
          :data-icon="dropdownIcon"
          type="button"
          @click="$emit('click', $event)"
        />
        <slot />
      </div>
    `,
  },
  QCard: { template: '<section><slot /></section>' },
  QCardActions: { template: '<footer><slot /></footer>' },
  QCardSection: { template: '<section><slot /></section>' },
  QCheckbox: {
    emits: ['update:modelValue'],
    props: ['modelValue'],
    template: '<input type="checkbox" :checked="modelValue" @change="$emit(\'update:modelValue\', true)" />',
  },
  QDialog: {
    props: ['modelValue', 'persistent'],
    template: `
      <div
        v-if="modelValue"
        :data-persistent="persistent !== false && persistent !== undefined ? 'true' : undefined"
      >
        <slot />
      </div>
    `,
  },
  QForm: { template: '<form @submit.prevent="$emit(\'submit\', $event)"><slot /></form>' },
  QIcon: { props: ['name'], template: '<span :data-icon="name"><slot /></span>' },
  QInput: { template: '<input />' },
  QItem: {
    emits: ['click'],
    props: ['disable'],
    template: '<button type="button" :disabled="disable" @click="$emit(\'click\', $event)"><slot /></button>',
  },
  QItemLabel: { template: '<span><slot /></span>' },
  QItemSection: { template: '<span><slot /></span>' },
  QList: { template: '<div><slot /></div>' },
  QMarkupTable: { template: '<table><slot /></table>' },
  QPage: { template: '<main><slot /></main>' },
  QSpinner: { template: '<span />' },
};

describe('TaskDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setActivePinia(createPinia());
    localStorage.clear();
    routerMock.route.params.id = '2';
    routerMock.route.fullPath = '/tasks/2';
    vi.mocked(listProjects).mockResolvedValue([
      projectFixture({ id: 1, name: 'Project', task_count: 3 }),
    ]);
    vi.mocked(getTask).mockResolvedValue(
      taskDetailFixture({
        task: taskFixture({ id: 2, name: 'Current task', project_id: 1, parent_id: 1 }),
        requirements: [],
      }),
    );
    vi.mocked(listTasks).mockResolvedValue([
      taskFixture({ id: 1, name: 'Parent task', project_id: 1 }),
      taskFixture({ id: 2, name: 'Current task', project_id: 1, parent_id: 1 }),
      taskFixture({ id: 3, name: 'Leaf child', project_id: 1, parent_id: 2 }),
    ]);
    vi.mocked(getRenderedTaskDescriptions).mockResolvedValue({ descriptions: [] });
    vi.mocked(deleteTask).mockResolvedValue(undefined);
    vi.mocked(deleteRequirement).mockResolvedValue(requirementMutationFixture());
  });

  afterEach(() => {
    vi.restoreAllMocks();
    localStorage.clear();
  });

  function mountPage() {
    return mount(TaskDetailPage, {
      global: {
        directives: {
          ClosePopup: {},
        },
        stubs: quasarStubs,
      },
    });
  }

  it('renders the actions menu for the current task row and child task rows', async () => {
    const wrapper = mountPage();
    await flushPromises();

    expect(wrapper.text()).not.toContain('Actions');
    expect(wrapper.findAll('[data-icon="more_vert"]')).toHaveLength(2);
    expect(wrapper.findAll('[data-action="edit-task"]')).toHaveLength(2);
    expect(wrapper.findAll('[data-action="delete-task"]')).toHaveLength(2);
  });

  it('asks before switching project context for a pasted task URL', async () => {
    vi.mocked(listProjects).mockResolvedValue([
      projectFixture({ id: 1, name: 'Project A', task_count: 1 }),
      projectFixture({ id: 2, name: 'Project B', task_count: 1 }),
    ]);
    vi.mocked(getTask).mockResolvedValue(
      taskDetailFixture({
        task: taskFixture({ id: 2, name: 'Current task', project_id: 2 }),
        requirements: [],
      }),
    );
    vi.mocked(listTasks).mockResolvedValue([
      taskFixture({ id: 2, name: 'Current task', project_id: 2 }),
    ]);
    const wrapper = mountPage();
    await flushPromises();

    expect(wrapper.text()).toContain('Switch project?');
    expect(wrapper.text()).toContain('Project A');
    expect(wrapper.text()).toContain('Project B');

    const switchButton = wrapper.findAll('button').find((button) => button.text() === 'Switch');
    await switchButton?.trigger('click');

    const projectSelection = useProjectSelectionStore();
    expect(projectSelection.currentProjectId).toBe(2);
    expect(projectSelection.routeDrivenTargetPath).toBe('/tasks/2');
  });

  it('renders parent descriptions from the rendered descriptions endpoint', async () => {
    vi.mocked(listTasks).mockResolvedValue([
      taskFixture({
        id: 1,
        name: 'Parent task',
        project_id: 1,
        description: '**Parent**',
        description_html: '',
      }),
      taskFixture({ id: 2, name: 'Current task', project_id: 1, parent_id: 1 }),
    ]);
    vi.mocked(getRenderedTaskDescriptions).mockResolvedValue({
      descriptions: [{ id: 1, description_html: '<p><strong>Parent</strong></p>' }],
    });
    const wrapper = mountPage();
    await flushPromises();

    expect(getRenderedTaskDescriptions).toHaveBeenCalledWith([1]);
    expect(wrapper.html()).toContain('<strong>Parent</strong>');
  });

  it('opens the task edit route from the menu action', async () => {
    const wrapper = mountPage();
    await flushPromises();

    await wrapper.find('[data-action="edit-task"]').trigger('click');

    expect(routerMock.push).toHaveBeenCalledWith('/tasks/edit/2');
  });

  it('disables delete for the current task when it has children', async () => {
    const wrapper = mountPage();
    await flushPromises();

    expect(wrapper.find('[data-action="delete-task"]').attributes('disabled')).toBeDefined();
  });

  it('confirms current task deletion when it is a leaf task', async () => {
    vi.mocked(listTasks).mockResolvedValue([
      taskFixture({ id: 1, name: 'Parent task', project_id: 1 }),
      taskFixture({ id: 2, name: 'Current task', project_id: 1, parent_id: 1 }),
    ]);
    const wrapper = mountPage();
    await flushPromises();

    await wrapper.find('[data-action="delete-task"]').trigger('click');
    await flushPromises();

    expect(wrapper.text()).toContain('Are you sure?');
    expect(wrapper.text()).toContain('Cancel');
    expect(wrapper.find('[data-persistent="true"]').exists()).toBe(true);
    expect(deleteTask).not.toHaveBeenCalled();

    const confirmButton = wrapper.findAll('button').find((button) => button.text() === 'OK');
    expect(confirmButton).toBeDefined();
    await confirmButton?.trigger('click');
    await flushPromises();

    expect(deleteTask).toHaveBeenCalledWith(2);
    expect(listTasks).toHaveBeenCalledWith(1);
  });

  it('does not delete a leaf task before confirmation', async () => {
    vi.mocked(listTasks).mockResolvedValue([
      taskFixture({ id: 1, name: 'Parent task', project_id: 1 }),
      taskFixture({ id: 2, name: 'Current task', project_id: 1, parent_id: 1 }),
    ]);
    const wrapper = mountPage();
    await flushPromises();

    await wrapper.find('[data-action="delete-task"]').trigger('click');

    expect(deleteTask).not.toHaveBeenCalled();
  });

  it('renders requirement actions in the last-column menu and opens the edit dialog', async () => {
    vi.mocked(getTask).mockResolvedValue(
      taskDetailFixture({
        task: taskFixture({ id: 2, name: 'Current task', project_id: 1, parent_id: 1 }),
        requirements: [requirementFixture({ id: 9, task_id: 2, definition: 'Existing req' })],
      }),
    );
    const wrapper = mountPage();
    await flushPromises();

    expect(wrapper.findAll('[data-action="edit-requirement"]')).toHaveLength(1);
    expect(wrapper.findAll('[data-action="delete-requirement"]')).toHaveLength(1);

    await wrapper.find('[data-action="edit-requirement"]').trigger('click');

    expect(wrapper.text()).toContain('Edit Requirement');
  });

  it('confirms requirement deletion with the shared delete dialog', async () => {
    const requirement = requirementFixture({ id: 9, task_id: 2, definition: 'Delete me' });
    vi.mocked(getTask).mockResolvedValue(
      taskDetailFixture({
        task: taskFixture({ id: 2, name: 'Current task', project_id: 1, parent_id: 1 }),
        requirements: [requirement],
      }),
    );
    vi.mocked(deleteRequirement).mockResolvedValue(
      requirementMutationFixture({
        task: taskFixture({ id: 2, project_id: 1 }),
        requirements: [],
      }),
    );
    const wrapper = mountPage();
    await flushPromises();

    await wrapper.find('[data-action="delete-requirement"]').trigger('click');
    await flushPromises();

    expect(wrapper.text()).toContain('Are you sure?');
    expect(wrapper.text()).toContain('Cancel');
    expect(wrapper.find('[data-persistent="true"]').exists()).toBe(true);
    expect(deleteRequirement).not.toHaveBeenCalled();

    const okButton = wrapper.findAll('button').find((button) => button.text() === 'OK');
    expect(okButton?.attributes('data-color')).toBe('negative');
    await okButton?.trigger('click');
    await flushPromises();

    expect(deleteRequirement).toHaveBeenCalledWith(9);
    expect(listTasks).toHaveBeenCalledWith(1);
  });
});
