import { mount } from '@vue/test-utils';
import { taskFixture } from '../model/task.fixtures';
import TaskBoard from './TaskBoard.vue';

const quasarStubs = {
  QBadge: { props: ['label'], template: '<span>{{ label }}<slot /></span>' },
  QBtn: {
    emits: ['click'],
    props: ['icon'],
    template: '<button :data-icon="icon" @click="$emit(\'click\', $event)"><slot /></button>',
  },
  QCard: {
    emits: ['click'],
    template: '<article class="task-card-stub" @click="$emit(\'click\', $event)"><slot /></article>',
  },
  QCardActions: { template: '<div><slot /></div>' },
  QCardSection: { template: '<section><slot /></section>' },
  QIcon: { props: ['name'], template: '<span :data-icon="name"><slot /></span>' },
  QLinearProgress: { template: '<div />' },
  QSelect: {
    emits: ['update:modelValue'],
    props: ['modelValue'],
    template:
      '<button class="phase-select" @click="$emit(\'update:modelValue\', \'review\')">{{ modelValue }}</button>',
  },
  QTooltip: { template: '<span><slot /></span>' },
};

describe('TaskBoard', () => {
  it('renders tasks by phase and emits task actions', async () => {
    const backlogTask = taskFixture({ id: 1, name: 'Backlog task', task_phase: 'backlog' });
    const reviewTask = taskFixture({ id: 2, name: 'Review task', task_phase: 'review' });
    const wrapper = mount(TaskBoard, {
      props: {
        hasSelectedProject: true,
        boardPhases: [
          { slug: 'backlog', priority: 1 },
          { slug: 'review', priority: 2 },
        ],
        tasksByPhase: {
          backlog: [backlogTask],
          review: [reviewTask],
        },
        phaseOptions: [
          { label: 'backlog', value: 'backlog' },
          { label: 'review', value: 'review' },
        ],
      },
      global: {
        stubs: quasarStubs,
      },
    });

    expect(wrapper.text()).toContain('Backlog task');
    expect(wrapper.text()).toContain('Review task');

    await wrapper.find('.task-card-stub').trigger('click');
    await wrapper.find('.phase-select').trigger('click');
    await wrapper.find('[data-icon="delete"]').trigger('click');

    expect(wrapper.emitted('open-task')).toEqual([[backlogTask]]);
    expect(wrapper.emitted('move-task')).toEqual([[backlogTask, 'review']]);
    expect(wrapper.emitted('delete-task')).toEqual([[backlogTask]]);
  });
});
