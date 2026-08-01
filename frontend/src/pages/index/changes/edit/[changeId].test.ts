import { flushPromises, mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { listEpics } from '@/features/epics/api/epicApi';
import {
  changeDetailFixture,
  changeFixture,
  changePhasesFixture,
  changeTypesFixture,
} from '@/features/changes/model/change.fixtures';
import {
  getChange,
  getChangePhases,
  getChangeTypes,
  listChanges,
  updateChangeDefinition,
  updateChangePR,
  updateChangeSpec,
  updateChangeTitle,
} from '@/features/changes/api/changeApi';
import { createQuasarStubs } from '@/test/quasarStubs';
import ChangeEditPage from './[changeId].vue';

const routerMock = vi.hoisted(() => ({
  push: vi.fn(),
  route: {
    params: {
      changeId: '2',
    },
  },
}));

vi.mock('vue-router', () => ({
  useRoute: () => routerMock.route,
  useRouter: () => ({
    push: routerMock.push,
  }),
}));

vi.mock('@/features/changes/api/changeApi', () => ({
  getChange: vi.fn(),
  getChangePhases: vi.fn(),
  getChangeTypes: vi.fn(),
  listChanges: vi.fn(),
  updateChangeEpic: vi.fn(),
  updateChangeDefinition: vi.fn(),
  updateChangeOpen: vi.fn(),
  updateChangePR: vi.fn(),
  updateChangePRUrl: vi.fn(),
  updateChangePhase: vi.fn(),
  updateChangeSpec: vi.fn(),
  updateChangeTitle: vi.fn(),
  updateChangeTypes: vi.fn(),
}));

vi.mock('@/features/epics/api/epicApi', () => ({
  listEpics: vi.fn(),
}));

const quasarStubs = createQuasarStubs({
  QToggle: {
    emits: ['update:modelValue'],
    props: ['disable', 'label', 'modelValue'],
    template:
      '<input type="checkbox" :aria-label="label" :checked="modelValue" :disabled="disable" @change="$emit(\'update:modelValue\', $event.target.checked)" />',
  },
});

describe('ChangeEditPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setActivePinia(createPinia());
    vi.mocked(getChangePhases).mockResolvedValue(changePhasesFixture());
    vi.mocked(getChangeTypes).mockResolvedValue(changeTypesFixture());
    vi.mocked(getChange).mockResolvedValue(
      changeDetailFixture({
        change: changeFixture({
          id: 2,
          title: 'Current change',
          def: '# Current change\n\nDefinition',
          spec: '# Current spec\n\nBody',
          pr: '# Current PR\n\nBody',
          pr_url: 'https://example.test/pr/2',
        }),
      }),
    );
    vi.mocked(listChanges).mockResolvedValue([]);
    vi.mocked(listEpics).mockResolvedValue([]);
  });

  function mountPage() {
    return mount(ChangeEditPage, {
      global: {
        stubs: quasarStubs,
      },
    });
  }

  it('trims and saves a changed definition through the definition contract', async () => {
    vi.mocked(updateChangeDefinition).mockResolvedValue(
      changeFixture({
        id: 2,
        title: 'Current change',
        def: '# Rewritten definition\n\nBody',
        spec: '# Current spec\n\nBody',
        pr: '# Current PR\n\nBody',
        pr_url: 'https://example.test/pr/2',
      }),
    );
    const wrapper = mountPage();
    await flushPromises();

    await wrapper
      .get('textarea[aria-label="Definition"]')
      .setValue('  # Rewritten definition\n\nBody  ');
    await wrapper.get('form').trigger('submit');
    await flushPromises();

    expect(updateChangeDefinition).toHaveBeenCalledWith(2, '# Rewritten definition\n\nBody');
    expect(updateChangeSpec).not.toHaveBeenCalled();
    expect(updateChangePR).not.toHaveBeenCalled();
    expect(routerMock.push).toHaveBeenCalledWith('/changes/2');
  });

  it('does not save a blank definition', async () => {
    const wrapper = mountPage();
    await flushPromises();

    await wrapper.get('textarea[aria-label="Definition"]').setValue('   ');
    await wrapper.get('form').trigger('submit');
    await flushPromises();

    expect(updateChangeDefinition).not.toHaveBeenCalled();
    expect(updateChangeTitle).not.toHaveBeenCalled();
    expect(updateChangeSpec).not.toHaveBeenCalled();
    expect(updateChangePR).not.toHaveBeenCalled();
    expect(routerMock.push).not.toHaveBeenCalled();
  });

  it('shows definition update failures and stays on the edit page', async () => {
    vi.mocked(updateChangeDefinition).mockRejectedValue(new Error('Definition update failed.'));
    const wrapper = mountPage();
    await flushPromises();

    await wrapper.get('textarea[aria-label="Definition"]').setValue('Rewritten definition');
    await wrapper.get('form').trigger('submit');
    await flushPromises();

    expect(updateChangeDefinition).toHaveBeenCalledWith(2, 'Rewritten definition');
    expect(wrapper.text()).toContain('Definition update failed.');
    expect(routerMock.push).not.toHaveBeenCalled();
  });

  it('rejects empty changed artifact values before saving any field', async () => {
    const wrapper = mountPage();
    await flushPromises();

    await wrapper.get('input[aria-label="Change title"]').setValue('Renamed change');
    await wrapper.get('textarea[aria-label="Spec"]').setValue('');
    await wrapper.get('form').trigger('submit');
    await flushPromises();

    expect(wrapper.text()).toContain('Spec is required.');
    expect(updateChangeTitle).not.toHaveBeenCalled();
    expect(updateChangeSpec).not.toHaveBeenCalled();
    expect(updateChangePR).not.toHaveBeenCalled();
  });

  it.each(['Plain text spec body', '## Summary\n\nLower-level heading'])(
    'saves a changed non-empty spec without requiring a top-level heading: %s',
    async (nextSpec) => {
      vi.mocked(updateChangeSpec).mockResolvedValue(
        changeFixture({
          id: 2,
          title: 'Current change',
          def: '# Current change\n\nDefinition',
          spec: nextSpec,
          pr: '# Current PR\n\nBody',
          pr_url: 'https://example.test/pr/2',
        }),
      );
      const wrapper = mountPage();
      await flushPromises();

      await wrapper.get('textarea[aria-label="Spec"]').setValue(nextSpec);
      await wrapper.get('form').trigger('submit');
      await flushPromises();

      expect(updateChangeSpec).toHaveBeenCalledWith(2, nextSpec, false);
    },
  );

  it('saves changed plain text PR bodies', async () => {
    vi.mocked(updateChangePR).mockResolvedValue(
      changeFixture({
        id: 2,
        title: 'Current change',
        def: '# Current change\n\nDefinition',
        spec: '# Current spec\n\nBody',
        pr: 'Plain text PR body',
        pr_url: 'https://example.test/pr/2',
      }),
    );
    const wrapper = mountPage();
    await flushPromises();

    await wrapper.get('textarea[aria-label="PR body"]').setValue('Plain text PR body');
    await wrapper.get('form').trigger('submit');
    await flushPromises();

    expect(updateChangePR).toHaveBeenCalledWith(2, 'Plain text PR body', false);
    expect(updateChangeSpec).not.toHaveBeenCalled();
  });

  it('rejects empty changed PR bodies before saving any field', async () => {
    const wrapper = mountPage();
    await flushPromises();

    await wrapper.get('input[aria-label="Change title"]').setValue('Renamed change');
    await wrapper.get('textarea[aria-label="PR body"]').setValue('');
    await wrapper.get('form').trigger('submit');
    await flushPromises();

    expect(wrapper.text()).toContain('PR is required.');
    expect(updateChangeTitle).not.toHaveBeenCalled();
    expect(updateChangePR).not.toHaveBeenCalled();
  });
});
