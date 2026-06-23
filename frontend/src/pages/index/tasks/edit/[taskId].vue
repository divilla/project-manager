<template>
  <q-page class="app-page">
    <div class="task-create-parent-line">
      <q-btn flat round icon="arrow_back" aria-label="Back to task" @click="goBack" />
      <div>
        <div class="text-caption text-grey-7">Task</div>
        <div v-if="loadedTask" class="text-subtitle1">#{{ loadedTask.id }} {{ loadedTask.name }}</div>
        <div v-else class="text-subtitle1">Loading</div>
      </div>
    </div>

    <q-banner v-if="error" class="status-banner status-banner--error" rounded>
      <template #avatar>
        <q-icon name="warning" />
      </template>
      {{ error }}
    </q-banner>

    <q-form class="task-create-form" @submit.prevent="saveTaskFromPage">
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

      <q-input
        v-model="description"
        outlined
        type="textarea"
        label="Description"
        class="task-description-input"
        input-style="min-height: 600px"
        :disable="loading || saving"
      />

      <div class="task-create-actions">
        <q-btn flat icon="close" label="Cancel" :disable="saving" @click="goBack" />
        <q-btn color="primary" icon="save" label="Save" type="submit" :loading="saving" />
      </div>
    </q-form>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  getTask,
  getTaskReferences,
  updateTask,
  updateTaskPhase,
} from '@/features/tasks/api/taskApi';
import { useTaskCacheStore } from '@/features/tasks/model/taskCache.store';
import type { SelectOption, Task } from '@/features/tasks/model/task.types';

const route = useRoute();
const router = useRouter();
const taskCache = useTaskCacheStore();

const loadedTask = ref<Task | null>(null);
const name = ref('');
const description = ref('');
const taskType = ref('');
const taskPhase = ref('');
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

const taskId = computed(() => {
  const value = Number(route.params.taskId);
  return Number.isInteger(value) && value > 0 ? value : 0;
});

async function loadEditContext() {
  if (!taskId.value) {
    error.value = 'Invalid task.';
    return;
  }

  loading.value = true;
  error.value = '';

  try {
    const [references, detail] = await Promise.all([getTaskReferences(), getTask(taskId.value)]);
    typeOptions.value = references.types.map((type) => ({ label: type.slug, value: type.slug }));
    phaseOptions.value = references.phases.map((phase) => ({ label: phase.slug, value: phase.slug }));
    loadedTask.value = detail.task;
    name.value = detail.task.name;
    description.value = detail.task.description;
    taskType.value = detail.task.task_type;
    taskPhase.value = detail.task.task_phase;
    taskCache.upsertTask(detail.task);
  } catch (err) {
    loadedTask.value = null;
    error.value = err instanceof Error ? err.message : 'Unable to load task.';
  } finally {
    loading.value = false;
  }
}

async function saveTaskFromPage() {
  if (saving.value || !loadedTask.value) return;

  const taskName = name.value.trim();
  if (!taskName) return;

  saving.value = true;
  error.value = '';

  try {
    let task = await updateTask({
      id: loadedTask.value.id,
      name: taskName,
      description: description.value.trim(),
      task_type: taskType.value,
    });
    if (taskPhase.value && taskPhase.value !== task.task_phase) {
      task = await updateTaskPhase(task.id, taskPhase.value);
    }

    await taskCache.loadProjectTasks(task.project_id);
    taskCache.upsertTask(task);
    void router.push(`/tasks/${task.id}`);
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unable to update task.';
  } finally {
    saving.value = false;
  }
}

function goBack() {
  if (loadedTask.value) {
    void router.push(`/tasks/${loadedTask.value.id}`);
    return;
  }

  void router.push('/tasks');
}

onMounted(() => {
  void loadEditContext();
});

watch(taskId, () => {
  void loadEditContext();
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

.task-create-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.task-description-input :deep(.q-field__native) {
  min-height: 600px;
}
</style>
