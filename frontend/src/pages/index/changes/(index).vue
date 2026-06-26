<template>
  <q-page class="app-page">
    <section class="changes-page-actions">
      <ChangeSearchForm
        :visible="Boolean(currentProject)"
        v-model:title="changeTitle"
        v-model:change-type="changeType"
        v-model:change-phase="changePhase"
        :type-options="typeOptions"
        :phase-options="phaseOptions"
        :loading="loading"
        @new-change="openChangeCreate"
        @search="searchChanges"
        @clear="clearChangeSearch"
      />
    </section>

    <q-banner v-if="error" class="status-banner status-banner--error" rounded>
      <template #avatar>
        <q-icon name="warning" />
      </template>
      {{ error }}
    </q-banner>

    <section v-if="currentProject" class="epic-panel">
      <form class="epic-create-row" @submit.prevent="createEpicFromForm">
        <q-input
          v-model="epicName"
          dense
          outlined
          label="Epic name"
          class="create-input"
          :disable="loading"
        />
        <q-btn
          color="secondary"
          icon="add"
          label="Add Epic"
          type="submit"
          :loading="loading"
          :disable="!epicName.trim()"
          no-caps
        />
      </form>

      <q-markup-table v-if="epics.length" flat bordered class="epic-table">
        <thead>
          <tr>
            <th class="text-right">nr</th>
            <th class="text-left">Epic</th>
            <th class="text-center">Complete</th>
            <th class="text-center">Version</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="epic in epics" :key="epic.id">
            <td class="text-right">#{{ epic.id }}</td>
            <td class="text-left">{{ epic.name }}</td>
            <td class="text-center">{{ epic.done_req }}/{{ epic.total_req }} - {{ epic.completed }}%</td>
            <td class="text-center">{{ epic.version }}</td>
            <td class="text-right">
              <q-btn dense flat round icon="edit" aria-label="Rename epic" @click="startEpicRename(epic)">
                <q-tooltip>Rename epic</q-tooltip>
              </q-btn>
              <q-btn dense flat round icon="delete" color="negative" aria-label="Delete epic" @click="removeEpic(epic)">
                <q-tooltip>Delete epic</q-tooltip>
              </q-btn>
            </td>
          </tr>
        </tbody>
      </q-markup-table>
    </section>

    <main class="change-board">
      <ChangeBoard
        :has-selected-project="Boolean(currentProject)"
        :board-phases="boardPhases"
        :changes-by-phase="changesByPhase"
        :phase-options="phaseOptions"
        @open-change="openChangePage"
        @move-change="moveChange"
        @delete-change="removeChange"
      />
    </main>

    <DeleteConfirmationDialog v-model:open="confirmationDialogOpen" @confirm="confirm" />

    <q-dialog v-model="epicDialogOpen">
      <q-card class="dialog-card">
        <q-card-section>
          <div class="text-subtitle1">Rename Epic</div>
        </q-card-section>
        <q-card-section>
          <q-input v-model="epicEditName" autofocus outlined label="Epic name" />
        </q-card-section>
        <q-card-actions align="right">
          <q-btn flat label="Cancel" no-caps v-close-popup />
          <q-btn color="primary" label="Save" no-caps :disable="!epicEditName.trim()" @click="saveEpicName" />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useProjectsPage } from '@/features/projects/composables/useProjectsPage';
import ChangeBoard from '@/features/changes/components/ChangeBoard.vue';
import ChangeSearchForm from '@/features/changes/components/ChangeSearchForm.vue';
import type { Change } from '@/features/changes/model/change.types';
import DeleteConfirmationDialog from '@/shared/ui/DeleteConfirmationDialog.vue';

const router = useRouter();

const {
  changeTitle,
  changeType,
  changePhase,
  epicName,
  epicEditName,
  epicDialogOpen,
  loading,
  error,
  currentProject,
  epics,
  phaseOptions,
  typeOptions,
  boardPhases,
  changesByPhase,
  confirmationDialogOpen,
  createEpicFromForm,
  startEpicRename,
  saveEpicName,
  searchChanges,
  clearChangeSearch,
  moveChange,
  removeChange,
  removeEpic,
  confirm,
} = useProjectsPage();

function openChangePage(change: Change) {
  void router.push(`/changes/${change.id}`);
}

function openChangeCreate() {
  void router.push('/changes/create');
}
</script>
