<template>
  <q-dialog :model-value="open" @update:model-value="$emit('update:open', Boolean($event))">
    <q-card class="dialog-card">
      <q-card-section>
        <div class="text-subtitle1">{{ taskDetail?.task.name }}</div>
        <div v-if="taskDetail" class="task-card-meta">
          {{ taskDetail.task.task_phase }} · {{ taskDetail.task.completed }}%
        </div>
      </q-card-section>
      <q-card-section v-if="taskDetail">
        <q-input
          :model-value="taskEditName"
          outlined
          label="Task name"
          class="q-mb-md"
          @update:model-value="(value) => $emit('update:taskEditName', value == null ? '' : String(value))"
        />
        <q-input
          :model-value="taskEditDescription"
          outlined
          type="textarea"
          label="Description"
          class="q-mb-md"
          @update:model-value="
            (value) => $emit('update:taskEditDescription', value == null ? '' : String(value))
          "
        />
        <q-select
          :model-value="taskEditType"
          outlined
          emit-value
          map-options
          label="Type"
          :options="typeOptions"
          class="q-mb-md"
          @update:model-value="(value) => $emit('update:taskEditType', value == null ? '' : String(value))"
        />
        <div class="requirements-list">
          <div class="requirements-heading">
            <div class="text-subtitle2">Requirements</div>
            <q-badge color="grey-7" :label="`${taskDetail.task.completed}%`" />
          </div>
          <RequirementCreateForm
            :model-value="requirementDefinition"
            @update:model-value="$emit('update:requirementDefinition', $event)"
            @create="$emit('create-requirement')"
          />
          <RequirementList
            :requirements="taskDetail.requirements"
            :editing-requirement-id="editingRequirementId"
            :editing-requirement-definition="editingRequirementDefinition"
            @update:editing-requirement-definition="
              $emit('update:editingRequirementDefinition', $event)
            "
            @toggle="(requirement, done) => $emit('toggle-requirement', requirement, done)"
            @edit="$emit('edit-requirement', $event)"
            @save="$emit('save-requirement', $event)"
            @cancel="$emit('cancel-requirement-edit')"
            @delete="$emit('delete-requirement', $event)"
          />
        </div>
      </q-card-section>
      <q-card-actions align="right">
        <q-btn flat label="Cancel" no-caps v-close-popup />
        <q-btn
          color="primary"
          label="Save"
          no-caps
          :disable="!taskEditName.trim()"
          @click="$emit('save-task')"
        />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import type { Requirement } from '@/features/requirements/model/requirement.types';
import RequirementCreateForm from '@/features/requirements/components/RequirementCreateForm.vue';
import RequirementList from '@/features/requirements/components/RequirementList.vue';
import type { SelectOption, TaskDetail } from '../model/task.types';

defineProps<{
  open: boolean;
  taskDetail: TaskDetail | null;
  taskEditName: string;
  taskEditDescription: string;
  taskEditType: string;
  requirementDefinition: string;
  editingRequirementId: number;
  editingRequirementDefinition: string;
  typeOptions: SelectOption[];
}>();

defineEmits<{
  'update:open': [value: boolean];
  'update:taskEditName': [value: string];
  'update:taskEditDescription': [value: string];
  'update:taskEditType': [value: string];
  'update:requirementDefinition': [value: string];
  'update:editingRequirementDefinition': [value: string];
  'save-task': [];
  'create-requirement': [];
  'toggle-requirement': [requirement: Requirement, done: boolean];
  'edit-requirement': [requirement: Requirement];
  'save-requirement': [requirement: Requirement];
  'cancel-requirement-edit': [];
  'delete-requirement': [requirement: Requirement];
}>();
</script>
