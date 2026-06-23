<template>
  <div v-if="!hasSelectedProject" class="empty-state">
    <q-icon name="view_kanban" size="44px" />
    <span>Select or create a project to view its task board.</span>
  </div>

  <div v-else class="phase-board">
    <section v-for="phase in boardPhases" :key="phase.slug" class="phase-column">
      <header class="phase-heading">
        <span>{{ phase.slug }}</span>
        <q-badge color="grey-7" :label="tasksByPhase[phase.slug]?.length || 0" />
      </header>

      <TaskCard
        v-for="task in tasksByPhase[phase.slug]"
        :key="task.id"
        :task="task"
        :phase-options="phaseOptions"
        @open="$emit('open-task', $event)"
        @move="(item, phase) => $emit('move-task', item, phase)"
        @delete="$emit('delete-task', $event)"
      />

      <div v-if="!tasksByPhase[phase.slug]?.length" class="phase-empty">No tasks</div>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { ReferenceOption, SelectOption, Task } from '../model/task.types';
import TaskCard from './TaskCard.vue';

defineProps<{
  hasSelectedProject: boolean;
  boardPhases: ReferenceOption[];
  tasksByPhase: Record<string, Task[]>;
  phaseOptions: SelectOption[];
}>();

defineEmits<{
  'open-task': [task: Task];
  'move-task': [task: Task, phase: string];
  'delete-task': [task: Task];
}>();
</script>
