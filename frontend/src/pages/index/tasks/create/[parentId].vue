<template>
  <q-page class="app-page">
    <div class="task-create-parent-line">
      <q-btn flat round icon="arrow_back" aria-label="Back to tasks" @click="goBack" />
      <div>
        <div class="text-caption text-grey-7">Parent</div>
        <div v-if="parentTask" class="text-subtitle1">
          #{{ parentTask.id }} {{ parentTask.name }}
        </div>
        <div v-else class="text-subtitle1">None</div>
      </div>
    </div>

    <q-banner v-if="error" class="status-banner status-banner--error" rounded>
      <template #avatar>
        <q-icon name="warning" />
      </template>
      {{ error }}
    </q-banner>

    <q-form class="task-create-form" @submit.prevent="createTaskFromPage">
      <q-input
        v-model="name"
        outlined
        label="Task name"
        :disable="loading || saving"
        :rules="requiredRules"
        autofocus
      />

      <q-select
        v-model="taskType"
        outlined
        emit-value
        map-options
        clearable
        label="Type"
        :options="typeOptions"
        :disable="loading || saving"
      />

      <q-select
        v-model="taskPhase"
        outlined
        emit-value
        map-options
        clearable
        label="Phase"
        :options="phaseOptions"
        :disable="loading || saving"
      />

      <div class="task-create-number-row">
        <q-input
          v-model.number="difficulty"
          outlined
          type="number"
          min="1"
          label="Difficulty"
          :disable="loading || saving"
        />
        <q-input
          v-model.number="priority"
          outlined
          type="number"
          label="Priority"
          :disable="loading || saving"
        />
      </div>

      <q-input
        v-model="description"
        outlined
        type="textarea"
        label="Description"
        :disable="loading || saving"
      />

      <div class="task-create-actions">
        <q-btn flat icon="close" label="Cancel" :disable="saving" @click="goBack" />
        <q-btn
          color="primary"
          icon="save"
          label="Create"
          type="submit"
          :loading="saving"
        />
      </div>
    </q-form>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { storeToRefs } from 'pinia';
import { useRoute, useRouter } from 'vue-router';
import { useProjectSelectionStore } from '@/features/projects/model/projectSelection.store';
import { createTask, getTask, getTaskReferences } from '@/features/tasks/api/taskApi';
import { useTaskCacheStore } from '@/features/tasks/model/taskCache.store';
import type { SelectOption, Task, TaskCreateInput } from '@/features/tasks/model/task.types';

const route = useRoute();
const router = useRouter();
const projectSelection = useProjectSelectionStore();
const taskCache = useTaskCacheStore();
const { currentProjectId } = storeToRefs(projectSelection);

const parentTask = ref<Task | null>(null);
const name = ref('');
const description = ref('');
const taskType = ref('');
const taskPhase = ref('');
const difficulty = ref(1);
const priority = ref(0);
const typeOptions = ref<SelectOption[]>([]);
const phaseOptions = ref<SelectOption[]>([]);
const loading = ref(false);
const saving = ref(false);
const error = ref('');
const requiredRules = [
  (value: unknown) => {
    if (typeof value === 'string') return Boolean(value.trim()) || 'Required';
    return value != null || 'Required';
  },
];

const parentId = computed(() => {
  const value = Number(route.params.parentId);
  return Number.isInteger(value) && value > 0 ? value : 0;
});

async function loadCreateContext() {
  loading.value = true;
  error.value = '';

  try {
    const [references] = await Promise.all([
      getTaskReferences(),
      projectSelection.hasLoaded ? Promise.resolve() : projectSelection.loadProjects(),
    ]);
    typeOptions.value = references.types.map((type) => ({ label: type.slug, value: type.slug }));
    phaseOptions.value = references.phases.map((phase) => ({ label: phase.slug, value: phase.slug }));
    if (!taskType.value) taskType.value = references.types[0]?.slug || '';
    if (!taskPhase.value) taskPhase.value = references.phases[0]?.slug || '';

    if (parentId.value) {
      const detail = await getTask(parentId.value);
      parentTask.value = detail.task;
      return;
    }

    parentTask.value = null;
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unable to load task creation context.';
  } finally {
    loading.value = false;
  }
}

async function createTaskFromPage() {
  if (saving.value) return;

  const taskName = name.value.trim();
  const projectId = parentTask.value?.project_id || currentProjectId.value;
  if (!projectId) {
    error.value = 'Select a project before creating a task.';
    return;
  }
  if (!taskName) return;

  saving.value = true;
  error.value = '';

  try {
    const input: TaskCreateInput = {
      project_id: projectId,
      name: taskName,
      difficulty: Number(difficulty.value) || 1,
      priority: Number(priority.value) || 0,
    };
    const trimmedDescription = description.value.trim();
    if (trimmedDescription) input.description = trimmedDescription;
    if (taskPhase.value) input.task_phase = taskPhase.value;
    if (taskType.value) input.task_type = taskType.value;
    if (parentTask.value) input.parent_id = parentTask.value.id;

    const task = await createTask(input);
    await Promise.all([
      taskCache.loadProjectTasks(task.project_id),
      projectSelection.loadProjects().catch(() => undefined),
    ]);
    taskCache.upsertTask(task);
    void router.push(`/tasks/${task.id}`);
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unable to create task.';
  } finally {
    saving.value = false;
  }
}

function goBack() {
  if (parentTask.value) {
    void router.push(`/tasks/${parentTask.value.id}`);
    return;
  }

  void router.push('/tasks');
}

onMounted(() => {
  void loadCreateContext();
});

watch(parentId, () => {
  void loadCreateContext();
});
</script>

<style scoped lang="scss">
.task-create-parent-line {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.task-create-form {
  display: grid;
  gap: 16px;
  max-width: 840px;
}

.task-create-number-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.task-create-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 640px) {
  .task-create-number-row {
    grid-template-columns: 1fr;
  }
}
</style>
