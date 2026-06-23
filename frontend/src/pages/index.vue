<template>
  <q-layout view="hHh Lpr lFf">
    <q-header class="app-header">
      <q-toolbar class="app-toolbar">
        <q-toolbar-title class="app-title">AI Project Manager</q-toolbar-title>

        <ProjectSelector dark :show-hint="false" class="top-project-selector" />

        <div class="desktop-tabs-wrap">
          <q-tabs :model-value="activeTab" shrink stretch class="desktop-tabs">
            <q-route-tab name="home" to="/" label="Home" exact />
            <q-route-tab name="tasks" to="/tasks" label="Tasks" />
            <q-route-tab name="projects" to="/projects" label="Projects" />
            <q-route-tab name="planning" to="/planning" label="Planning" />
            <q-route-tab name="help" to="/help" label="Help" />
          </q-tabs>
        </div>

        <q-space />

        <q-btn
          flat
          dense
          round
          icon="menu"
          class="mobile-menu"
          aria-label="Open navigation"
          @click="drawerOpen = true"
        />
      </q-toolbar>
    </q-header>

    <q-drawer v-model="drawerOpen" side="right" bordered :width="240">
      <q-list padding>
        <q-item clickable to="/" exact @click="drawerOpen = false">
          <q-item-section avatar><q-icon name="dashboard" /></q-item-section>
          <q-item-section>Home</q-item-section>
        </q-item>
        <q-item clickable to="/tasks" @click="drawerOpen = false">
          <q-item-section avatar><q-icon name="task_alt" /></q-item-section>
          <q-item-section>Tasks</q-item-section>
        </q-item>
        <q-item clickable to="/projects" @click="drawerOpen = false">
          <q-item-section avatar><q-icon name="view_kanban" /></q-item-section>
          <q-item-section>Projects</q-item-section>
        </q-item>
        <q-item clickable to="/planning" @click="drawerOpen = false">
          <q-item-section avatar><q-icon name="psychology" /></q-item-section>
          <q-item-section>Planning</q-item-section>
        </q-item>
        <q-item clickable to="/help" @click="drawerOpen = false">
          <q-item-section avatar><q-icon name="help" /></q-item-section>
          <q-item-section>Help</q-item-section>
        </q-item>
      </q-list>
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { storeToRefs } from 'pinia';
import { useRoute, useRouter } from 'vue-router';
import ProjectSelector from '@/features/projects/components/ProjectSelector.vue';
import { useProjectSelectionStore } from '@/features/projects/model/projectSelection.store';

const route = useRoute();
const router = useRouter();
const drawerOpen = ref(false);
const projectSelection = useProjectSelectionStore();
const { projects, hasLoaded } = storeToRefs(projectSelection);

const activeTab = computed(() => {
  if (route.path.startsWith('/planning')) return 'planning';
  if (route.path.startsWith('/projects')) return 'projects';
  if (route.path.startsWith('/tasks')) return 'tasks';
  if (route.path.startsWith('/help')) return 'help';
  return 'home';
});

onMounted(() => {
  void projectSelection.loadProjects().catch(() => undefined);
});

watch([hasLoaded, projects, () => route.path], () => {
  if (hasLoaded.value && !projects.value.length && route.path !== '/projects') {
    void router.push('/projects');
  }
});
</script>
