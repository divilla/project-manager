import { useProjectSelectionStore } from '@/features/projects/model/projectSelection.store';
import { useTaskCacheStore } from '@/features/tasks/model/taskCache.store';

export async function refreshProjectScope() {
  const projectSelection = useProjectSelectionStore();
  const taskCache = useTaskCacheStore();

  await projectSelection.loadProjects();

  if (projectSelection.currentProjectId) {
    await taskCache.loadProjectTasks(projectSelection.currentProjectId);
    return;
  }

  taskCache.setTasks([], 0);
}
