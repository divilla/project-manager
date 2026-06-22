package task_test

import (
	"aipm/api-tests/shared"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskHistoryFieldsAndSmallintRoundTrip(t *testing.T) {
	client := shared.NewClient(t)
	db := shared.NewDB(t)
	projectID := createProject(t, client)
	defer client.Post(t, "/api/v1/project/delete", map[string]any{"id": projectID}, nil)

	parentA := createTaskRecord(t, client, map[string]any{
		"project_id": projectID,
		"name":       uniqueTaskName("history-parent-a"),
	})
	parentB := createTaskRecord(t, client, map[string]any{
		"project_id": projectID,
		"name":       uniqueTaskName("history-parent-b"),
	})
	current := createTaskRecord(t, client, map[string]any{
		"project_id": projectID,
		"parent_id":  parentA.ID,
		"name":       uniqueTaskName("history-child"),
		"difficulty": int16(7),
		"priority":   int16(11),
	})
	assert.Equal(t, int16(7), current.Difficulty)
	assert.Equal(t, int16(11), current.Priority)
	assert.NotNil(t, current.ParentID)
	assert.Equal(t, parentA.ID, *current.ParentID)

	current = updateTaskRecord(t, client, current, map[string]any{"name": current.Name + "-renamed"})
	shared.AssertHistoryCount(t, db, "task_history", current.ID, false, 1)

	current = updateTaskRecord(t, client, current, map[string]any{"description": "History-bearing description."})
	shared.AssertHistoryCount(t, db, "task_history", current.ID, false, 2)

	var refs taskReferences
	status := client.Post(t, "/api/v1/task/reference", map[string]any{}, &refs)
	require.Equal(t, http.StatusOK, status)
	alternateType := current.TaskType
	for _, option := range refs.Types {
		if option.Slug != current.TaskType {
			alternateType = option.Slug
			break
		}
	}
	require.NotEqual(t, current.TaskType, alternateType)
	current = updateTaskRecord(t, client, current, map[string]any{"task_type": alternateType})
	shared.AssertHistoryCount(t, db, "task_history", current.ID, false, 3)

	current = updateTaskParentRecord(t, client, current, &parentB.ID)
	shared.AssertHistoryCount(t, db, "task_history", current.ID, false, 4)
	assert.Equal(t, parentB.ID, *current.ParentID)

	current = updateTaskPriorityRecord(t, client, current, int16(12))
	assert.Equal(t, int16(12), current.Priority)
	shared.AssertHistoryCount(t, db, "task_history", current.ID, false, 4)

	current = updateTaskDifficultyRecord(t, client, current, int16(8))
	assert.Equal(t, int16(8), current.Difficulty)
	shared.AssertHistoryCount(t, db, "task_history", current.ID, false, 4)
}

func TestTaskPhaseRecalculationAfterInsertUpdateAndDelete(t *testing.T) {
	client := shared.NewClient(t)
	projectID := createProject(t, client)
	defer client.Post(t, "/api/v1/project/delete", map[string]any{"id": projectID}, nil)

	parent := createTaskRecord(t, client, map[string]any{
		"project_id": projectID,
		"name":       uniqueTaskName("phase-parent"),
		"task_phase": "production",
	})
	createTaskRecord(t, client, map[string]any{
		"project_id": projectID,
		"parent_id":  parent.ID,
		"name":       uniqueTaskName("phase-sibling"),
		"task_phase": "production",
	})
	child := createTaskRecord(t, client, map[string]any{
		"project_id": projectID,
		"parent_id":  parent.ID,
		"name":       uniqueTaskName("phase-child"),
		"task_phase": "backlog",
	})
	parent = getTaskRecord(t, client, parent.ID)
	assert.Equal(t, "backlog", parent.TaskPhase)

	var moved task
	status := client.Post(t, "/api/v1/task/update-phase", map[string]any{
		"id": child.ID, "task_phase": "review",
	}, &moved)
	require.Equal(t, http.StatusOK, status)
	parent = getTaskRecord(t, client, parent.ID)
	assert.Equal(t, "review", parent.TaskPhase)

	status = client.Post(t, "/api/v1/task/delete", map[string]any{"id": moved.ID}, nil)
	require.Equal(t, http.StatusNoContent, status)
	parent = getTaskRecord(t, client, parent.ID)
	assert.Equal(t, "production", parent.TaskPhase)
}

func TestTaskParentCanBeClearedWithoutChangingOtherFields(t *testing.T) {
	client := shared.NewClient(t)
	projectID := createProject(t, client)
	defer client.Post(t, "/api/v1/project/delete", map[string]any{"id": projectID}, nil)
	parent := createTaskRecord(t, client, map[string]any{"project_id": projectID, "name": uniqueTaskName("parent")})
	child := createTaskRecord(t, client, map[string]any{"project_id": projectID, "parent_id": parent.ID, "name": uniqueTaskName("child")})
	updated := updateTaskParentRecord(t, client, child, nil)
	assert.Nil(t, updated.ParentID)
	assert.Equal(t, child.Version+1, updated.Version)
}

func TestTaskParentMoveRecalculatesRequirementCounters(t *testing.T) {
	client := shared.NewClient(t)
	projectID := createProject(t, client)
	defer client.Post(t, "/api/v1/project/delete", map[string]any{"id": projectID}, nil)

	parentA := createTaskRecord(t, client, map[string]any{
		"project_id": projectID,
		"name":       uniqueTaskName("counter-parent-a"),
	})
	parentB := createTaskRecord(t, client, map[string]any{
		"project_id": projectID,
		"name":       uniqueTaskName("counter-parent-b"),
	})
	child := createTaskRecord(t, client, map[string]any{
		"project_id": projectID,
		"parent_id":  parentA.ID,
		"name":       uniqueTaskName("counter-child"),
	})
	requirementID := createRequirementRecord(t, client, child.ID, "Counters follow re-parented tasks.")
	status := client.Post(t, "/api/v1/requirement/update-done", map[string]any{
		"id": requirementID, "done": true,
	}, nil)
	require.Equal(t, http.StatusOK, status)

	parentA = getTaskRecord(t, client, parentA.ID)
	assert.Equal(t, int16(1), parentA.DoneReq)
	assert.Equal(t, int16(1), parentA.TotalReq)

	child = updateTaskParentRecord(t, client, child, &parentB.ID)
	assert.Equal(t, parentB.ID, *child.ParentID)

	parentA = getTaskRecord(t, client, parentA.ID)
	parentB = getTaskRecord(t, client, parentB.ID)
	child = getTaskRecord(t, client, child.ID)
	assert.Equal(t, int16(0), parentA.DoneReq)
	assert.Equal(t, int16(0), parentA.TotalReq)
	assert.Equal(t, int16(1), parentB.DoneReq)
	assert.Equal(t, int16(1), parentB.TotalReq)
	assert.Equal(t, int16(1), child.DoneReq)
	assert.Equal(t, int16(1), child.TotalReq)
}

func TestTaskParentCannotBeDescendant(t *testing.T) {
	client := shared.NewClient(t)
	projectID := createProject(t, client)
	defer client.Post(t, "/api/v1/project/delete", map[string]any{"id": projectID}, nil)

	root := createTaskRecord(t, client, map[string]any{
		"project_id": projectID,
		"name":       uniqueTaskName("cycle-root"),
	})
	child := createTaskRecord(t, client, map[string]any{
		"project_id": projectID,
		"parent_id":  root.ID,
		"name":       uniqueTaskName("cycle-child"),
	})
	grandchild := createTaskRecord(t, client, map[string]any{
		"project_id": projectID,
		"parent_id":  child.ID,
		"name":       uniqueTaskName("cycle-grandchild"),
	})

	var updated task
	status := client.Post(t, "/api/v1/task/update-parent", map[string]any{
		"id": root.ID, "parent_id": grandchild.ID,
	}, &updated)
	require.Equal(t, http.StatusBadRequest, status)

	root = getTaskRecord(t, client, root.ID)
	assert.Nil(t, root.ParentID)
}

func createRequirementRecord(t *testing.T, client *shared.Client, taskID int, definition string) int {
	t.Helper()
	var mutation struct {
		Requirement *struct {
			ID int `json:"id"`
		} `json:"requirement"`
	}
	status := client.Post(t, "/api/v1/requirement/create", map[string]any{
		"task_id": taskID, "definition": definition,
	}, &mutation)
	require.Equal(t, http.StatusCreated, status)
	require.NotNil(t, mutation.Requirement)
	return mutation.Requirement.ID
}

func createTaskRecord(t *testing.T, client *shared.Client, payload map[string]any) task {
	t.Helper()
	var created task
	status := client.Post(t, "/api/v1/task/create", payload, &created)
	require.Equal(t, http.StatusCreated, status)
	return created
}

func updateTaskRecord(t *testing.T, client *shared.Client, current task, changes map[string]any) task {
	t.Helper()
	payload := map[string]any{
		"id":          current.ID,
		"name":        current.Name,
		"description": current.Description,
		"task_type":   current.TaskType,
	}
	for key, value := range changes {
		payload[key] = value
	}
	var updated task
	status := client.Post(t, "/api/v1/task/update", payload, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, current.Version+1, updated.Version)
	return updated
}

func updateTaskParentRecord(t *testing.T, client *shared.Client, current task, parentID *int) task {
	t.Helper()
	var updated task
	status := client.Post(t, "/api/v1/task/update-parent", map[string]any{"id": current.ID, "parent_id": parentID}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, current.Version+1, updated.Version)
	return updated
}

func updateTaskPriorityRecord(t *testing.T, client *shared.Client, current task, priority int16) task {
	t.Helper()
	var updated task
	status := client.Post(t, "/api/v1/task/update-priority", map[string]any{"id": current.ID, "priority": priority}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, current.Version, updated.Version)
	return updated
}

func updateTaskDifficultyRecord(t *testing.T, client *shared.Client, current task, difficulty int16) task {
	t.Helper()
	var updated task
	status := client.Post(t, "/api/v1/task/update-difficulty", map[string]any{"id": current.ID, "difficulty": difficulty}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, current.Version, updated.Version)
	return updated
}

func getTaskRecord(t *testing.T, client *shared.Client, id int) task {
	t.Helper()
	var detail taskDetail
	status := client.Post(t, "/api/v1/task/get", map[string]any{"id": id}, &detail)
	require.Equal(t, http.StatusOK, status)
	return detail.Task
}

func uniqueTaskName(prefix string) string {
	return fmt.Sprintf("api-test-%s-%d", prefix, time.Now().UnixNano())
}
