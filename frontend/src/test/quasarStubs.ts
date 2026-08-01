import type { Component } from 'vue';

type QuasarStubs = Record<string, Component | boolean>;

export function createQuasarStubs(overrides: QuasarStubs = {}) {
  return {
    QBanner: { template: '<div><slot name="avatar" /><slot /></div>' },
    QBtn: {
      emits: ['click'],
      props: ['color', 'disable', 'flat', 'icon', 'label', 'loading'],
      template: `
        <button
          :data-color="color"
          :data-icon="icon"
          :disabled="disable || loading"
          type="button"
          @click="$emit('click', $event)"
        >
          {{ label }}<slot />
        </button>
      `,
    },
    QForm: { template: '<form @submit.prevent="$emit(\'submit\', $event)"><slot /></form>' },
    QIcon: { props: ['name'], template: '<span :data-icon="name"><slot /></span>' },
    QInput: {
      emits: ['update:modelValue'],
      props: ['disable', 'label', 'modelValue', 'type'],
      template: `
        <label>
          <span>{{ label }}</span>
          <textarea
            v-if="type === 'textarea'"
            :aria-label="label"
            :disabled="disable"
            :value="modelValue"
            @input="$emit('update:modelValue', $event.target.value)"
          />
          <input
            v-else
            :aria-label="label"
            :disabled="disable"
            :value="modelValue"
            @input="$emit('update:modelValue', $event.target.value)"
          />
        </label>
      `,
    },
    QPage: { template: '<main><slot /></main>' },
    QSelect: {
      emits: ['update:modelValue'],
      props: ['disable', 'label', 'modelValue', 'multiple', 'options'],
      template: `
        <select
          :aria-label="label"
          :disabled="disable"
          :multiple="multiple"
          :value="modelValue"
          @change="$emit('update:modelValue', Number($event.target.value))"
        >
          <option v-for="option in options" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
      `,
    },
    ...overrides,
  };
}
