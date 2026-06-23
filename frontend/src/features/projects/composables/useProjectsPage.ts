import { computed, onMounted, ref } from 'vue';
import { storeToRefs } from 'pinia';
import type { Project } from '@/features/projects/model/project.types';
import { useProjectSelectionStore } from '@/features/projects/model/projectSelection.store';
import { useTaskCacheStore } from '@/features/tasks/model/taskCache.store';
import {
  deleteTask,
  getTaskReferences,
  updateTaskPhase,
} from '@/features/tasks/api/taskApi';
import type { ReferenceOption, SelectOption, Task } from '@/features/tasks/model/task.types';

function errorMessage(err: unknown, fallback: string) {
  return err instanceof Error ? err.message : fallback;
}

interface UseProjectsPageOptions {
  tasksEnabled?: boolean;
}

export function useProjectsPage(options: UseProjectsPageOptions = {}) {
  const tasksEnabled = options.tasksEnabled ?? true;
  const projectSelection = useProjectSelectionStore();
  const taskCache = useTaskCacheStore();
  const { projects, currentProjectId, currentProject } = storeToRefs(projectSelection);
  const { tasks } = storeToRefs(taskCache);
  const phases = ref<ReferenceOption[]>([]);
  const types = ref<ReferenceOption[]>([]);
  const projectName = ref('');
  const taskName = ref('');
  const taskType = ref('');
  const taskPhase = ref('');
  const loading = ref(false);
  const error = ref('');

  const projectDialogOpen = ref(false);
  const projectEditId = ref(0);
  const projectEditName = ref('');

  const confirmationDialogOpen = ref(false);
  let confirmationAction: (() => Promise<void>) | null = null;

  const phaseOptions = computed<SelectOption[]>(() =>
    phases.value.map((phase) => ({ label: phase.slug, value: phase.slug })),
  );
  const typeOptions = computed<SelectOption[]>(() =>
    types.value.map((type) => ({ label: type.slug, value: type.slug })),
  );
  const boardPhases = computed(() => (phases.value.length ? phases.value : uniqueTaskPhases.value));

  const uniqueTaskPhases = computed<ReferenceOption[]>(() =>
    [...new Set(tasks.value.map((task) => task.task_phase))].map((slug, index) => ({
      slug,
      priority: index,
    })),
  );

  const filteredTasks = computed<Task[]>(() => {
    const name = taskName.value.trim().toLowerCase();
    const type = taskType.value;
    const phase = taskPhase.value;

    return tasks.value.filter((task) => {
      if (name && !task.name.toLowerCase().includes(name)) return false;
      if (type && task.task_type !== type) return false;
      if (phase && task.task_phase !== phase) return false;
      return true;
    });
  });

  const tasksByPhase = computed<Record<string, Task[]>>(() => {
    const grouped: Record<string, Task[]> = {};
    for (const phase of boardPhases.value) grouped[phase.slug] = [];
    for (const task of filteredTasks.value) {
      const group = grouped[task.task_phase] || [];
      group.push(task);
      grouped[task.task_phase] = group;
    }
    return grouped;
  });

  async function loadAll() {
    loading.value = true;
    error.value = '';

    try {
      if (tasksEnabled) {
        const [references] = await Promise.all([
          getTaskReferences(),
          projectSelection.loadProjects(),
        ]);
        phases.value = references.phases;
        types.value = references.types;

        if (currentProjectId.value) {
          await loadTasks(currentProjectId.value);
        } else {
          taskCache.setTasks([]);
        }
      } else {
        await projectSelection.loadProjects();
        taskCache.setTasks([]);
      }
    } catch (err) {
      error.value = errorMessage(err, 'Unable to load projects.');
    } finally {
      loading.value = false;
    }
  }

  async function loadTasks(projectId: number) {
    await taskCache.loadProjectTasks(projectId);
  }

  async function selectProject(projectId: number) {
    projectSelection.selectProject(projectId);
    error.value = '';
    try {
      if (!tasksEnabled) {
        taskCache.setTasks([]);
      } else if (currentProjectId.value) {
        await loadTasks(currentProjectId.value);
      } else {
        taskCache.setTasks([]);
      }
    } catch (err) {
      error.value = errorMessage(err, 'Unable to load tasks.');
    }
  }

  async function createProjectFromForm() {
    const name = projectName.value.trim();
    if (!name) return;

    try {
      const project = await projectSelection.createProject(name);
      projectName.value = '';
      if (tasksEnabled) await loadTasks(project.id);
    } catch (err) {
      error.value = errorMessage(err, 'Unable to create project.');
    }
  }

  function startProjectRename(project: Project) {
    projectEditId.value = project.id;
    projectEditName.value = project.name;
    projectDialogOpen.value = true;
  }

  async function saveProjectName() {
    const name = projectEditName.value.trim();
    if (!projectEditId.value || !name) return;

    try {
      await projectSelection.renameProject(projectEditId.value, name);
      projectDialogOpen.value = false;
    } catch (err) {
      error.value = errorMessage(err, 'Unable to update project.');
    }
  }

  function requestConfirmation(action: () => Promise<void>) {
    confirmationAction = action;
    confirmationDialogOpen.value = true;
  }

  async function confirm() {
    const action = confirmationAction;
    if (!action) return;

    confirmationAction = null;
    confirmationDialogOpen.value = false;
    await action();
  }

  function removeProject(project: Project) {
    if (project.task_count > 0) {
      error.value = 'Delete all project tasks before deleting this project.';
      return;
    }

    requestConfirmation(() => removeProjectConfirmed(project));
  }

  async function removeProjectConfirmed(project: Project) {
    try {
      const wasSelected = currentProjectId.value === project.id;
      await projectSelection.removeProject(project);
      if (tasksEnabled && wasSelected) {
        taskCache.setTasks([]);
        if (currentProjectId.value) await loadTasks(currentProjectId.value);
      }
    } catch (err) {
      error.value = errorMessage(err, 'Unable to delete project.');
    }
  }

  async function searchTasks() {
    await loadAll();
  }

  async function clearTaskSearch() {
    taskName.value = '';
    taskType.value = '';
    taskPhase.value = '';
    await loadAll();
  }

  async function moveTask(task: Task, phase: string) {
    try {
      const moved = await updateTaskPhase(task.id, phase);
      taskCache.upsertTask(moved);
      await refreshCurrentProjectTasks();
    } catch (err) {
      error.value = errorMessage(err, 'Unable to move task.');
    }
  }

  async function refreshCurrentProjectTasks() {
    if (currentProjectId.value) {
      await loadTasks(currentProjectId.value);
    }
  }

  function removeTask(task: Task) {
    requestConfirmation(() => removeTaskConfirmed(task));
  }

  async function removeTaskConfirmed(task: Task) {
    try {
      await deleteTask(task.id);
      await refreshCurrentProjectTasks();
      await projectSelection.loadProjects();
    } catch (err) {
      error.value = errorMessage(err, 'Unable to delete task.');
    }
  }

  onMounted(() => {
    void loadAll();
  });

  return {
    projects,
    tasks,
    phases,
    types,
    currentProjectId,
    projectName,
    taskName,
    taskType,
    taskPhase,
    loading,
    error,
    projectDialogOpen,
    projectEditId,
    projectEditName,
    confirmationDialogOpen,
    currentProject,
    phaseOptions,
    typeOptions,
    boardPhases,
    tasksByPhase,
    loadAll,
    selectProject,
    createProjectFromForm,
    startProjectRename,
    saveProjectName,
    removeProject,
    searchTasks,
    clearTaskSearch,
    moveTask,
    removeTask,
    confirm,
  };
}
