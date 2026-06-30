<template>
  <q-page class="app-page">
    <div class="change-form-heading">
      <q-btn flat round icon="arrow_back" aria-label="Back to changes" @click="goBack" />
      <div>
        <div class="text-caption text-grey-7">Change</div>
        <div class="text-subtitle1">New change</div>
      </div>
    </div>

    <q-banner v-if="error" class="status-banner status-banner--error" rounded>
      <template #avatar>
        <q-icon name="warning" />
      </template>
      {{ error }}
    </q-banner>

    <q-form class="change-create-form" @submit.prevent="createChangeFromPage">
      <q-input
        v-model="title"
        outlined
        label="Change title"
        :disable="loading || saving"
        :rules="requiredRules"
        autofocus
      />

      <q-select
        v-model="changeTypes"
        outlined
        emit-value
        map-options
        multiple
        label="Types"
        :options="typeOptions"
        :disable="loading || saving"
      />

      <q-select
        v-model="epicId"
        outlined
        emit-value
        map-options
        clearable
        label="Epic"
        :options="epicOptions"
        :disable="loading || saving"
      />

      <q-input
        v-model="requirementBody"
        outlined
        type="textarea"
        label="Requirement body"
        input-style="min-height: 240px"
        :disable="loading || saving"
      />

      <div class="change-create-actions">
        <q-btn flat icon="close" label="Cancel" :disable="saving" @click="goBack" />
        <q-btn color="primary" icon="save" label="Create" type="submit" :loading="saving" />
      </div>
    </q-form>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { storeToRefs } from 'pinia';
import { useRouter } from 'vue-router';
import { createChange, getChangeReferences } from '@/features/changes/api/changeApi';
import { useChangeCacheStore } from '@/features/changes/model/changeCache.store';
import type { ChangeCreateInput, SelectOption } from '@/features/changes/model/change.types';
import { useProjectSelectionStore } from '@/features/projects/model/projectSelection.store';

const router = useRouter();
const projectSelection = useProjectSelectionStore();
const changeCache = useChangeCacheStore();
const { currentProjectId } = storeToRefs(projectSelection);
const { epics } = storeToRefs(changeCache);

const title = ref('');
const requirementBody = ref('');
const changeTypes = ref<string[]>([]);
const epicId = ref<number | null>(null);
const typeOptions = ref<SelectOption[]>([]);
const loading = ref(false);
const saving = ref(false);
const error = ref('');
const requiredRules = [
  (value: unknown) => {
    if (typeof value === 'string') return Boolean(value.trim()) || 'Required';
    return value != null || 'Required';
  },
];

const epicOptions = computed(() =>
  epics.value.map((epic) => ({ label: epic.name, value: epic.id })),
);

async function loadCreateContext() {
  loading.value = true;
  error.value = '';

  try {
    const [references] = await Promise.all([
      getChangeReferences(),
      projectSelection.hasLoaded ? Promise.resolve() : projectSelection.loadProjects(),
    ]);
    typeOptions.value = references.types.map((type) => ({ label: type.slug, value: type.slug }));
    if (!changeTypes.value.length && references.types[0]) changeTypes.value = [references.types[0].slug];
    if (currentProjectId.value) await changeCache.loadProjectChanges(currentProjectId.value);
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unable to load change creation context.';
  } finally {
    loading.value = false;
  }
}

async function createChangeFromPage() {
  if (saving.value) return;

  const changeTitle = title.value.trim();
  const projectId = currentProjectId.value;
  if (!projectId) {
    error.value = 'Select a project before creating a change.';
    return;
  }
  if (!changeTitle || !changeTypes.value.length) return;

  saving.value = true;
  error.value = '';

  try {
    const input: ChangeCreateInput = {
      project_id: projectId,
      epic_id: epicId.value || null,
      title: changeTitle,
      requirement_body: requirementBody.value.trim(),
      change_types: changeTypes.value,
    };

    const change = await createChange(input);
    await Promise.all([
      changeCache.loadProjectChanges(change.project_id),
      projectSelection.loadProjects().catch(() => undefined),
    ]);
    changeCache.upsertChange(change);
    void router.push(`/changes/${change.id}`);
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unable to create change.';
  } finally {
    saving.value = false;
  }
}

function goBack() {
  void router.push('/changes');
}

onMounted(() => {
  void loadCreateContext();
});
</script>
