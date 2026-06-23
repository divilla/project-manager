package task

import (
	"context"
	"errors"

	"aipm/internal/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Repo struct {
		pool *pgxpool.Pool
	}

	Repository interface {
		References(ctx context.Context) (dto.TaskReferences, error)
		List(ctx context.Context, projectID int) ([]dto.Task, error)
		Get(ctx context.Context, id int) (dto.TaskDetail, error)
		Descriptions(ctx context.Context, ids []int) ([]dto.Task, error)
		Create(ctx context.Context, req dto.TaskCreateRequest) (dto.Task, error)
		Update(ctx context.Context, req dto.TaskUpdateRequest) (dto.Task, error)
		UpdateDifficulty(ctx context.Context, req dto.TaskUpdateDifficultyRequest) (dto.Task, error)
		UpdatePriority(ctx context.Context, req dto.TaskUpdatePriorityRequest) (dto.Task, error)
		UpdateParent(ctx context.Context, req dto.TaskUpdateParentRequest) (dto.Task, error)
		UpdatePhase(ctx context.Context, req dto.TaskUpdatePhaseRequest) (dto.Task, error)
		Delete(ctx context.Context, req dto.TaskIDRequest) error
	}
)

const taskColumns = `
	id,
	version,
	task_type,
	name,
	coalesce(description, ''),
	difficulty,
	priority,
	task_phase,
	parent_id,
	project_id,
	done_req,
	total_req,
	completed,
	created,
	modified`

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) References(ctx context.Context) (dto.TaskReferences, error) {
	phases, err := r.referenceOptions(ctx, "task_phase")
	if err != nil {
		return dto.TaskReferences{}, err
	}
	types, err := r.referenceOptions(ctx, "task_type")
	if err != nil {
		return dto.TaskReferences{}, err
	}
	return dto.TaskReferences{Phases: phases, Types: types}, nil
}

func (r *Repo) List(ctx context.Context, projectID int) ([]dto.Task, error) {
	rows, err := r.pool.Query(ctx, `
		select `+taskColumns+`
		from public.vw_task
		where project_id = $1
		order by (select priority from task_phase where slug = vw_task.task_phase), priority, created
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]dto.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *Repo) Get(ctx context.Context, id int) (dto.TaskDetail, error) {
	task, err := getTask(ctx, r.pool, id)
	if err != nil {
		return dto.TaskDetail{}, err
	}
	requirements, err := listRequirements(ctx, r.pool, id)
	if err != nil {
		return dto.TaskDetail{}, err
	}
	return dto.TaskDetail{Task: task, Requirements: requirements}, nil
}

func (r *Repo) Descriptions(ctx context.Context, ids []int) ([]dto.Task, error) {
	rows, err := r.pool.Query(ctx, `
		select requested.id::integer, coalesce(task.description, '')
		from unnest($1::bigint[]) with ordinality as requested(id, ord)
		join public.task task on task.id = requested.id
		order by requested.ord
	`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]dto.Task, 0, len(ids))
	for rows.Next() {
		var task dto.Task
		if err := rows.Scan(&task.ID, &task.Description); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *Repo) Create(ctx context.Context, req dto.TaskCreateRequest) (dto.Task, error) {
	if err := r.ensureProject(ctx, req.ProjectID); err != nil {
		return dto.Task{}, err
	}
	if req.TaskPhase != "" {
		if err := r.ensureReference(ctx, "task_phase", req.TaskPhase); err != nil {
			return dto.Task{}, err
		}
	}
	if req.TaskType != "" {
		if err := r.ensureReference(ctx, "task_type", req.TaskType); err != nil {
			return dto.Task{}, err
		}
	}
	if req.ParentID != nil {
		if err := r.ensureParentTask(ctx, req.ProjectID, *req.ParentID); err != nil {
			return dto.Task{}, err
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Task{}, err
	}
	defer tx.Rollback(ctx)

	var id int
	err = tx.QueryRow(ctx, `
		insert into public.task (
			project_id, name, description, task_phase, task_type, difficulty, priority, parent_id
		)
		values (
			$1, $2, nullif($3, ''), coalesce(nullif($4, ''), 'backlog'),
			coalesce(nullif($5, ''), 'task'), case when $6 > 0 then $6 else 1 end, $7, $8
		)
		returning id
	`, req.ProjectID, req.Name, req.Description, req.TaskPhase, req.TaskType, req.Difficulty, req.Priority, req.ParentID).Scan(&id)
	if err != nil {
		return dto.Task{}, err
	}
	if _, err := tx.Exec(ctx, "call public.sp_task_phase_recalculate($1)", req.ParentID); err != nil {
		return dto.Task{}, err
	}
	task, err := getTask(ctx, tx, id)
	if err != nil {
		return dto.Task{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return dto.Task{}, err
	}
	return task, nil
}

func (r *Repo) Update(ctx context.Context, req dto.TaskUpdateRequest) (dto.Task, error) {
	if req.TaskType != "" {
		if err := r.ensureReference(ctx, "task_type", req.TaskType); err != nil {
			return dto.Task{}, err
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Task{}, err
	}
	defer tx.Rollback(ctx)

	current, err := getTaskState(ctx, tx, req.ID)
	if err != nil {
		return dto.Task{}, err
	}

	taskType := current.TaskType
	if req.TaskType != "" {
		taskType = req.TaskType
	}
	historyChanged := taskType != current.TaskType || req.Name != current.Name || req.Description != current.Description
	if historyChanged {
		if _, err := tx.Exec(ctx, "call public.sp_task_to_history($1, false)", req.ID); err != nil {
			return dto.Task{}, err
		}
	}

	query := `
		update public.task
		set name = $2,
			description = nullif($3, ''),
			task_type = $4,
			modified = now()
		where id = $1
	`
	if historyChanged {
		query = `
			update public.task
			set name = $2,
				description = nullif($3, ''),
				task_type = $4,
				version = version + 1,
				modified = now()
			where id = $1
		`
	}
	tag, err := tx.Exec(ctx, query, req.ID, req.Name, req.Description, taskType)
	if err != nil {
		return dto.Task{}, err
	}
	if tag.RowsAffected() == 0 {
		return dto.Task{}, ErrNotFound
	}
	if _, err := tx.Exec(ctx, "call public.sp_task_phase_recalculate($1)", current.ParentID); err != nil {
		return dto.Task{}, err
	}
	return finishTaskMutation(ctx, tx, req.ID)
}

func (r *Repo) UpdateDifficulty(ctx context.Context, req dto.TaskUpdateDifficultyRequest) (dto.Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Task{}, err
	}
	defer tx.Rollback(ctx)
	current, err := getTaskState(ctx, tx, req.ID)
	if err != nil {
		return dto.Task{}, err
	}
	tag, err := tx.Exec(ctx, "update public.task set difficulty = $2, modified = now() where id = $1", req.ID, req.Difficulty)
	if err != nil {
		return dto.Task{}, err
	}
	if tag.RowsAffected() == 0 {
		return dto.Task{}, ErrNotFound
	}
	if _, err := tx.Exec(ctx, "call public.sp_task_phase_recalculate($1)", current.ParentID); err != nil {
		return dto.Task{}, err
	}
	return finishTaskMutation(ctx, tx, req.ID)
}

func (r *Repo) UpdatePriority(ctx context.Context, req dto.TaskUpdatePriorityRequest) (dto.Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Task{}, err
	}
	defer tx.Rollback(ctx)
	current, err := getTaskState(ctx, tx, req.ID)
	if err != nil {
		return dto.Task{}, err
	}
	tag, err := tx.Exec(ctx, "update public.task set priority = $2, modified = now() where id = $1", req.ID, req.Priority)
	if err != nil {
		return dto.Task{}, err
	}
	if tag.RowsAffected() == 0 {
		return dto.Task{}, ErrNotFound
	}
	if _, err := tx.Exec(ctx, "call public.sp_task_phase_recalculate($1)", current.ParentID); err != nil {
		return dto.Task{}, err
	}
	return finishTaskMutation(ctx, tx, req.ID)
}

func (r *Repo) UpdateParent(ctx context.Context, req dto.TaskUpdateParentRequest) (dto.Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Task{}, err
	}
	defer tx.Rollback(ctx)
	current, err := getTaskState(ctx, tx, req.ID)
	if err != nil {
		return dto.Task{}, err
	}
	if req.ParentID != nil {
		if *req.ParentID == req.ID {
			return dto.Task{}, ErrInvalidReference
		}
		if err := ensureParentTask(ctx, tx, current.ProjectID, *req.ParentID); err != nil {
			return dto.Task{}, err
		}
		isDescendant, err := isTaskDescendant(ctx, tx, req.ID, *req.ParentID)
		if err != nil {
			return dto.Task{}, err
		}
		if isDescendant {
			return dto.Task{}, ErrInvalidReference
		}
	}
	if equalIntPointers(current.ParentID, req.ParentID) {
		return finishTaskMutation(ctx, tx, req.ID)
	}
	if _, err := tx.Exec(ctx, "call public.sp_task_to_history($1, false)", req.ID); err != nil {
		return dto.Task{}, err
	}
	tag, err := tx.Exec(ctx, `
		update public.task
		set parent_id = $2, version = version + 1, modified = now()
		where id = $1
	`, req.ID, req.ParentID)
	if err != nil {
		return dto.Task{}, err
	}
	if tag.RowsAffected() == 0 {
		return dto.Task{}, ErrNotFound
	}
	if err := recalculateTaskRequirements(ctx, tx, current.ParentID, &req.ID); err != nil {
		return dto.Task{}, err
	}
	for _, parentID := range uniqueParentIDs(current.ParentID, req.ParentID) {
		if _, err := tx.Exec(ctx, "call public.sp_task_phase_recalculate($1)", parentID); err != nil {
			return dto.Task{}, err
		}
	}
	return finishTaskMutation(ctx, tx, req.ID)
}

func (r *Repo) UpdatePhase(ctx context.Context, req dto.TaskUpdatePhaseRequest) (dto.Task, error) {
	if err := r.ensureReference(ctx, "task_phase", req.TaskPhase); err != nil {
		return dto.Task{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Task{}, err
	}
	defer tx.Rollback(ctx)
	current, err := getTaskState(ctx, tx, req.ID)
	if err != nil {
		return dto.Task{}, err
	}
	tag, err := tx.Exec(ctx, "update public.task set task_phase = $2, modified = now() where id = $1", req.ID, req.TaskPhase)
	if err != nil {
		return dto.Task{}, err
	}
	if tag.RowsAffected() == 0 {
		return dto.Task{}, ErrNotFound
	}
	if _, err := tx.Exec(ctx, "call public.sp_task_phase_recalculate($1)", current.ParentID); err != nil {
		return dto.Task{}, err
	}
	return finishTaskMutation(ctx, tx, req.ID)
}

func (r *Repo) Delete(ctx context.Context, req dto.TaskIDRequest) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := getTaskState(ctx, tx, req.ID); err != nil {
		return err
	}
	taskIDs, err := taskTreeIDs(ctx, tx, req.ID)
	if err != nil {
		return err
	}
	requirements, err := requirementsForTasks(ctx, tx, taskIDs)
	if err != nil {
		return err
	}
	parentIDs, err := parentsForTasks(ctx, tx, taskIDs)
	if err != nil {
		return err
	}
	for _, requirement := range requirements {
		if _, err := tx.Exec(ctx, "call public.sp_requirement_to_history($1, true)", requirement.ID); err != nil {
			return err
		}
	}
	for _, id := range taskIDs {
		if _, err := tx.Exec(ctx, "call public.sp_task_to_history($1, true)", id); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, "delete from public.requirement where task_id = any($1::bigint[])", taskIDs); err != nil {
		return err
	}
	for _, taskID := range uniqueRequirementTaskIDs(requirements) {
		if _, err := tx.Exec(ctx, "call public.sp_task_requirement_recalculate($1)", taskID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, "delete from public.task where id = any($1::bigint[])", taskIDs); err != nil {
		return err
	}
	for _, parentID := range parentIDs {
		if _, err := tx.Exec(ctx, "call public.sp_task_phase_recalculate($1)", parentID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repo) referenceOptions(ctx context.Context, table string) ([]dto.ReferenceOption, error) {
	rows, err := r.pool.Query(ctx, "select slug, priority from public."+table+" order by priority, slug")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	options := make([]dto.ReferenceOption, 0)
	for rows.Next() {
		var option dto.ReferenceOption
		if err := rows.Scan(&option.Slug, &option.Priority); err != nil {
			return nil, err
		}
		options = append(options, option)
	}
	return options, rows.Err()
}

func (r *Repo) ensureReference(ctx context.Context, table, slug string) error {
	var exists bool
	if err := r.pool.QueryRow(ctx, "select exists(select 1 from public."+table+" where slug = $1)", slug).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrInvalidReference
	}
	return nil
}

func (r *Repo) ensureProject(ctx context.Context, id int) error {
	var exists bool
	if err := r.pool.QueryRow(ctx, "select exists(select 1 from public.project where id = $1)", id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrInvalidReference
	}
	return nil
}

func (r *Repo) ensureParentTask(ctx context.Context, projectID, parentID int) error {
	return ensureParentTask(ctx, r.pool, projectID, parentID)
}

func ensureParentTask(ctx context.Context, q queryer, projectID, parentID int) error {
	var exists bool
	if err := q.QueryRow(ctx, `
		select exists(select 1 from public.task where id = $1 and project_id = $2)
	`, parentID, projectID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrInvalidReference
	}
	return nil
}

func getTask(ctx context.Context, q queryer, id int) (dto.Task, error) {
	task, err := scanTask(q.QueryRow(ctx, "select "+taskColumns+" from public.vw_task where id = $1", id))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Task{}, ErrNotFound
	}
	return task, err
}

func listRequirements(ctx context.Context, q queryer, taskID int) ([]dto.Requirement, error) {
	rows, err := q.Query(ctx, `
		select id, version, definition, done, task_id, created, modified
		from public.requirement where task_id = $1 order by created, definition
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requirements := make([]dto.Requirement, 0)
	for rows.Next() {
		var item dto.Requirement
		if err := rows.Scan(&item.ID, &item.Version, &item.Definition, &item.Done, &item.TaskID, &item.Created, &item.Modified); err != nil {
			return nil, err
		}
		requirements = append(requirements, item)
	}
	return requirements, rows.Err()
}

func scanTask(row pgx.Row) (dto.Task, error) {
	var task dto.Task
	var parentID pgtype.Int8
	err := row.Scan(
		&task.ID, &task.Version, &task.TaskType, &task.Name, &task.Description,
		&task.Difficulty, &task.Priority, &task.TaskPhase, &parentID, &task.ProjectID,
		&task.DoneReq, &task.TotalReq, &task.Completed, &task.Created, &task.Modified,
	)
	if err != nil {
		return dto.Task{}, err
	}
	if parentID.Valid {
		value := int(parentID.Int64)
		task.ParentID = &value
	}
	return task, nil
}

type taskState struct {
	TaskType    string
	Name        string
	Description string
	ParentID    *int
	ProjectID   int
}

func getTaskState(ctx context.Context, tx pgx.Tx, id int) (taskState, error) {
	var state taskState
	var parentID pgtype.Int8
	err := tx.QueryRow(ctx, `
		select task_type, name, coalesce(description, ''), parent_id, project_id
		from public.task where id = $1
	`, id).Scan(&state.TaskType, &state.Name, &state.Description, &parentID, &state.ProjectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return taskState{}, ErrNotFound
	}
	if err != nil {
		return taskState{}, err
	}
	if parentID.Valid {
		value := int(parentID.Int64)
		state.ParentID = &value
	}
	return state, nil
}

func finishTaskMutation(ctx context.Context, tx pgx.Tx, id int) (dto.Task, error) {
	task, err := getTask(ctx, tx, id)
	if err != nil {
		return dto.Task{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return dto.Task{}, err
	}
	return task, nil
}

func isTaskDescendant(ctx context.Context, tx pgx.Tx, taskID, candidateParentID int) (bool, error) {
	var isDescendant bool
	err := tx.QueryRow(ctx, `
		select $2 = any(public.fn_task_descendants($1, ARRAY[]::bigint[]))
	`, taskID, candidateParentID).Scan(&isDescendant)
	return isDescendant, err
}

func recalculateTaskRequirements(ctx context.Context, tx pgx.Tx, ids ...*int) error {
	seen := make(map[int]struct{})
	for _, id := range ids {
		if id == nil {
			continue
		}
		if _, ok := seen[*id]; ok {
			continue
		}
		seen[*id] = struct{}{}
		if _, err := tx.Exec(ctx, "call public.sp_task_requirement_recalculate($1)", *id); err != nil {
			return err
		}
	}
	return nil
}

func taskTreeIDs(ctx context.Context, tx pgx.Tx, taskID int) ([]int, error) {
	rows, err := tx.Query(ctx, `
		with recursive task_tree as (
			select id from public.task where id = $1
			union all
			select child.id from public.task child join task_tree parent on child.parent_id = parent.id
		)
		select id from task_tree
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

type requirementRef struct {
	ID     int
	TaskID int
}

func requirementsForTasks(ctx context.Context, tx pgx.Tx, taskIDs []int) ([]requirementRef, error) {
	rows, err := tx.Query(ctx, `
		select id, task_id from public.requirement where task_id = any($1::bigint[])
	`, taskIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]requirementRef, 0)
	for rows.Next() {
		var item requirementRef
		if err := rows.Scan(&item.ID, &item.TaskID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func parentsForTasks(ctx context.Context, tx pgx.Tx, taskIDs []int) ([]*int, error) {
	rows, err := tx.Query(ctx, "select parent_id from public.task where id = any($1::bigint[])", taskIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	parents := make([]*int, 0)
	for rows.Next() {
		var parentID pgtype.Int8
		if err := rows.Scan(&parentID); err != nil {
			return nil, err
		}
		if !parentID.Valid {
			parents = append(parents, nil)
			continue
		}
		value := int(parentID.Int64)
		parents = append(parents, &value)
	}
	return parents, rows.Err()
}

func uniqueRequirementTaskIDs(requirements []requirementRef) []int {
	seen := make(map[int]struct{})
	ids := make([]int, 0)
	for _, requirement := range requirements {
		if _, ok := seen[requirement.TaskID]; ok {
			continue
		}
		seen[requirement.TaskID] = struct{}{}
		ids = append(ids, requirement.TaskID)
	}
	return ids
}

func equalIntPointers(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func uniqueParentIDs(values ...*int) []*int {
	seen := make(map[int]struct{})
	result := make([]*int, 0, len(values))
	hasNil := false
	for _, value := range values {
		if value == nil {
			if !hasNil {
				result = append(result, nil)
				hasNil = true
			}
			continue
		}
		if _, ok := seen[*value]; ok {
			continue
		}
		seen[*value] = struct{}{}
		result = append(result, value)
	}
	return result
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
