import { defineStore } from 'pinia';
import { ref } from 'vue';
import { listTasks } from '../api/taskApi';
import type { Task } from './task.types';

export const useTaskCacheStore = defineStore('taskCache', () => {
  const tasks = ref<Task[]>([]);
  const projectId = ref(0);
  const loading = ref(false);

  async function loadProjectTasks(nextProjectId: number) {
    loading.value = true;
    try {
      tasks.value = await listTasks(nextProjectId);
      projectId.value = nextProjectId;
      return tasks.value;
    } finally {
      loading.value = false;
    }
  }

  function setTasks(items: Task[], nextProjectId = projectId.value) {
    tasks.value = items;
    projectId.value = nextProjectId;
  }

  function upsertTask(task: Task) {
    if (projectId.value && task.project_id !== projectId.value) return;

    const exists = tasks.value.some((item) => item.id === task.id);
    tasks.value = exists
      ? tasks.value.map((item) => (item.id === task.id ? task : item))
      : [...tasks.value, task];
    projectId.value = task.project_id;
  }

  return {
    tasks,
    projectId,
    loading,
    loadProjectTasks,
    setTasks,
    upsertTask,
  };
});
