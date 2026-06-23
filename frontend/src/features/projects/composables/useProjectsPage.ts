import { computed, onMounted, ref } from 'vue';
import {
  createProject,
  deleteProject,
  listProjects,
  updateProject,
} from '@/features/projects/api/projectApi';
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

export function useProjectsPage() {
  const projects = ref<Project[]>([]);
  const tasks = ref<Task[]>([]);
  const phases = ref<ReferenceOption[]>([]);
  const types = ref<ReferenceOption[]>([]);
  const selectedProjectId = ref(0);
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

  const selectedProject = computed(() =>
    projects.value.find((project) => project.id === selectedProjectId.value),
  );
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
      const [references, loadedProjects] = await Promise.all([getTaskReferences(), listProjects()]);
      phases.value = references.phases;
      types.value = references.types;
      if (!taskType.value) taskType.value = references.types[0]?.slug || '';
      if (!taskPhase.value) taskPhase.value = references.phases[0]?.slug || '';
      projects.value = loadedProjects;

      if (!selectedProjectId.value && loadedProjects.length) {
        selectedProjectId.value = loadedProjects[0]!.id;
      }

      if (selectedProjectId.value) {
        await loadTasks(selectedProjectId.value);
      }
    } catch (err) {
      error.value = errorMessage(err, 'Unable to load projects.');
    } finally {
      loading.value = false;
    }
  }

  async function loadTasks(projectId: number) {
    tasks.value = await listTasks(projectId);
  }

  async function selectProject(projectId: number) {
    selectedProjectId.value = projectId;
    error.value = '';
    try {
      await loadTasks(projectId);
    } catch (err) {
      error.value = errorMessage(err, 'Unable to load tasks.');
    }
  }

  async function createProjectFromForm() {
    const name = projectName.value.trim();
    if (!name) return;

    try {
      const project = await createProject(name);
      projects.value = [...projects.value, project].sort((a, b) => a.name.localeCompare(b.name));
      projectName.value = '';
      await selectProject(project.id);
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
      const project = await updateProject(projectEditId.value, name);
      projects.value = projects.value.map((item) => (item.id === project.id ? project : item));
      projectDialogOpen.value = false;
    } catch (err) {
      error.value = errorMessage(err, 'Unable to update project.');
    }
  }

  async function removeProject(project: Project) {
    try {
      await deleteProject(project.id);
      projects.value = projects.value.filter((item) => item.id !== project.id);
      if (selectedProjectId.value === project.id) {
        selectedProjectId.value = projects.value[0]?.id || 0;
        tasks.value = [];
        if (selectedProjectId.value) await loadTasks(selectedProjectId.value);
      }
    } catch (err) {
      error.value = errorMessage(err, 'Unable to delete project.');
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
