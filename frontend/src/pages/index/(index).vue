<template>
  <q-page class="app-page">
    <section class="page-heading">
      <div>
        <h1>Home</h1>
        <p>Completeness overview and local system status.</p>
      </div>
      <q-btn
        color="primary"
        icon="refresh"
        label="Refresh"
        :loading="loading"
        no-caps
        @click="loadHealth"
      />
    </section>

    <q-banner v-if="error" class="status-banner status-banner--error" rounded>
      <template #avatar>
        <q-icon name="warning" />
      </template>
      {{ error }}
    </q-banner>

    <q-banner v-else-if="health" class="status-banner" rounded>
      <template #avatar>
        <q-icon :name="health.status === 'ok' ? 'check_circle' : 'error'" />
      </template>
      API: {{ health.api }} · Database: {{ health.database }}
    </q-banner>

    <div class="content-grid">
      <q-card flat bordered>
        <q-card-section>
          <div class="text-subtitle1">Project Completeness</div>
          <div class="empty-state">
            <q-icon name="donut_large" size="40px" />
            <span>Dashboard metrics will appear after project and task APIs are connected.</span>
          </div>
        </q-card-section>
      </q-card>

      <q-card flat bordered>
        <q-card-section>
          <div class="text-subtitle1">Phase Summary</div>
          <div class="empty-state">
            <q-icon name="view_column" size="40px" />
            <span>Task phases will be loaded from the existing database reference data.</span>
          </div>
        </q-card-section>
      </q-card>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { getHealth, type HealthResponse } from '@/services/api';

const health = ref<HealthResponse | null>(null);
const loading = ref(false);
const error = ref('');

async function loadHealth() {
  loading.value = true;
  error.value = '';

  try {
    health.value = await getHealth();
  } catch (err) {
    health.value = null;
    error.value = err instanceof Error ? err.message : 'Unable to reach the backend.';
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void loadHealth();
});
</script>
