import { mount } from '@vue/test-utils';
import ProjectList from './ProjectList.vue';

const quasarStubs = {
  QBadge: { template: '<span><slot /></span>' },
  QBtn: {
    emits: ['click'],
    props: ['icon'],
    template: '<button :data-icon="icon" @click="$emit(\'click\', $event)"><slot /></button>',
  },
  QIcon: { template: '<span><slot /></span>' },
  QItem: {
    emits: ['click'],
    props: ['active'],
    template: '<li :data-active="active" @click="$emit(\'click\', $event)"><slot /></li>',
  },
  QItemLabel: { template: '<span><slot /></span>' },
  QItemSection: { template: '<div><slot /></div>' },
  QList: { template: '<ul><slot /></ul>' },
  QTooltip: { template: '<span><slot /></span>' },
};

describe('ProjectList', () => {
  it('emits select, rename, and delete events with the expected project', async () => {
    const project = { id: 1, name: 'Project' };
    const wrapper = mount(ProjectList, {
      props: {
        projects: [project],
        selectedProjectId: 1,
        loading: false,
      },
      global: {
        stubs: quasarStubs,
      },
    });

    await wrapper.find('li').trigger('click');
    await wrapper.find('[data-icon="edit"]').trigger('click');
    await wrapper.find('[data-icon="delete"]').trigger('click');

    expect(wrapper.emitted('select')).toEqual([[1]]);
    expect(wrapper.emitted('rename')).toEqual([[project]]);
    expect(wrapper.emitted('delete')).toEqual([[project]]);
  });
});
