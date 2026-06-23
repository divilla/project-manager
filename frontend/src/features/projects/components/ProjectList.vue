<template>
  <q-list bordered separator class="project-list">
    <q-item
      v-for="project in projects"
      :key="project.id"
      clickable
      :active="project.id === selectedProjectId"
      active-class="selected-project"
      @click="$emit('select', project.id)"
    >
      <q-item-section>
        <q-item-label>{{ project.name }}</q-item-label>
      </q-item-section>
      <q-item-section side>
        <div class="item-actions">
          <q-btn dense flat round icon="edit" @click.stop="$emit('rename', project)">
            <q-tooltip>Rename project</q-tooltip>
          </q-btn>
          <q-btn dense flat round icon="delete" color="negative" @click.stop="$emit('delete', project)">
            <q-tooltip>Delete project</q-tooltip>
          </q-btn>
        </div>
      </q-item-section>
    </q-item>
  </q-list>

  <div v-if="!projects.length && !loading" class="empty-state compact-empty">
    <q-icon name="folder_open" size="32px" />
    <span>Create a project to start tracking tasks.</span>
  </div>
</template>

<script setup lang="ts">
import type { Project } from '../model/project.types';

defineProps<{
  projects: Project[];
  selectedProjectId: number;
  loading: boolean;
}>();

defineEmits<{
  select: [projectId: number];
  rename: [project: Project];
  delete: [project: Project];
}>();
</script>
