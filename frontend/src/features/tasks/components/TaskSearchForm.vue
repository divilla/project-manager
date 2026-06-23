<template>
  <form v-if="visible" class="task-search-row" @submit.prevent="$emit('search')">
    <q-btn
      color="primary"
      icon="add_task"
      type="button"
      :loading="loading"
      no-caps
      label="New Task"
      @click="$emit('new-task')"
    />
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
      clearable
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
      clearable
      label="Phase"
      :options="phaseOptions"
      class="task-select"
      @update:model-value="(value) => $emit('update:taskPhase', value == null ? '' : String(value))"
    />
    <q-btn color="primary" icon="search" type="submit" :loading="loading" no-caps label="Search" />
    <q-btn
      color="negative"
      icon="clear"
      type="button"
      :loading="loading"
      outline
      no-caps
      label="Clear"
      @click="$emit('clear')"
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
  loading: boolean;
}>();

defineEmits<{
  'update:name': [value: string];
  'update:taskType': [value: string];
  'update:taskPhase': [value: string];
  'new-task': [];
  search: [];
  clear: [];
}>();
</script>
