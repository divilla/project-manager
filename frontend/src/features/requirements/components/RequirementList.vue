<template>
  <q-list v-if="requirements.length" bordered separator>
    <RequirementListItem
      v-for="requirement in requirements"
      :key="requirement.id"
      :requirement="requirement"
      :is-editing="editingRequirementId === requirement.id"
      :editing-definition="editingRequirementDefinition"
      @update:editing-definition="$emit('update:editingRequirementDefinition', $event)"
      @toggle="(item, done) => $emit('toggle', item, done)"
      @edit="$emit('edit', $event)"
      @save="$emit('save', $event)"
      @cancel="$emit('cancel')"
      @delete="$emit('delete', $event)"
    />
  </q-list>
  <div v-else class="phase-empty">No requirements yet</div>
</template>

<script setup lang="ts">
import type { Requirement } from '../model/requirement.types';
import RequirementListItem from './RequirementListItem.vue';

defineProps<{
  requirements: Requirement[];
  editingRequirementId: number;
  editingRequirementDefinition: string;
}>();

defineEmits<{
  'update:editingRequirementDefinition': [value: string];
  toggle: [requirement: Requirement, done: boolean];
  edit: [requirement: Requirement];
  save: [requirement: Requirement];
  cancel: [];
  delete: [requirement: Requirement];
}>();
</script>
