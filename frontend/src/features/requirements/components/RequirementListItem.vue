<template>
  <q-item class="requirement-item">
    <q-item-section avatar>
      <q-checkbox
        :model-value="requirement.done"
        @update:model-value="$emit('toggle', requirement, Boolean($event))"
      />
    </q-item-section>
    <q-item-section>
      <q-input
        v-if="isEditing"
        :model-value="editingDefinition"
        dense
        outlined
        autofocus
        @update:model-value="
          (value) => $emit('update:editingDefinition', value == null ? '' : String(value))
        "
      />
      <span v-else>{{ requirement.definition }}</span>
    </q-item-section>
    <q-item-section side>
      <div class="item-actions">
        <template v-if="isEditing">
          <q-btn
            dense
            flat
            round
            icon="check"
            color="primary"
            :disable="!editingDefinition.trim()"
            @click="$emit('save', requirement)"
          >
            <q-tooltip>Save requirement</q-tooltip>
          </q-btn>
          <q-btn dense flat round icon="close" @click="$emit('cancel')">
            <q-tooltip>Cancel</q-tooltip>
          </q-btn>
        </template>
        <template v-else>
          <q-btn dense flat round icon="edit" @click="$emit('edit', requirement)">
            <q-tooltip>Edit requirement</q-tooltip>
          </q-btn>
          <q-btn dense flat round icon="delete" color="negative" @click="$emit('delete', requirement)">
            <q-tooltip>Delete requirement</q-tooltip>
          </q-btn>
        </template>
      </div>
    </q-item-section>
  </q-item>
</template>

<script setup lang="ts">
import type { Requirement } from '../model/requirement.types';

defineProps<{
  requirement: Requirement;
  isEditing: boolean;
  editingDefinition: string;
}>();

defineEmits<{
  'update:editingDefinition': [value: string];
  toggle: [requirement: Requirement, done: boolean];
  edit: [requirement: Requirement];
  save: [requirement: Requirement];
  cancel: [];
  delete: [requirement: Requirement];
}>();
</script>
