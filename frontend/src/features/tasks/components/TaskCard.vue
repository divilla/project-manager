<template>
  <q-card flat bordered class="task-card" @click="$emit('open', task)">
    <q-card-section>
      <div class="task-card-title">{{ task.name }}</div>
      <div class="task-card-meta">{{ task.task_type }} · {{ task.completed }}%</div>
      <q-linear-progress :value="task.completed / 100" rounded class="q-mt-sm" />
    </q-card-section>
    <q-card-actions align="between">
      <q-select
        :model-value="task.task_phase"
        dense
        borderless
        emit-value
        map-options
        :options="phaseOptions"
        class="phase-move"
        @click.stop
        @update:model-value="(phase) => $emit('move', task, String(phase))"
      />
      <q-btn dense flat round icon="delete" color="negative" @click.stop="$emit('delete', task)">
        <q-tooltip>Delete task</q-tooltip>
      </q-btn>
    </q-card-actions>
  </q-card>
</template>

<script setup lang="ts">
import type { SelectOption, Task } from '../model/task.types';

defineProps<{
  task: Task;
  phaseOptions: SelectOption[];
}>();

defineEmits<{
  open: [task: Task];
  move: [task: Task, phase: string];
  delete: [task: Task];
}>();
</script>
