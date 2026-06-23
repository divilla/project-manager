<template>
  <q-page class="app-page">
    <section class="page-heading">
      <div>
        <h1>Projects</h1>
        <p>Create, rename, select, and safely delete projects.</p>
      </div>
      <q-btn
        color="primary"
        icon="refresh"
        label="Refresh"
        :loading="loading"
        no-caps
        @click="loadAll"
      />
    </section>

    <q-banner v-if="error" class="status-banner status-banner--error" rounded>
      <template #avatar>
        <q-icon name="warning" />
      </template>
      {{ error }}
    </q-banner>

    <section class="projects-shell projects-shell--crud">
      <aside class="project-panel">
        <ProjectCreateForm v-model="projectName" @create="createProjectFromForm" />
        <ProjectList
          :projects="projects"
          :selected-project-id="selectedProjectId"
          :loading="loading"
          @select="selectProject"
          @rename="startProjectRename"
          @delete="removeProject"
        />
      </aside>
    </section>

    <ProjectRenameDialog
      v-model:open="projectDialogOpen"
      v-model:name="projectEditName"
      @save="saveProjectName"
    />
  </q-page>
</template>

<script setup lang="ts">
import ProjectCreateForm from '@/features/projects/components/ProjectCreateForm.vue';
import ProjectList from '@/features/projects/components/ProjectList.vue';
import ProjectRenameDialog from '@/features/projects/components/ProjectRenameDialog.vue';
import { useProjectsPage } from '@/features/projects/composables/useProjectsPage';

const {
  projects,
  selectedProjectId,
  projectName,
  loading,
  error,
  projectDialogOpen,
  projectEditName,
  loadAll,
  selectProject,
  createProjectFromForm,
  startProjectRename,
  saveProjectName,
  removeProject,
} = useProjectsPage({ tasksEnabled: false });
</script>
