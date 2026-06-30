<template>
  <q-card flat bordered class="change-card" @click="$emit('open', change)">
    <q-card-section>
      <div class="change-card-title">{{ change.title }}</div>
      <div class="change-card-meta">#{{ change.ref }} · {{ change.change_types.join(', ') }} · {{ change.completed }}%</div>
      <q-linear-progress :value="change.completed / 100" rounded class="q-mt-sm" />
    </q-card-section>
    <q-card-actions align="between">
      <q-select
        :model-value="change.change_phase"
        dense
        borderless
        emit-value
        map-options
        :options="phaseOptions"
        class="phase-move"
        @click.stop
        @update:model-value="(phase) => $emit('move', change, String(phase))"
      />
      <q-btn dense flat round icon="delete" color="negative" aria-label="Delete change" @click.stop="$emit('delete', change)">
        <q-tooltip>Delete change</q-tooltip>
      </q-btn>
    </q-card-actions>
  </q-card>
</template>

<script setup lang="ts">
import type { SelectOption, Change } from '../model/change.types';

defineProps<{
  change: Change;
  phaseOptions: SelectOption[];
}>();

defineEmits<{
  open: [change: Change];
  move: [change: Change, phase: string];
  delete: [change: Change];
}>();
</script>
