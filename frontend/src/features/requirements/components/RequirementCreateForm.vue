<template>
  <form class="requirement-create-row" @submit.prevent="$emit('create')">
    <q-input
      :model-value="modelValue"
      dense
      outlined
      label="Requirement"
      class="create-input"
      @update:model-value="(value) => $emit('update:modelValue', value == null ? '' : String(value))"
    />
    <q-btn
      color="primary"
      icon="playlist_add"
      type="submit"
      :disable="!modelValue.trim()"
      round
      unelevated
    >
      <q-tooltip>Add requirement</q-tooltip>
    </q-btn>
  </form>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: string;
}>();

defineEmits<{
  'update:modelValue': [value: string];
  create: [];
}>();
</script>
