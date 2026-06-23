<template>
  <q-page class="app-page">
    <section class="tasks-page-actions">
      <TaskSearchForm
        :visible="Boolean(currentProject)"
        v-model:name="taskName"
        v-model:task-type="taskType"
        v-model:task-phase="taskPhase"
        :type-options="typeOptions"
        :phase-options="phaseOptions"
        :loading="loading"
        @new-task="openTaskCreate"
        @search="searchTasks"
        @clear="clearTaskSearch"
      />
    </section>

    <q-banner v-if="error" class="status-banner status-banner--error" rounded>
      <template #avatar>
        <q-icon name="warning" />
      </template>
      {{ error }}
    </q-banner>

    <main class="task-board">
      <TaskBoard
        :has-selected-project="Boolean(currentProject)"
        :board-phases="boardPhases"
        :tasks-by-phase="tasksByPhase"
        :phase-options="phaseOptions"
        @open-task="openTaskPage"
        @move-task="moveTask"
        @delete-task="removeTask"
      />
    </main>

    <DeleteConfirmationDialog v-model:open="confirmationDialogOpen" @confirm="confirm" />
  </q-page>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useProjectsPage } from '@/features/projects/composables/useProjectsPage';
import TaskBoard from '@/features/tasks/components/TaskBoard.vue';
import TaskSearchForm from '@/features/tasks/components/TaskSearchForm.vue';
import type { Task } from '@/features/tasks/model/task.types';
import DeleteConfirmationDialog from '@/shared/ui/DeleteConfirmationDialog.vue';

const router = useRouter();

const {
  taskName,
  taskType,
  taskPhase,
  loading,
  error,
  currentProject,
  phaseOptions,
  typeOptions,
  boardPhases,
  tasksByPhase,
  confirmationDialogOpen,
  searchTasks,
  clearTaskSearch,
  moveTask,
  removeTask,
  confirm,
} = useProjectsPage();

function openTaskPage(task: Task) {
  void router.push(`/tasks/${task.id}`);
}

function openTaskCreate() {
  void router.push('/tasks/create/0');
}
</script>
