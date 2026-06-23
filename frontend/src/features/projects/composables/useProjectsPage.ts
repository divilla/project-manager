import { computed, onMounted, ref, watch } from 'vue';
import { storeToRefs } from 'pinia';
import type { Project } from '@/features/projects/model/project.types';
import { useProjectSelectionStore } from '@/features/projects/model/projectSelection.store';
import {
  createTask,
  deleteTask,
  getTask,
  getTaskReferences,
  listTasks,
  updateTask,
  updateTaskPhase,
} from '@/features/tasks/api/taskApi';
import type {
  ReferenceOption,
  SelectOption,
  Task,
  TaskCreateInput,
  TaskDetail,
} from '@/features/tasks/model/task.types';
import {
  createRequirement,
  deleteRequirement,
  updateRequirement,
  updateRequirementDone,
} from '@/features/requirements/api/requirementApi';
import type {
  Requirement,
  RequirementMutation,
} from '@/features/requirements/model/requirement.types';

function errorMessage(err: unknown, fallback: string) {
  return err instanceof Error ? err.message : fallback;
}

interface UseProjectsPageOptions {
  tasksEnabled?: boolean;
}

export function useProjectsPage(options: UseProjectsPageOptions = {}) {
  const tasksEnabled = options.tasksEnabled ?? true;
  const projectSelection = useProjectSelectionStore();
  const {
    projects,
    activeProjectId: selectedProjectId,
    activeProject: selectedProject,
  } = storeToRefs(projectSelection);
  const tasks = ref<Task[]>([]);
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

  const taskDialogOpen = ref(false);
  const taskDetail = ref<TaskDetail | null>(null);
  const taskEditName = ref('');
  const taskEditDescription = ref('');
  const taskEditType = ref('');
  const requirementDefinition = ref('');
  const editingRequirementId = ref(0);
  const editingRequirementDefinition = ref('');
  let suppressProjectWatch = false;

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

  const tasksByPhase = computed<Record<string, Task[]>>(() => {
    const grouped: Record<string, Task[]> = {};
    for (const phase of boardPhases.value) grouped[phase.slug] = [];
    for (const task of tasks.value) {
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
      suppressProjectWatch = true;
      if (tasksEnabled) {
        const [references] = await Promise.all([
          getTaskReferences(),
          projectSelection.loadProjects(),
        ]);
        phases.value = references.phases;
        types.value = references.types;
        if (!taskType.value) taskType.value = references.types[0]?.slug || '';
        if (!taskPhase.value) taskPhase.value = references.phases[0]?.slug || '';

        if (selectedProjectId.value) {
          await loadTasks(selectedProjectId.value);
        } else {
          tasks.value = [];
        }
      } else {
        await projectSelection.loadProjects();
        tasks.value = [];
      }
    } catch (err) {
      error.value = errorMessage(err, 'Unable to load projects.');
    } finally {
      suppressProjectWatch = false;
      loading.value = false;
    }
  }

  async function loadTasks(projectId: number) {
    tasks.value = await listTasks(projectId);
  }

  async function selectProject(projectId: number) {
    suppressProjectWatch = true;
    projectSelection.selectProject(projectId);
    error.value = '';
    try {
      if (!tasksEnabled) {
        tasks.value = [];
      } else if (selectedProjectId.value) {
        await loadTasks(selectedProjectId.value);
      } else {
        tasks.value = [];
      }
    } catch (err) {
      error.value = errorMessage(err, 'Unable to load tasks.');
    } finally {
      suppressProjectWatch = false;
    }
  }

  async function createProjectFromForm() {
    const name = projectName.value.trim();
    if (!name) return;

    try {
      suppressProjectWatch = true;
      const project = await projectSelection.createProject(name);
      projectName.value = '';
      if (tasksEnabled) await loadTasks(project.id);
    } catch (err) {
      error.value = errorMessage(err, 'Unable to create project.');
    } finally {
      suppressProjectWatch = false;
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

  async function removeProject(project: Project) {
    if (project.task_count > 0) {
      error.value = 'Delete all project tasks before deleting this project.';
      return;
    }

    if (
      typeof window !== 'undefined' &&
      !window.confirm(`Delete project "${project.name}"? This cannot be undone.`)
    ) {
      return;
    }

    try {
      const wasSelected = selectedProjectId.value === project.id;
      suppressProjectWatch = true;
      await projectSelection.removeProject(project);
      if (tasksEnabled && wasSelected) {
        tasks.value = [];
        if (selectedProjectId.value) await loadTasks(selectedProjectId.value);
      }
    } catch (err) {
      error.value = errorMessage(err, 'Unable to delete project.');
    } finally {
      suppressProjectWatch = false;
    }
  }

  async function createTaskFromForm() {
    const name = taskName.value.trim();
    if (!selectedProjectId.value || !name) return;

    try {
      const input: TaskCreateInput = {
        project_id: selectedProjectId.value,
        name,
      };
      if (taskPhase.value) input.task_phase = taskPhase.value;
      if (taskType.value) input.task_type = taskType.value;

      const task = await createTask(input);
      tasks.value = [...tasks.value, task];
      await projectSelection.loadProjects();
      taskName.value = '';
    } catch (err) {
      error.value = errorMessage(err, 'Unable to create task.');
    }
  }

  async function moveTask(task: Task, phase: string) {
    try {
      const moved = await updateTaskPhase(task.id, phase);
      tasks.value = tasks.value.map((item) => (item.id === moved.id ? moved : item));
      await refreshSelectedProjectTasks();
    } catch (err) {
      error.value = errorMessage(err, 'Unable to move task.');
    }
  }

  async function openTask(task: Task) {
    try {
      taskDetail.value = await getTask(task.id);
      taskEditName.value = taskDetail.value.task.name;
      taskEditDescription.value = taskDetail.value.task.description;
      taskEditType.value = taskDetail.value.task.task_type;
      requirementDefinition.value = '';
      cancelRequirementEdit();
      taskDialogOpen.value = true;
    } catch (err) {
      error.value = errorMessage(err, 'Unable to load task.');
    }
  }

  async function saveTask() {
    if (!taskDetail.value || !taskEditName.value.trim()) return;

    try {
      const task = await updateTask({
        id: taskDetail.value.task.id,
        name: taskEditName.value.trim(),
        description: taskEditDescription.value.trim(),
        task_type: taskEditType.value,
      });
      tasks.value = tasks.value.map((item) => (item.id === task.id ? task : item));
      if (taskDetail.value) {
        taskDetail.value = {
          ...taskDetail.value,
          task,
        };
      }
      taskDialogOpen.value = false;
    } catch (err) {
      error.value = errorMessage(err, 'Unable to update task.');
    }
  }

  async function createRequirementFromForm() {
    const definition = requirementDefinition.value.trim();
    if (!taskDetail.value || !definition) return;

    try {
      const mutation = await createRequirement(taskDetail.value.task.id, definition);
      applyRequirementMutation(mutation);
      await refreshSelectedProjectTasks();
      requirementDefinition.value = '';
    } catch (err) {
      error.value = errorMessage(err, 'Unable to create requirement.');
    }
  }

  async function toggleRequirement(requirement: Requirement, done: boolean) {
    try {
      const mutation = await updateRequirementDone(requirement.id, done);
      applyRequirementMutation(mutation);
      await refreshSelectedProjectTasks();
    } catch (err) {
      error.value = errorMessage(err, 'Unable to update requirement.');
    }
  }

  function startRequirementEdit(requirement: Requirement) {
    editingRequirementId.value = requirement.id;
    editingRequirementDefinition.value = requirement.definition;
  }

  function cancelRequirementEdit() {
    editingRequirementId.value = 0;
    editingRequirementDefinition.value = '';
  }

  async function saveRequirement(requirement: Requirement) {
    const definition = editingRequirementDefinition.value.trim();
    if (!definition) return;

    try {
      const mutation = await updateRequirement({
        id: requirement.id,
        definition,
      });
      applyRequirementMutation(mutation);
      await refreshSelectedProjectTasks();
      cancelRequirementEdit();
    } catch (err) {
      error.value = errorMessage(err, 'Unable to update requirement.');
    }
  }

  async function removeRequirement(requirement: Requirement) {
    try {
      const mutation = await deleteRequirement(requirement.id);
      applyRequirementMutation(mutation);
      await refreshSelectedProjectTasks();
      if (editingRequirementId.value === requirement.id) cancelRequirementEdit();
    } catch (err) {
      error.value = errorMessage(err, 'Unable to delete requirement.');
    }
  }

  function applyRequirementMutation(mutation: RequirementMutation) {
    tasks.value = tasks.value.map((item) => (item.id === mutation.task.id ? mutation.task : item));
    if (!taskDetail.value || taskDetail.value.task.id !== mutation.task.id) return;

    taskDetail.value = {
      task: mutation.task,
      requirements: mutation.requirements,
    };
  }

  async function refreshSelectedProjectTasks() {
    if (selectedProjectId.value) {
      await loadTasks(selectedProjectId.value);
    }
  }

  async function removeTask(task: Task) {
    try {
      await deleteTask(task.id);
      await refreshSelectedProjectTasks();
      await projectSelection.loadProjects();
    } catch (err) {
      error.value = errorMessage(err, 'Unable to delete task.');
    }
  }

  onMounted(() => {
    void loadAll();
  });

  watch(selectedProjectId, (projectId) => {
    if (!tasksEnabled) return;
    if (suppressProjectWatch) return;
    error.value = '';

    if (!projectId) {
      tasks.value = [];
      return;
    }

    void loadTasks(projectId).catch((err: unknown) => {
      error.value = errorMessage(err, 'Unable to load tasks.');
    });
  });

  return {
    projects,
    tasks,
    phases,
    types,
    selectedProjectId,
    projectName,
    taskName,
    taskType,
    taskPhase,
    loading,
    error,
    projectDialogOpen,
    projectEditId,
    projectEditName,
    taskDialogOpen,
    taskDetail,
    taskEditName,
    taskEditDescription,
    taskEditType,
    requirementDefinition,
    editingRequirementId,
    editingRequirementDefinition,
    selectedProject,
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
    createTaskFromForm,
    moveTask,
    openTask,
    saveTask,
    createRequirementFromForm,
    toggleRequirement,
    startRequirementEdit,
    cancelRequirementEdit,
    saveRequirement,
    removeRequirement,
    removeTask,
  };
}
