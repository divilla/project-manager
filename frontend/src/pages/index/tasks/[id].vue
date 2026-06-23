<template>
  <q-page class="app-page">
    <q-banner v-if="error" class="status-banner status-banner--error" rounded>
      <template #avatar>
        <q-icon name="warning" />
      </template>
      {{ error }}
    </q-banner>

    <div v-if="loading" class="empty-state">
      <q-spinner size="32px" color="primary" />
    </div>

    <div v-else-if="!currentTask" class="empty-state">
      <q-icon name="task_alt" size="44px" />
      <span>Task not found.</span>
    </div>

    <template v-else>
      <q-markup-table flat bordered>
        <thead>
          <tr>
            <th class="text-right">nr</th>
            <th class="text-center">Type</th>
            <th class="text-left">Name</th>
            <th class="text-center">Diff</th>
            <th class="text-center">Pri</th>
            <th class="text-center">Phase</th>
            <th class="text-center">Complete</th>
            <th class="text-center">Modified</th>
            <th class="text-center">Version</th>
            <th class="task-actions-cell"></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="task in ancestorTasks"
            :key="task.id"
            class="bg-deep-purple-3 cursor-pointer"
            @click="openTask(task.id)"
          >
            <td class="text-right">#{{ task.id }}</td>
            <td class="text-center">
              <q-btn :color="taskTypeColor(task.task_type)" :label="task.task_type" />
            </td>
            <td class="text-left">{{ task.name }}</td>
            <td class="text-center">{{ task.difficulty }}</td>
            <td class="text-center">{{ task.priority }}</td>
            <td class="text-center">{{ task.task_phase }}</td>
            <td class="text-center">{{ completionLabel(task) }}</td>
            <td class="text-center">{{ formatModified(task.modified) }}</td>
            <td class="text-center">{{ task.version }}</td>
            <td class="task-actions-cell"></td>
          </tr>

          <tr class="bg-primary text-white text-weight-bold">
            <td class="text-right">#{{ currentTask.id }}</td>
            <td class="text-center">
              <q-btn :color="taskTypeColor(currentTask.task_type)" :label="currentTask.task_type" />
            </td>
            <td class="text-left">{{ currentTask.name }}</td>
            <td class="text-center">{{ currentTask.difficulty }}</td>
            <td class="text-center">{{ currentTask.priority }}</td>
            <td class="text-center">{{ currentTask.task_phase }}</td>
            <td class="text-center">{{ completionLabel(currentTask) }}</td>
            <td class="text-center">{{ formatModified(currentTask.modified) }}</td>
            <td class="text-center">{{ currentTask.version }}</td>
            <td class="task-actions-cell">
              <q-btn-dropdown
                flat
                round
                color="white"
                dropdown-icon="more_vert"
                aria-label="Task actions"
                @click.stop
              >
                <q-list>
                  <q-item
                    clickable
                    v-close-popup
                    data-action="edit-task"
                    @click.stop="openTaskEdit(currentTask.id)"
                  >
                    <q-item-section>
                      <q-item-label>Edit</q-item-label>
                    </q-item-section>
                  </q-item>

                  <q-item
                    clickable
                    v-close-popup
                    data-action="delete-task"
                    :disable="!isLeafTask(currentTask) || isTaskDeleting(currentTask.id)"
                    @click.stop="confirmDeleteTask(currentTask)"
                  >
                    <q-item-section>
                      <q-item-label>Delete</q-item-label>
                    </q-item-section>
                  </q-item>
                </q-list>
              </q-btn-dropdown>
            </td>
          </tr>

          <tr
            v-for="task in childTasks"
            :key="task.id"
            class="bg-teal-4 cursor-pointer"
            @click="openTask(task.id)"
          >
            <td class="text-right">#{{ task.id }}</td>
            <td class="text-center">
              <q-btn :color="taskTypeColor(task.task_type)" :label="task.task_type" />
            </td>
            <td class="text-left">{{ task.name }}</td>
            <td class="text-center">{{ task.difficulty }}</td>
            <td class="text-center">{{ task.priority }}</td>
            <td class="text-center">{{ task.task_phase }}</td>
            <td class="text-center">{{ completionLabel(task) }}</td>
            <td class="text-center">{{ formatModified(task.modified) }}</td>
            <td class="text-center">{{ task.version }}</td>
            <td class="task-actions-cell">
              <q-btn-dropdown
                flat
                round
                color="white"
                dropdown-icon="more_vert"
                aria-label="Task actions"
                @click.stop
              >
                <q-list>
                  <q-item
                    clickable
                    v-close-popup
                    data-action="edit-task"
                    @click.stop="openTaskEdit(task.id)"
                  >
                    <q-item-section>
                      <q-item-label>Edit</q-item-label>
                    </q-item-section>
                  </q-item>

                  <q-item
                    clickable
                    v-close-popup
                    data-action="delete-task"
                    :disable="!isLeafTask(task) || isTaskDeleting(task.id)"
                    @click.stop="confirmDeleteTask(task)"
                  >
                    <q-item-section>
                      <q-item-label>Delete</q-item-label>
                    </q-item-section>
                  </q-item>
                </q-list>
              </q-btn-dropdown>
            </td>
          </tr>
          <tr>
            <td colspan="2"></td>
            <td colspan="8">
              <q-btn
                rounded
                color="primary"
                icon="add_box"
                label="Add Child Task"
                @click="openTaskCreate(currentTask.id)"
              />
            </td>
          </tr>

          <tr class="text-weight-bold">
            <td class="text-right">nr</td>
            <td class="text-center">&nbsp;</td>
            <td class="text-left" colspan="4">Requirement</td>
            <td class="text-center">Complete</td>
            <td class="text-center">Modified</td>
            <td class="text-center">Version</td>
            <td class="task-actions-cell"></td>
          </tr>
          <tr v-for="requirement in requirements" :key="requirement.id">
            <td class="text-right">#{{ requirement.id }}</td>
            <td class="text-center">&nbsp;</td>
            <td class="text-left" colspan="4">
              {{ requirement.definition }}
            </td>
            <td class="text-center">
              <q-checkbox
                :model-value="requirement.done"
                dense
                color="secondary"
                :disable="isRequirementUpdating(requirement.id)"
                @update:model-value="toggleRequirement(requirement, Boolean($event))"
              />
            </td>
            <td class="text-center">{{ formatModified(requirement.modified) }}</td>
            <td class="text-center">{{ requirement.version }}</td>
            <td class="task-actions-cell">
              <q-btn-dropdown
                flat
                round
                color="secondary"
                dropdown-icon="more_vert"
                aria-label="Requirement actions"
                @click.stop
              >
                <q-list>
                  <q-item
                    clickable
                    v-close-popup
                    data-action="edit-requirement"
                    :disable="isRequirementUpdating(requirement.id)"
                    @click.stop="openRequirementEdit(requirement)"
                  >
                    <q-item-section>
                      <q-item-label>Edit</q-item-label>
                    </q-item-section>
                  </q-item>

                  <q-item
                    clickable
                    v-close-popup
                    data-action="delete-requirement"
                    :disable="isRequirementUpdating(requirement.id)"
                    @click.stop="confirmDeleteRequirement(requirement)"
                  >
                    <q-item-section>
                      <q-item-label>Delete</q-item-label>
                    </q-item-section>
                  </q-item>
                </q-list>
              </q-btn-dropdown>
            </td>
          </tr>
          <tr>
            <td>&nbsp;</td>
            <td>&nbsp;</td>
            <td colspan="8">
              <q-btn
                rounded
                color="secondary"
                icon="library_add_check"
                label="Add Requirement"
                @click="openRequirementCreate"
              />
            </td>
          </tr>
        </tbody>
      </q-markup-table>

      <q-markup-table flat bordered class="task-detail-table">
        <tbody>
          <tr>
            <td class="task-detail-description-cell" style="padding-top: 32px; padding-bottom: 32px;">
              <div class="apply-markdown" v-html="currentTask.description_html" />
            </td>
          </tr>
          <tr v-for="task in parentDescriptionTasks" :key="`parent-description-${task.id}`">
            <td class="task-detail-description-cell" style="padding-top: 32px; padding-bottom: 32px;">
              <div class="apply-markdown" v-html="descriptionHTML(task)" />
            </td>
          </tr>
        </tbody>
      </q-markup-table>
    </template>

    <q-dialog v-model="requirementEditOpen">
      <q-card class="requirement-edit-dialog">
        <q-card-section>
          <div class="text-h6">Edit Requirement</div>
        </q-card-section>

        <q-form @submit.prevent="saveRequirementEdit">
          <q-card-section>
            <q-input
              v-model="editingRequirementDefinition"
              autofocus
              filled
              autogrow
              type="textarea"
              input-style="min-height: 72px"
              label="Requirement"
            />
          </q-card-section>

          <q-card-actions align="right">
            <q-btn flat icon="close" label="Cancel" :disable="savingRequirement" @click="closeRequirementEdit" />
            <q-btn
              color="secondary"
              icon="save"
              label="Save"
              type="submit"
              :loading="savingRequirement"
              :disable="!editingRequirementDefinition.trim()"
            />
          </q-card-actions>
        </q-form>
      </q-card>
    </q-dialog>

    <q-dialog v-model="requirementCreateOpen">
      <q-card class="requirement-edit-dialog">
        <q-card-section>
          <div class="text-h6">Add Requirement</div>
        </q-card-section>

        <q-form @submit.prevent="saveRequirementCreate">
          <q-card-section>
            <q-input
              v-model="newRequirementDefinition"
              autofocus
              filled
              autogrow
              type="textarea"
              input-style="min-height: 72px"
              label="Requirement"
            />
          </q-card-section>

          <q-card-actions align="right">
            <q-btn
              flat
              icon="close"
              label="Cancel"
              :disable="creatingRequirement"
              @click="closeRequirementCreate"
            />
            <q-btn
              color="secondary"
              icon="save"
              label="Add"
              type="submit"
              :loading="creatingRequirement"
              :disable="!newRequirementDefinition.trim()"
            />
          </q-card-actions>
        </q-form>
      </q-card>
    </q-dialog>

    <DeleteConfirmationDialog v-model:open="deleteConfirmationOpen" @confirm="confirm" />

    <q-dialog v-model="projectMismatchOpen" persistent>
      <q-card class="project-mismatch-dialog">
        <q-card-section>
          <div class="text-h6">Switch project?</div>
        </q-card-section>

        <q-card-section>
          This task belongs to {{ projectMismatch?.requiredProjectName }}. The selected project is
          {{ projectMismatch?.currentProjectName }}.
        </q-card-section>

        <q-card-actions align="right">
          <q-btn flat label="Stay" @click="stayOnCurrentProject" />
          <q-btn color="primary" label="Switch" @click="switchToTaskProject" />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useProjectSelectionStore } from '@/features/projects/model/projectSelection.store';
import {
  createRequirement,
  deleteRequirement,
  updateRequirement,
  updateRequirementDone,
} from '@/features/requirements/api/requirementApi';
import type { Requirement } from '@/features/requirements/model/requirement.types';
import { deleteTask, getRenderedTaskDescriptions, getTask } from '@/features/tasks/api/taskApi';
import { useTaskCacheStore } from '@/features/tasks/model/taskCache.store';
import type { Task } from '@/features/tasks/model/task.types';
import DeleteConfirmationDialog from '@/shared/ui/DeleteConfirmationDialog.vue';

interface ProjectMismatch {
  currentProjectId: number;
  currentProjectName: string;
  requiredProjectId: number;
  requiredProjectName: string;
}

const route = useRoute();
const router = useRouter();
const projectSelection = useProjectSelectionStore();
const taskCache = useTaskCacheStore();
const tasks = computed(() => taskCache.tasks);

const requirements = ref<Requirement[]>([]);
const detailTask = ref<Task | null>(null);
const detailLoading = ref(false);
const deletingTaskIds = ref<number[]>([]);
const updatingRequirementIds = ref<number[]>([]);
const requirementCreateOpen = ref(false);
const newRequirementDefinition = ref('');
const creatingRequirement = ref(false);
const requirementEditOpen = ref(false);
const editingRequirement = ref<Requirement | null>(null);
const editingRequirementDefinition = ref('');
const savingRequirement = ref(false);
const deleteConfirmationOpen = ref(false);
const pendingDeleteAction = ref<(() => Promise<void>) | null>(null);
const ancestorDescriptionHTML = ref<Record<number, string>>({});
const projectMismatchOpen = ref(false);
const projectMismatch = ref<ProjectMismatch | null>(null);
const requirementError = ref('');
const error = computed(() => projectSelection.error || requirementError.value);
const loading = computed(() => detailLoading.value || taskCache.loading);
const taskId = computed(() => Number(route.params.id));
const taskMap = computed(() => new Map(tasks.value.map((task) => [task.id, task])));
const currentTask = computed(() =>
  detailTask.value?.id === taskId.value ? detailTask.value : taskMap.value.get(taskId.value) || null,
);

const ancestorTasks = computed(() => {
  const ancestors: Task[] = [];
  const seen = new Set<number>();
  let task = currentTask.value;

  while (task?.parent_id) {
    if (seen.has(task.parent_id)) break;
    seen.add(task.parent_id);

    const parent = taskMap.value.get(task.parent_id);
    if (!parent) break;

    ancestors.unshift(parent);
    task = parent;
  }

  return ancestors;
});

const parentDescriptionTasks = computed(() => ancestorTasks.value.slice().reverse());

const childTasks = computed(() =>
  tasks.value
    .filter((task) => task.parent_id === currentTask.value?.id)
    .sort((left, right) => left.priority - right.priority || left.id - right.id),
);

async function ensureProjectsLoaded() {
  if (!projectSelection.hasLoaded) {
    await projectSelection.loadProjects();
  }
}

async function loadTaskDetail() {
  if (!taskId.value) return;

  detailLoading.value = true;
  requirementError.value = '';
  try {
    const detail = await getTask(taskId.value);
    detailTask.value = detail.task;
    requirements.value = detail.requirements;
    await ensureProjectsLoaded();
    detectProjectMismatch(detail.task);
    await taskCache.loadProjectTasks(detail.task.project_id);
    taskCache.upsertTask(detail.task);
    await loadAncestorDescriptions();
  } catch (err) {
    detailTask.value = null;
    requirements.value = [];
    ancestorDescriptionHTML.value = {};
    requirementError.value = err instanceof Error ? err.message : 'Unable to load task.';
  } finally {
    detailLoading.value = false;
  }
}

async function loadAncestorDescriptions() {
  const ids = parentDescriptionTasks.value.map((task) => task.id);
  ancestorDescriptionHTML.value = {};
  if (!ids.length) return;

  try {
    const rendered = await getRenderedTaskDescriptions(ids);
    ancestorDescriptionHTML.value = Object.fromEntries(
      rendered.descriptions.map((description) => [description.id, description.description_html]),
    );
  } catch (err) {
    requirementError.value =
      err instanceof Error ? err.message : 'Unable to load parent task descriptions.';
  }
}

function descriptionHTML(task: Task) {
  return ancestorDescriptionHTML.value[task.id] ?? task.description_html;
}

function projectName(projectId: number) {
  return projectSelection.projects.find((project) => project.id === projectId)?.name || `#${projectId}`;
}

function detectProjectMismatch(task: Task) {
  const currentProjectId = projectSelection.currentProjectId;
  if (!currentProjectId || currentProjectId === task.project_id || projectSelection.isSwitchingProject) {
    projectMismatchOpen.value = false;
    projectMismatch.value = null;
    return;
  }

  projectMismatch.value = {
    currentProjectId,
    currentProjectName: projectName(currentProjectId),
    requiredProjectId: task.project_id,
    requiredProjectName: projectName(task.project_id),
  };
  projectMismatchOpen.value = true;
}

function switchToTaskProject() {
  const mismatch = projectMismatch.value;
  if (!mismatch) return;

  projectMismatchOpen.value = false;
  projectMismatch.value = null;
  projectSelection.beginRouteDrivenProjectSwitch(route.fullPath);
  projectSelection.selectProject(mismatch.requiredProjectId);
}

function stayOnCurrentProject() {
  projectMismatchOpen.value = false;
  projectMismatch.value = null;

  if (typeof window !== 'undefined' && window.history.length > 1) {
    router.back();
    return;
  }

  void router.replace('/tasks');
}

function applyRequirementMutation(requirementList: Requirement[], task: Task) {
  requirements.value = requirementList;
  detailTask.value = task;
  taskCache.upsertTask(task);
}

function isRequirementUpdating(id: number) {
  return updatingRequirementIds.value.includes(id);
}

function isLeafTask(task: Task) {
  return !tasks.value.some((candidate) => candidate.parent_id === task.id);
}

function isTaskDeleting(id: number) {
  return deletingTaskIds.value.includes(id);
}

function confirmDeleteTask(task: Task) {
  if (!isLeafTask(task) || isTaskDeleting(task.id)) return;
  pendingDeleteAction.value = () => deleteTaskConfirmed(task);
  deleteConfirmationOpen.value = true;
}

async function confirm() {
  const action = pendingDeleteAction.value;
  deleteConfirmationOpen.value = false;
  pendingDeleteAction.value = null;
  if (!action) return;

  await action();
}

async function deleteTaskConfirmed(task: Task) {
  if (!task || !isLeafTask(task) || isTaskDeleting(task.id)) return;

  deletingTaskIds.value = [...deletingTaskIds.value, task.id];
  requirementError.value = '';

  try {
    await deleteTask(task.id);
    await taskCache.loadProjectTasks(task.project_id);
    await projectSelection.loadProjects().catch(() => undefined);

    if (task.id === currentTask.value?.id) {
      const parentID = task.parent_id;
      void router.push(parentID ? `/tasks/${parentID}` : '/tasks');
    }
  } catch (err) {
    requirementError.value = err instanceof Error ? err.message : 'Unable to delete task.';
  } finally {
    deletingTaskIds.value = deletingTaskIds.value.filter((id) => id !== task.id);
  }
}

function confirmDeleteRequirement(requirement: Requirement) {
  if (isRequirementUpdating(requirement.id)) return;
  pendingDeleteAction.value = () => deleteRequirementConfirmed(requirement);
  deleteConfirmationOpen.value = true;
}

async function deleteRequirementConfirmed(requirement: Requirement) {
  if (isRequirementUpdating(requirement.id)) return;

  updatingRequirementIds.value = [...updatingRequirementIds.value, requirement.id];
  requirementError.value = '';

  try {
    const mutation = await deleteRequirement(requirement.id);
    applyRequirementMutation(mutation.requirements, mutation.task);
    await taskCache.loadProjectTasks(mutation.task.project_id);
    if (editingRequirement.value?.id === requirement.id) closeRequirementEdit();
  } catch (err) {
    requirementError.value = err instanceof Error ? err.message : 'Unable to delete requirement.';
  } finally {
    updatingRequirementIds.value = updatingRequirementIds.value.filter((id) => id !== requirement.id);
  }
}

async function toggleRequirement(requirement: Requirement, done: boolean) {
  if (isRequirementUpdating(requirement.id)) return;

  updatingRequirementIds.value = [...updatingRequirementIds.value, requirement.id];
  requirementError.value = '';

  try {
    const mutation = await updateRequirementDone(requirement.id, done);
    applyRequirementMutation(mutation.requirements, mutation.task);
    await taskCache.loadProjectTasks(mutation.task.project_id);
  } catch (err) {
    requirementError.value = err instanceof Error ? err.message : 'Unable to update requirement.';
  } finally {
    updatingRequirementIds.value = updatingRequirementIds.value.filter((id) => id !== requirement.id);
  }
}

function openRequirementEdit(requirement: Requirement) {
  editingRequirement.value = requirement;
  editingRequirementDefinition.value = requirement.definition;
  requirementEditOpen.value = true;
}

function closeRequirementEdit() {
  if (savingRequirement.value) return;

  requirementEditOpen.value = false;
  editingRequirement.value = null;
  editingRequirementDefinition.value = '';
}

function openRequirementCreate() {
  newRequirementDefinition.value = '';
  requirementCreateOpen.value = true;
}

function closeRequirementCreate() {
  if (creatingRequirement.value) return;

  requirementCreateOpen.value = false;
  newRequirementDefinition.value = '';
}

async function saveRequirementCreate() {
  if (!currentTask.value) return;

  const definition = newRequirementDefinition.value.trim();
  if (!definition) return;

  creatingRequirement.value = true;
  requirementError.value = '';

  try {
    const mutation = await createRequirement(currentTask.value.id, definition);
    applyRequirementMutation(mutation.requirements, mutation.task);
    requirementCreateOpen.value = false;
    newRequirementDefinition.value = '';
    await taskCache.loadProjectTasks(mutation.task.project_id);
  } catch (err) {
    requirementError.value = err instanceof Error ? err.message : 'Unable to create requirement.';
  } finally {
    creatingRequirement.value = false;
  }
}

async function saveRequirementEdit() {
  if (!editingRequirement.value) return;

  const definition = editingRequirementDefinition.value.trim();
  if (!definition) return;

  savingRequirement.value = true;
  requirementError.value = '';

  try {
    const mutation = await updateRequirement({
      id: editingRequirement.value.id,
      definition,
    });
    applyRequirementMutation(mutation.requirements, mutation.task);
    requirementEditOpen.value = false;
    editingRequirement.value = null;
    editingRequirementDefinition.value = '';
    await taskCache.loadProjectTasks(mutation.task.project_id);
  } catch (err) {
    requirementError.value = err instanceof Error ? err.message : 'Unable to update requirement.';
  } finally {
    savingRequirement.value = false;
  }
}

function openTask(id: number) {
  void router.push(`/tasks/${id}`);
}

function openTaskCreate(parentId: number) {
  void router.push(`/tasks/create/${parentId}`);
}

function openTaskEdit(taskID: number) {
  void router.push(`/tasks/edit/${taskID}`);
}

function formatModified(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;

  return new Intl.DateTimeFormat(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}

function completionLabel(task: Task) {
  return `${task.done_req}/${task.total_req} - ${task.completed}%`;
}

function taskTypeColor(type: string) {
  if (type === 'epic' || type === 'group') return 'purple';
  if (type === 'issue') return 'red';

  return 'teal';
}

onMounted(() => {
  void ensureProjectsLoaded();
  void loadTaskDetail();
});

watch(taskId, () => {
  closeRequirementEdit();
  closeRequirementCreate();
  ancestorDescriptionHTML.value = {};
  projectMismatchOpen.value = false;
  projectMismatch.value = null;
  void ensureProjectsLoaded();
  void loadTaskDetail();
});
</script>

<style scoped lang="scss">
.requirement-edit-dialog {
  width: min(720px, calc(100vw - 32px));
}

.project-mismatch-dialog {
  width: min(420px, calc(100vw - 32px));
}

.task-detail-table {
  margin-top: 32px;
}

.task-actions-cell {
  padding-left: 4px;
  padding-right: 4px;
  text-align: center;
  width: 1%;
  white-space: nowrap;
}

.task-detail-table :deep(table) {
  table-layout: fixed;
  width: 100%;
}

.task-detail-title-cell,
.task-detail-description-cell {
  min-width: 0;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.task-detail-description-cell {
  white-space: normal;
  width: 100%;
}

.apply-markdown {
  overflow-wrap: anywhere;
  word-break: break-word;
  white-space: normal;
}

.apply-markdown :deep(pre) {
  max-width: 100%;
  overflow-x: auto;
  white-space: pre;
}

.apply-markdown :deep(:not(pre, pre *)) {
  overflow-wrap: anywhere;
  word-break: break-word;
}

.apply-markdown :deep(p),
.apply-markdown :deep(li),
.apply-markdown :deep(blockquote),
.apply-markdown :deep(td),
.apply-markdown :deep(th) {
  white-space: normal;
}

.apply-markdown :deep(table) {
  table-layout: fixed;
  width: 100%;
}

.apply-markdown :deep(table td),
.apply-markdown :deep(table th) {
  min-width: 0;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.apply-markdown :deep(img),
.apply-markdown :deep(table) {
  max-width: 100%;
}
</style>
