<template>
  <form v-if="visible" class="task-create-row" @submit.prevent="$emit('create')">
    <q-input
      :model-value="name"
      dense
      outlined
      label="Task name"
      class="create-input"
      @update:model-value="(value) => $emit('update:name', value == null ? '' : String(value))"
    />
    <q-select
      :model-value="taskType"
      dense
      outlined
      emit-value
      map-options
      label="Type"
      :options="typeOptions"
      class="task-select"
      @update:model-value="(value) => $emit('update:taskType', value == null ? '' : String(value))"
    />
    <q-select
      :model-value="taskPhase"
      dense
      outlined
      emit-value
      map-options
      label="Phase"
      :options="phaseOptions"
      class="task-select"
      @update:model-value="(value) => $emit('update:taskPhase', value == null ? '' : String(value))"
    />
    <q-btn
      color="primary"
      icon="add_task"
      type="submit"
      :disable="!name.trim()"
      no-caps
      label="Task"
    />
  </form>
</template>

<script setup lang="ts">
import type { SelectOption } from '../model/task.types';

defineProps<{
  visible: boolean;
  name: string;
  taskType: string;
  taskPhase: string;
  typeOptions: SelectOption[];
  phaseOptions: SelectOption[];
}>();

defineEmits<{
  'update:name': [value: string];
  'update:taskType': [value: string];
  'update:taskPhase': [value: string];
  create: [];
}>();
</script>
