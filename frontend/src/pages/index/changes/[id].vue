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

    <div v-else-if="!currentChange" class="empty-state">
      <q-icon name="published_with_changes" size="44px" />
      <span>Change not found.</span>
    </div>

    <template v-else>
      <q-markup-table flat bordered class="change-detail-table">
        <thead>
          <tr>
            <th class="text-right">nr</th>
            <th class="text-center">Types</th>
            <th class="text-left">Title</th>
            <th class="text-center">Epic</th>
            <th class="text-center">Phase</th>
            <th class="text-center">Closed</th>
            <th class="text-center">Complete</th>
            <th class="text-center">Modified</th>
            <th class="text-center">Version</th>
            <th class="change-actions-cell"></th>
          </tr>
        </thead>
        <tbody>
          <tr class="bg-primary text-white text-weight-bold">
            <td class="text-right">#{{ currentChange.id }}</td>
            <td class="text-center">{{ currentChange.change_types.join(', ') }}</td>
            <td class="text-left">{{ currentChange.title }}</td>
            <td class="text-center">{{ epicName(currentChange.epic_id) }}</td>
            <td class="text-center">{{ currentChange.change_phase }}</td>
            <td class="text-center">{{ currentChange.closed ? 'Yes' : 'No' }}</td>
            <td class="text-center">{{ completionLabel(currentChange) }}</td>
            <td class="text-center">{{ formatModified(currentChange.modified) }}</td>
            <td class="text-center">{{ currentChange.version }}</td>
            <td class="change-actions-cell">
              <q-btn-dropdown
                flat
                round
                color="white"
                dropdown-icon="more_vert"
                aria-label="Change actions"
                @click.stop
              >
                <q-list>
                  <q-item clickable v-close-popup data-action="edit-change" @click.stop="openChangeEdit(currentChange.id)">
                    <q-item-section>
                      <q-item-label>Edit</q-item-label>
                    </q-item-section>
                  </q-item>

                  <q-item
                    clickable
                    v-close-popup
                    data-action="delete-change"
                    :disable="isChangeDeleting(currentChange.id)"
                    @click.stop="confirmDeleteChange(currentChange)"
                  >
                    <q-item-section>
                      <q-item-label>Delete</q-item-label>
                    </q-item-section>
                  </q-item>
                </q-list>
              </q-btn-dropdown>
            </td>
          </tr>

          <tr class="text-weight-bold">
            <td class="text-right">nr</td>
            <td class="text-center">&nbsp;</td>
            <td class="text-left" colspan="4">Requirement</td>
            <td class="text-center">Complete</td>
            <td class="text-center">Modified</td>
            <td class="text-center">Version</td>
            <td class="change-actions-cell"></td>
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
            <td class="change-actions-cell">
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

      <q-markup-table flat bordered class="change-detail-table">
        <tbody>
          <tr>
            <td class="change-detail-body-cell" style="padding-top: 32px; padding-bottom: 32px;">
              <div class="apply-markdown" v-html="currentChange.body_html" />
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
          <q-card-section class="requirement-edit-fields">
            <q-input
              v-model="editingRequirementDefinition"
              autofocus
              filled
              autogrow
              type="textarea"
              input-style="min-height: 72px"
              label="Requirement"
            />
            <q-select
              v-model="editingRequirementChangeId"
              filled
              emit-value
              map-options
              label="Change"
              :options="changeOptions"
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
          This change belongs to {{ projectMismatch?.requiredProjectName }}. The selected project is
          {{ projectMismatch?.currentProjectName }}.
        </q-card-section>

        <q-card-actions align="right">
          <q-btn flat label="Stay" @click="stayOnCurrentProject" />
          <q-btn color="primary" label="Switch" @click="switchToChangeProject" />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { storeToRefs } from 'pinia';
import { useRoute, useRouter } from 'vue-router';
import { deleteChange, getChange } from '@/features/changes/api/changeApi';
import { useChangeCacheStore } from '@/features/changes/model/changeCache.store';
import type { Change } from '@/features/changes/model/change.types';
import { useProjectSelectionStore } from '@/features/projects/model/projectSelection.store';
import {
  createRequirement,
  deleteRequirement,
  updateRequirement,
  updateRequirementChange,
  updateRequirementDone,
} from '@/features/requirements/api/requirementApi';
import type { Requirement } from '@/features/requirements/model/requirement.types';
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
const changeCache = useChangeCacheStore();
const { changes, epics } = storeToRefs(changeCache);

const requirements = ref<Requirement[]>([]);
const detailChange = ref<Change | null>(null);
const detailLoading = ref(false);
const deletingChangeIds = ref<number[]>([]);
const updatingRequirementIds = ref<number[]>([]);
const requirementCreateOpen = ref(false);
const newRequirementDefinition = ref('');
const creatingRequirement = ref(false);
const requirementEditOpen = ref(false);
const editingRequirement = ref<Requirement | null>(null);
const editingRequirementDefinition = ref('');
const editingRequirementChangeId = ref<number | null>(null);
const savingRequirement = ref(false);
const deleteConfirmationOpen = ref(false);
const pendingDeleteAction = ref<(() => Promise<void>) | null>(null);
const projectMismatchOpen = ref(false);
const projectMismatch = ref<ProjectMismatch | null>(null);
const requirementError = ref('');
const error = computed(() => projectSelection.error || requirementError.value);
const loading = computed(() => detailLoading.value || changeCache.loading);
const changeId = computed(() => Number(route.params.id));
const changeMap = computed(() => new Map(changes.value.map((change) => [change.id, change])));
const currentChange = computed(() =>
  detailChange.value?.id === changeId.value ? detailChange.value : changeMap.value.get(changeId.value) || null,
);
const changeOptions = computed(() =>
  changes.value.map((change) => ({ label: `#${change.id} ${change.title}`, value: change.id })),
);

async function ensureProjectsLoaded() {
  if (!projectSelection.hasLoaded) {
    await projectSelection.loadProjects();
  }
}

async function loadChangeDetail() {
  if (!changeId.value) return;

  detailLoading.value = true;
  requirementError.value = '';
  try {
    const detail = await getChange(changeId.value);
    detailChange.value = detail.change;
    requirements.value = detail.requirements;
    await ensureProjectsLoaded();
    detectProjectMismatch(detail.change);
    await changeCache.loadProjectChanges(detail.change.project_id);
    changeCache.upsertChange(detail.change);
  } catch (err) {
    detailChange.value = null;
    requirements.value = [];
    requirementError.value = err instanceof Error ? err.message : 'Unable to load change.';
  } finally {
    detailLoading.value = false;
  }
}

function projectName(projectId: number) {
  return projectSelection.projects.find((project) => project.id === projectId)?.name || `#${projectId}`;
}

function detectProjectMismatch(change: Change) {
  const currentProjectId = projectSelection.currentProjectId;
  if (!currentProjectId || currentProjectId === change.project_id || projectSelection.isSwitchingProject) {
    projectMismatchOpen.value = false;
    projectMismatch.value = null;
    return;
  }

  projectMismatch.value = {
    currentProjectId,
    currentProjectName: projectName(currentProjectId),
    requiredProjectId: change.project_id,
    requiredProjectName: projectName(change.project_id),
  };
  projectMismatchOpen.value = true;
}

function switchToChangeProject() {
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

  void router.replace('/changes');
}

function applyRequirementMutation(requirementList: Requirement[], change: Change) {
  if (change.id === changeId.value) {
    requirements.value = requirementList;
    detailChange.value = change;
  }
  changeCache.upsertChange(change);
}

function isRequirementUpdating(id: number) {
  return updatingRequirementIds.value.includes(id);
}

function isChangeDeleting(id: number) {
  return deletingChangeIds.value.includes(id);
}

function confirmDeleteChange(change: Change) {
  if (isChangeDeleting(change.id)) return;
  pendingDeleteAction.value = () => deleteChangeConfirmed(change);
  deleteConfirmationOpen.value = true;
}

async function confirm() {
  const action = pendingDeleteAction.value;
  deleteConfirmationOpen.value = false;
  pendingDeleteAction.value = null;
  if (!action) return;

  await action();
}

async function deleteChangeConfirmed(change: Change) {
  if (!change || isChangeDeleting(change.id)) return;

  deletingChangeIds.value = [...deletingChangeIds.value, change.id];
  requirementError.value = '';

  try {
    await deleteChange(change.id);
    await changeCache.loadProjectChanges(change.project_id);
    await projectSelection.loadProjects().catch(() => undefined);

    if (change.id === currentChange.value?.id) {
      void router.push('/changes');
    }
  } catch (err) {
    requirementError.value = err instanceof Error ? err.message : 'Unable to delete change.';
  } finally {
    deletingChangeIds.value = deletingChangeIds.value.filter((id) => id !== change.id);
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
    applyRequirementMutation(mutation.requirements, mutation.change);
    await changeCache.loadProjectChanges(mutation.change.project_id);
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
    applyRequirementMutation(mutation.requirements, mutation.change);
    await changeCache.loadProjectChanges(mutation.change.project_id);
  } catch (err) {
    requirementError.value = err instanceof Error ? err.message : 'Unable to update requirement.';
  } finally {
    updatingRequirementIds.value = updatingRequirementIds.value.filter((id) => id !== requirement.id);
  }
}

function openRequirementEdit(requirement: Requirement) {
  editingRequirement.value = requirement;
  editingRequirementDefinition.value = requirement.definition;
  editingRequirementChangeId.value = requirement.change_id;
  requirementEditOpen.value = true;
}

function closeRequirementEdit() {
  if (savingRequirement.value) return;

  requirementEditOpen.value = false;
  editingRequirement.value = null;
  editingRequirementDefinition.value = '';
  editingRequirementChangeId.value = null;
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
  if (!currentChange.value) return;

  const definition = newRequirementDefinition.value.trim();
  if (!definition) return;

  creatingRequirement.value = true;
  requirementError.value = '';

  try {
    const mutation = await createRequirement(currentChange.value.id, definition);
    applyRequirementMutation(mutation.requirements, mutation.change);
    requirementCreateOpen.value = false;
    newRequirementDefinition.value = '';
    await changeCache.loadProjectChanges(mutation.change.project_id);
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
    let mutation = await updateRequirement({
      id: editingRequirement.value.id,
      definition,
    });
    if (editingRequirementChangeId.value && editingRequirementChangeId.value !== editingRequirement.value.change_id) {
      mutation = await updateRequirementChange(editingRequirement.value.id, editingRequirementChangeId.value);
    }
    applyRequirementMutation(mutation.requirements, mutation.change);
    requirementEditOpen.value = false;
    editingRequirement.value = null;
    editingRequirementDefinition.value = '';
    editingRequirementChangeId.value = null;
    await changeCache.loadProjectChanges(mutation.change.project_id);
    if (mutation.change.id !== changeId.value) await loadChangeDetail();
  } catch (err) {
    requirementError.value = err instanceof Error ? err.message : 'Unable to update requirement.';
  } finally {
    savingRequirement.value = false;
  }
}

function openChangeEdit(changeID: number) {
  void router.push(`/changes/edit/${changeID}`);
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

function completionLabel(change: Change) {
  return `${change.done_req}/${change.total_req} - ${change.completed}%`;
}

function epicName(id: number | null | undefined) {
  if (!id) return 'Standalone';
  return epics.value.find((epic) => epic.id === id)?.name || `#${id}`;
}

onMounted(() => {
  void ensureProjectsLoaded();
  void loadChangeDetail();
});

watch(changeId, () => {
  closeRequirementEdit();
  closeRequirementCreate();
  projectMismatchOpen.value = false;
  projectMismatch.value = null;
  void ensureProjectsLoaded();
  void loadChangeDetail();
});
</script>

<style scoped lang="scss">
.requirement-edit-dialog {
  width: min(720px, calc(100vw - 32px));
}

.requirement-edit-fields {
  display: grid;
  gap: 16px;
}

.change-detail-table {
  margin-top: 16px;
}

.change-actions-cell {
  width: 64px;
}

.change-detail-body-cell {
  min-height: 160px;
}
</style>
