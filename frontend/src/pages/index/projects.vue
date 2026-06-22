<template>
  <q-page class="app-page">
    <section class="page-heading">
      <div>
        <h1>Projects</h1>
        <p>Project and task board backed by the existing database contract.</p>
      </div>
      <q-btn
        color="primary"
        icon="refresh"
        label="Refresh"
        :loading="loading"
        no-caps
        @click="loadAll"
      />
    </section>

    <q-banner v-if="error" class="status-banner status-banner--error" rounded>
      <template #avatar>
        <q-icon name="warning" />
      </template>
      {{ error }}
    </q-banner>

    <section class="projects-shell">
      <aside class="project-panel">
        <form class="create-row" @submit.prevent="createProjectFromForm">
          <q-input v-model="projectName" dense outlined label="Project name" class="create-input" />
          <q-btn
            color="primary"
            icon="add"
            type="submit"
            :disable="!projectName.trim()"
            round
            unelevated
          >
            <q-tooltip>Create project</q-tooltip>
          </q-btn>
        </form>

        <q-list bordered separator class="project-list">
          <q-item
            v-for="project in projects"
            :key="project.id"
            clickable
            :active="project.id === selectedProjectId"
            active-class="selected-project"
            @click="selectProject(project.id)"
          >
            <q-item-section>
              <q-item-label>{{ project.name }}</q-item-label>
            </q-item-section>
            <q-item-section side>
              <div class="item-actions">
                <q-btn dense flat round icon="edit" @click.stop="startProjectRename(project)">
                  <q-tooltip>Rename project</q-tooltip>
                </q-btn>
                <q-btn
                  dense
                  flat
                  round
                  icon="delete"
                  color="negative"
                  @click.stop="removeProject(project)"
                >
                  <q-tooltip>Delete project</q-tooltip>
                </q-btn>
              </div>
            </q-item-section>
          </q-item>
        </q-list>

        <div v-if="!projects.length && !loading" class="empty-state compact-empty">
          <q-icon name="folder_open" size="32px" />
          <span>Create a project to start tracking tasks.</span>
        </div>
      </aside>

      <main class="task-board">
        <form v-if="selectedProject" class="task-create-row" @submit.prevent="createTaskFromForm">
          <q-input v-model="taskName" dense outlined label="Task name" class="create-input" />
          <q-select
            v-model="taskType"
            dense
            outlined
            emit-value
            map-options
            label="Type"
            :options="typeOptions"
            class="task-select"
          />
          <q-select
            v-model="taskPhase"
            dense
            outlined
            emit-value
            map-options
            label="Phase"
            :options="phaseOptions"
            class="task-select"
          />
          <q-btn
            color="primary"
            icon="add_task"
            type="submit"
            :disable="!taskName.trim()"
            no-caps
            label="Task"
          />
        </form>

        <div v-if="!selectedProject" class="empty-state">
          <q-icon name="view_kanban" size="44px" />
          <span>Select or create a project to view its task board.</span>
        </div>

        <div v-else class="phase-board">
          <section v-for="phase in boardPhases" :key="phase.slug" class="phase-column">
            <header class="phase-heading">
              <span>{{ phase.slug }}</span>
              <q-badge color="grey-7" :label="tasksByPhase[phase.slug]?.length || 0" />
            </header>

            <q-card
              v-for="task in tasksByPhase[phase.slug]"
              :key="task.id"
              flat
              bordered
              class="task-card"
              @click="openTask(task)"
            >
              <q-card-section>
                <div class="task-card-title">{{ task.name }}</div>
                <div class="task-card-meta">{{ task.task_type }} · {{ task.completed }}%</div>
                <q-linear-progress :value="task.completed / 100" rounded class="q-mt-sm" />
              </q-card-section>
              <q-card-actions align="between">
                <q-select
                  :model-value="task.task_phase"
                  dense
                  borderless
                  emit-value
                  map-options
                  :options="phaseOptions"
                  class="phase-move"
                  @click.stop
                  @update:model-value="(phase) => moveTask(task, phase)"
                />
                <q-btn
                  dense
                  flat
                  round
                  icon="delete"
                  color="negative"
                  @click.stop="removeTask(task)"
                >
                  <q-tooltip>Delete task</q-tooltip>
                </q-btn>
              </q-card-actions>
            </q-card>

            <div v-if="!tasksByPhase[phase.slug]?.length" class="phase-empty">No tasks</div>
          </section>
        </div>
      </main>
    </section>

    <q-dialog v-model="projectDialogOpen">
      <q-card class="dialog-card">
        <q-card-section>
          <div class="text-subtitle1">Rename Project</div>
        </q-card-section>
        <q-card-section>
          <q-input v-model="projectEditName" autofocus outlined label="Project name" />
        </q-card-section>
        <q-card-actions align="right">
          <q-btn flat label="Cancel" no-caps v-close-popup />
          <q-btn
            color="primary"
            label="Save"
            no-caps
            :disable="!projectEditName.trim()"
            @click="saveProjectName"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <q-dialog v-model="taskDialogOpen">
      <q-card class="dialog-card">
        <q-card-section>
          <div class="text-subtitle1">{{ taskDetail?.task.name }}</div>
          <div v-if="taskDetail" class="task-card-meta">
            {{ taskDetail.task.task_phase }} · {{ taskDetail.task.completed }}%
          </div>
        </q-card-section>
        <q-card-section v-if="taskDetail">
          <q-input v-model="taskEditName" outlined label="Task name" class="q-mb-md" />
          <q-input
            v-model="taskEditDescription"
            outlined
            type="textarea"
            label="Description"
            class="q-mb-md"
          />
          <q-select
            v-model="taskEditType"
            outlined
            emit-value
            map-options
            label="Type"
            :options="typeOptions"
            class="q-mb-md"
          />
          <div class="requirements-list">
            <div class="requirements-heading">
              <div class="text-subtitle2">Requirements</div>
              <q-badge color="grey-7" :label="`${taskDetail.task.completed}%`" />
            </div>
            <form class="requirement-create-row" @submit.prevent="createRequirementFromForm">
              <q-input
                v-model="requirementDefinition"
                dense
                outlined
                label="Requirement"
                class="create-input"
              />
              <q-btn
                color="primary"
                icon="playlist_add"
                type="submit"
                :disable="!requirementDefinition.trim()"
                round
                unelevated
              >
                <q-tooltip>Add requirement</q-tooltip>
              </q-btn>
            </form>
            <q-list v-if="taskDetail.requirements.length" bordered separator>
              <q-item
                v-for="requirement in taskDetail.requirements"
                :key="requirement.id"
                class="requirement-item"
              >
                <q-item-section avatar>
                  <q-checkbox
                    :model-value="requirement.done"
                    @update:model-value="(done) => toggleRequirement(requirement, Boolean(done))"
                  />
                </q-item-section>
                <q-item-section>
                  <q-input
                    v-if="editingRequirementId === requirement.id"
                    v-model="editingRequirementDefinition"
                    dense
                    outlined
                    autofocus
                  />
                  <span v-else>{{ requirement.definition }}</span>
                </q-item-section>
                <q-item-section side>
                  <div class="item-actions">
                    <template v-if="editingRequirementId === requirement.id">
                      <q-btn
                        dense
                        flat
                        round
                        icon="check"
                        color="primary"
                        :disable="!editingRequirementDefinition.trim()"
                        @click="saveRequirement(requirement)"
                      >
                        <q-tooltip>Save requirement</q-tooltip>
                      </q-btn>
                      <q-btn dense flat round icon="close" @click="cancelRequirementEdit">
                        <q-tooltip>Cancel</q-tooltip>
                      </q-btn>
                    </template>
                    <template v-else>
                      <q-btn
                        dense
                        flat
                        round
                        icon="edit"
                        @click="startRequirementEdit(requirement)"
                      >
                        <q-tooltip>Edit requirement</q-tooltip>
                      </q-btn>
                      <q-btn
                        dense
                        flat
                        round
                        icon="delete"
                        color="negative"
                        @click="removeRequirement(requirement)"
                      >
                        <q-tooltip>Delete requirement</q-tooltip>
                      </q-btn>
                    </template>
                  </div>
                </q-item-section>
              </q-item>
            </q-list>
            <div v-else class="phase-empty">No requirements yet</div>
          </div>
        </q-card-section>
        <q-card-actions align="right">
          <q-btn flat label="Cancel" no-caps v-close-popup />
          <q-btn
            color="primary"
            label="Save"
            no-caps
            :disable="!taskEditName.trim()"
            @click="saveTask"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import {
  createProject,
  createRequirement,
  createTask,
  deleteProject,
  deleteRequirement,
  deleteTask,
  getTask,
  getTaskReferences,
  listProjects,
  listTasks,
  updateRequirement,
  updateRequirementDone,
  updateProject,
  updateTask,
  updateTaskPhase,
  type Project,
  type Requirement,
  type RequirementMutation,
  type ReferenceOption,
  type Task,
  type TaskDetail,
} from '@/services/api';

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
const phaseOptions = computed(() =>
  phases.value.map((phase) => ({ label: phase.slug, value: phase.slug })),
);
const typeOptions = computed(() =>
  types.value.map((type) => ({ label: type.slug, value: type.slug })),
);
const boardPhases = computed(() => (phases.value.length ? phases.value : uniqueTaskPhases.value));

const uniqueTaskPhases = computed(() =>
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
    error.value = err instanceof Error ? err.message : 'Unable to load projects.';
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
    error.value = err instanceof Error ? err.message : 'Unable to load tasks.';
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
    error.value = err instanceof Error ? err.message : 'Unable to create project.';
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
    error.value = err instanceof Error ? err.message : 'Unable to update project.';
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
    error.value = err instanceof Error ? err.message : 'Unable to delete project.';
  }
}

async function createTaskFromForm() {
  const name = taskName.value.trim();
  if (!selectedProjectId.value || !name) return;

  try {
    const input: {
      project_id: number;
      name: string;
      task_phase?: string;
      task_type?: string;
    } = {
      project_id: selectedProjectId.value,
      name,
    };
    if (taskPhase.value) input.task_phase = taskPhase.value;
    if (taskType.value) input.task_type = taskType.value;

    const task = await createTask(input);
    tasks.value = [...tasks.value, task];
    taskName.value = '';
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unable to create task.';
  }
}

async function moveTask(task: Task, phase: string) {
  try {
    const moved = await updateTaskPhase(task.id, phase);
    tasks.value = tasks.value.map((item) => (item.id === moved.id ? moved : item));
    await refreshSelectedProjectTasks();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unable to move task.';
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
    error.value = err instanceof Error ? err.message : 'Unable to load task.';
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
    error.value = err instanceof Error ? err.message : 'Unable to update task.';
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
    error.value = err instanceof Error ? err.message : 'Unable to create requirement.';
  }
}

async function toggleRequirement(requirement: Requirement, done: boolean) {
  try {
    const mutation = await updateRequirementDone(requirement.id, done);
    applyRequirementMutation(mutation);
    await refreshSelectedProjectTasks();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unable to update requirement.';
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
    error.value = err instanceof Error ? err.message : 'Unable to update requirement.';
  }
}

async function removeRequirement(requirement: Requirement) {
  try {
    const mutation = await deleteRequirement(requirement.id);
    applyRequirementMutation(mutation);
    await refreshSelectedProjectTasks();
    if (editingRequirementId.value === requirement.id) cancelRequirementEdit();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unable to delete requirement.';
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
    error.value = err instanceof Error ? err.message : 'Unable to delete task.';
  }
}

onMounted(() => {
  void loadAll();
});
</script>
