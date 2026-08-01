import { createPinia, setActivePinia } from 'pinia';
import { listChanges } from '@/features/changes/api/changeApi';
import { listEpics } from '@/features/epics/api/epicApi';
import { changeFixture, epicFixture } from './change.fixtures';
import type { ChangeListItem, Epic } from './change.types';
import { useChangeCacheStore } from './changeCache.store';

vi.mock('@/features/changes/api/changeApi', () => ({
  listChanges: vi.fn(),
}));

vi.mock('@/features/epics/api/epicApi', () => ({
  listEpics: vi.fn(),
}));

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((nextResolve) => {
    resolve = nextResolve;
  });
  return { promise, resolve };
}

function mockProjectLoads() {
  const firstChanges = deferred<ChangeListItem[]>();
  const firstEpics = deferred<Epic[]>();
  const secondChanges = deferred<ChangeListItem[]>();
  const secondEpics = deferred<Epic[]>();

  vi.mocked(listChanges).mockImplementation((projectId) =>
    projectId === 1 ? firstChanges.promise : secondChanges.promise,
  );
  vi.mocked(listEpics).mockImplementation((projectId) =>
    projectId === 1 ? firstEpics.promise : secondEpics.promise,
  );

  return { firstChanges, firstEpics, secondChanges, secondEpics };
}

describe('changeCache store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('keeps loading the newest project when an older request resolves first', async () => {
    const requests = mockProjectLoads();
    const store = useChangeCacheStore();
    const firstProjectChanges = [changeFixture({ id: 1, project_id: 1 })];
    const secondProjectChanges = [changeFixture({ id: 2, project_id: 2 })];
    const secondProjectEpics = [epicFixture({ id: 2, project_id: 2 })];

    const firstLoad = store.loadProjectChanges(1);
    const secondLoad = store.loadProjectChanges(2);

    requests.firstChanges.resolve(firstProjectChanges);
    requests.firstEpics.resolve([epicFixture({ id: 1, project_id: 1 })]);

    await expect(firstLoad).resolves.toEqual(firstProjectChanges);
    expect(store.projectId).toBe(0);
    expect(store.loading).toBe(true);

    requests.secondChanges.resolve(secondProjectChanges);
    requests.secondEpics.resolve(secondProjectEpics);

    await expect(secondLoad).resolves.toEqual(secondProjectChanges);
    expect(store.changes).toEqual(secondProjectChanges);
    expect(store.epics).toEqual(secondProjectEpics);
    expect(store.projectId).toBe(2);
    expect(store.loading).toBe(false);
  });

  it('does not let an older request overwrite a newer completed load', async () => {
    const requests = mockProjectLoads();
    const store = useChangeCacheStore();
    const firstProjectChanges = [changeFixture({ id: 1, project_id: 1 })];
    const secondProjectChanges = [changeFixture({ id: 2, project_id: 2 })];
    const secondProjectEpics = [epicFixture({ id: 2, project_id: 2 })];

    const firstLoad = store.loadProjectChanges(1);
    const secondLoad = store.loadProjectChanges(2);

    requests.secondChanges.resolve(secondProjectChanges);
    requests.secondEpics.resolve(secondProjectEpics);

    await expect(secondLoad).resolves.toEqual(secondProjectChanges);
    expect(store.loading).toBe(false);

    requests.firstChanges.resolve(firstProjectChanges);
    requests.firstEpics.resolve([epicFixture({ id: 1, project_id: 1 })]);

    await expect(firstLoad).resolves.toEqual(firstProjectChanges);
    expect(store.changes).toEqual(secondProjectChanges);
    expect(store.epics).toEqual(secondProjectEpics);
    expect(store.projectId).toBe(2);
    expect(store.loading).toBe(false);
  });
});
