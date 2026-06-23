import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import {
  createProject as createProjectRequest,
  deleteProject as deleteProjectRequest,
  listProjects,
  updateProject as updateProjectRequest,
} from '../api/projectApi';
import type { Project } from './project.types';

const ACTIVE_PROJECT_STORAGE_KEY = 'aipm.activeProjectId';

function readPersistedProjectId() {
  if (typeof localStorage === 'undefined') return 0;

  const value = Number(localStorage.getItem(ACTIVE_PROJECT_STORAGE_KEY));
  return Number.isInteger(value) && value > 0 ? value : 0;
}

function writePersistedProjectId(projectId: number) {
  if (typeof localStorage === 'undefined') return;

  if (projectId > 0) {
    localStorage.setItem(ACTIVE_PROJECT_STORAGE_KEY, String(projectId));
    return;
  }

  localStorage.removeItem(ACTIVE_PROJECT_STORAGE_KEY);
}

function lowestProjectId(items: Project[]) {
  return items.reduce((lowest, project) => Math.min(lowest, project.id), items[0]?.id || 0);
}

export const useProjectSelectionStore = defineStore('projectSelection', () => {
  const projects = ref<Project[]>([]);
  const activeProjectId = ref(readPersistedProjectId());
  const loading = ref(false);
  const error = ref('');
  const hasLoaded = ref(false);

  const activeProject = computed(
    () => projects.value.find((project) => project.id === activeProjectId.value) || null,
  );
  const projectOptions = computed(() =>
    projects.value.map((project) => ({ label: project.name, value: project.id })),
  );

  function setProjects(items: Project[]) {
    projects.value = items;
    validateActiveProject();
  }

  function selectProject(projectId: number) {
    activeProjectId.value = projects.value.some((project) => project.id === projectId)
      ? projectId
      : 0;
    writePersistedProjectId(activeProjectId.value);
  }

  function validateActiveProject() {
    if (!projects.value.length) {
      selectProject(0);
      return;
    }

    if (projects.value.some((project) => project.id === activeProjectId.value)) {
      writePersistedProjectId(activeProjectId.value);
      return;
    }

    const persistedProjectId = readPersistedProjectId();
    if (projects.value.some((project) => project.id === persistedProjectId)) {
      selectProject(persistedProjectId);
      return;
    }

    selectProject(lowestProjectId(projects.value));
  }

  async function loadProjects() {
    loading.value = true;
    error.value = '';

    try {
      setProjects(await listProjects());
      hasLoaded.value = true;
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Unable to load projects.';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function createProject(name: string) {
    const project = await createProjectRequest(name);
    setProjects([...projects.value, project]);
    selectProject(project.id);
    return project;
  }

  async function renameProject(id: number, name: string) {
    const project = await updateProjectRequest(id, name);
    setProjects(projects.value.map((item) => (item.id === project.id ? project : item)));
    return project;
  }

  async function removeProject(project: Project) {
    await deleteProjectRequest(project.id);
    setProjects(projects.value.filter((item) => item.id !== project.id));
  }

  return {
    projects,
    activeProjectId,
    activeProject,
    projectOptions,
    loading,
    error,
    hasLoaded,
    loadProjects,
    selectProject,
    createProject,
    renameProject,
    removeProject,
    validateActiveProject,
  };
});
