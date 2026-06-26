<template>
  <div v-if="!hasSelectedProject" class="empty-state">
    <q-icon name="view_kanban" size="44px" />
    <span>Select or create a project to view its change board.</span>
  </div>

  <div v-else class="phase-board">
    <section v-for="phase in boardPhases" :key="phase.slug" class="phase-column">
      <header class="phase-heading">
        <span>{{ phase.slug }}</span>
        <q-badge color="grey-7" :label="changesByPhase[phase.slug]?.length || 0" />
      </header>

      <ChangeCard
        v-for="change in changesByPhase[phase.slug]"
        :key="change.id"
        :change="change"
        :phase-options="phaseOptions"
        @open="$emit('open-change', $event)"
        @move="(item, phase) => $emit('move-change', item, phase)"
        @delete="$emit('delete-change', $event)"
      />

      <div v-if="!changesByPhase[phase.slug]?.length" class="phase-empty">No changes</div>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { ReferenceOption, SelectOption, Change } from '../model/change.types';
import ChangeCard from './ChangeCard.vue';

defineProps<{
  hasSelectedProject: boolean;
  boardPhases: ReferenceOption[];
  changesByPhase: Record<string, Change[]>;
  phaseOptions: SelectOption[];
}>();

defineEmits<{
  'open-change': [change: Change];
  'move-change': [change: Change, phase: string];
  'delete-change': [change: Change];
}>();
</script>
