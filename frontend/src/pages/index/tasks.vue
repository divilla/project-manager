<template>
  <q-page class="app-page">
    <section class="page-heading">
      <div>
        <h1>Tasks</h1>
        <p>Task board backed by the selected project and existing database contract.</p>
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

    <main class="task-board">
      <TaskCreateForm
        :visible="Boolean(selectedProject)"
        v-model:name="taskName"
        v-model:task-type="taskType"
        v-model:task-phase="taskPhase"
        :type-options="typeOptions"
        :phase-options="phaseOptions"
        @create="createTaskFromForm"
      />
      <TaskBoard
        :has-selected-project="Boolean(selectedProject)"
        :board-phases="boardPhases"
        :tasks-by-phase="tasksByPhase"
        :phase-options="phaseOptions"
        @open-task="openTask"
        @move-task="moveTask"
        @delete-task="removeTask"
      />
    </main>

    <TaskDetailDialog
      v-model:open="taskDialogOpen"
      v-model:task-edit-name="taskEditName"
      v-model:task-edit-description="taskEditDescription"
      v-model:task-edit-type="taskEditType"
      v-model:requirement-definition="requirementDefinition"
      v-model:editing-requirement-definition="editingRequirementDefinition"
      :task-detail="taskDetail"
      :editing-requirement-id="editingRequirementId"
      :type-options="typeOptions"
      @save-task="saveTask"
      @create-requirement="createRequirementFromForm"
      @toggle-requirement="toggleRequirement"
      @edit-requirement="startRequirementEdit"
      @save-requirement="saveRequirement"
      @cancel-requirement-edit="cancelRequirementEdit"
      @delete-requirement="removeRequirement"
    />
  </q-page>
</template>

<script setup lang="ts">
import { useProjectsPage } from '@/features/projects/composables/useProjectsPage';
import TaskBoard from '@/features/tasks/components/TaskBoard.vue';
import TaskCreateForm from '@/features/tasks/components/TaskCreateForm.vue';
import TaskDetailDialog from '@/features/tasks/components/TaskDetailDialog.vue';

const {
  taskName,
  taskType,
  taskPhase,
  loading,
  error,
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
} = useProjectsPage();
</script>
