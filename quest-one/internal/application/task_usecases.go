package application

import (
	"context"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/quest-one/quest-one/internal/domain"
	"github.com/quest-one/quest-one/internal/domain/services"
	"github.com/quest-one/quest-one/internal/ports"
)

// ---- Add ----

// AddTaskInput is the input DTO for adding a new task.
type AddTaskInput struct {
	Title       string
	Description string
	Tags        []string
	ProjectID   *string
	DueDate     *time.Time
}

// AddTask creates and persists a new task.
func (a *App) AddTask(ctx context.Context, in AddTaskInput) (domain.Task, error) {
	if in.Title == "" {
		return domain.Task{}, fmt.Errorf("%w: title is required", domain.ErrInvalidInput)
	}

	id := domain.TaskID(ulid.Make().String())
	task := domain.NewTask(id, in.Title)
	task.Description = in.Description
	task.Tags = in.Tags
	task.ProjectID = in.ProjectID
	task.DueDate = in.DueDate
	task.Priority.CreationOrder = string(id)

	if task.DueDate != nil {
		days := int(time.Until(*task.DueDate).Hours() / 24)
		task.Priority.DeadlineUrgency = domain.DeadlineUrgencyScore(days)
	}

	if err := a.Tasks.Save(ctx, task); err != nil {
		return domain.Task{}, fmt.Errorf("add_task: save: %w", err)
	}

	_ = a.Events.Publish(ctx, domain.TaskCreatedEvent{TaskID: task.ID, CreatedAt: task.CreatedAt})
	_ = a.Search.Index(ctx, task)

	return task, nil
}

// ---- Next ----

// NextTask returns the highest-priority active task.
func (a *App) NextTask(ctx context.Context) (*domain.Task, error) {
	return a.Tasks.NextPriority(ctx)
}

// ---- List ----

// ListTasksInput specifies filters for listing tasks.
type ListTasksInput struct {
	Statuses  []domain.TaskStatus
	Tags      []string
	ProjectID *string
	Limit     int
	Offset    int
}

// ListTasksOutput is the paginated result.
type ListTasksOutput struct {
	Tasks []domain.Task
	Total int
}

// ListTasks returns a filtered, paginated list of tasks.
func (a *App) ListTasks(ctx context.Context, in ListTasksInput) (ListTasksOutput, error) {
	tasks, total, err := a.Tasks.List(ctx, ports.TaskFilter{
		Statuses:  in.Statuses,
		Tags:      in.Tags,
		ProjectID: in.ProjectID,
		Limit:     in.Limit,
		Offset:    in.Offset,
	})
	if err != nil {
		return ListTasksOutput{}, fmt.Errorf("list_tasks: %w", err)
	}
	services.SortByPriority(tasks)
	return ListTasksOutput{Tasks: tasks, Total: total}, nil
}

// ---- Complete ----

// CompleteTask marks a task as done.
func (a *App) CompleteTask(ctx context.Context, id domain.TaskID) (domain.Task, error) {
	task, err := a.Tasks.FindByID(ctx, id)
	if err != nil {
		return domain.Task{}, fmt.Errorf("complete_task: find: %w", err)
	}
	if !task.IsActive() {
		return domain.Task{}, fmt.Errorf("complete_task: %w: task is not active", domain.ErrConflict)
	}

	event := task.Complete()

	if err := a.Tasks.Save(ctx, task); err != nil {
		return domain.Task{}, fmt.Errorf("complete_task: save: %w", err)
	}

	_ = a.Events.Publish(ctx, event)
	_ = a.Search.Index(ctx, task)

	return task, nil
}

// ---- Cancel ----

// CancelTask marks a task as cancelled.
func (a *App) CancelTask(ctx context.Context, id domain.TaskID) (domain.Task, error) {
	task, err := a.Tasks.FindByID(ctx, id)
	if err != nil {
		return domain.Task{}, fmt.Errorf("cancel_task: find: %w", err)
	}

	event := task.Cancel()

	if err := a.Tasks.Save(ctx, task); err != nil {
		return domain.Task{}, fmt.Errorf("cancel_task: save: %w", err)
	}

	_ = a.Events.Publish(ctx, event)
	return task, nil
}

// ---- Promote ----

// PromoteTask increases the manual urgency of a task by one level.
func (a *App) PromoteTask(ctx context.Context, id domain.TaskID) (domain.Task, error) {
	task, err := a.Tasks.FindByID(ctx, id)
	if err != nil {
		return domain.Task{}, fmt.Errorf("promote_task: find: %w", err)
	}

	event := task.Promote()

	if err := a.Tasks.Save(ctx, task); err != nil {
		return domain.Task{}, fmt.Errorf("promote_task: save: %w", err)
	}

	_ = a.Events.Publish(ctx, event)
	return task, nil
}

// ---- Memo ----

// AddMemo sets the memo text of a task.
func (a *App) AddMemo(ctx context.Context, id domain.TaskID, text string) (domain.Task, error) {
	task, err := a.Tasks.FindByID(ctx, id)
	if err != nil {
		return domain.Task{}, fmt.Errorf("add_memo: find: %w", err)
	}

	task.AddMemo(text)

	if err := a.Tasks.Save(ctx, task); err != nil {
		return domain.Task{}, fmt.Errorf("add_memo: save: %w", err)
	}
	return task, nil
}

// ---- Search ----

// SearchTasks performs a full-text search over task titles and descriptions.
func (a *App) SearchTasks(ctx context.Context, query string, limit int) ([]ports.SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("%w: query is required", domain.ErrInvalidInput)
	}
	return a.Search.Search(ctx, query, limit)
}

// ---- Candidates ----

// Candidates returns the top N highest-priority active tasks.
func (a *App) Candidates(ctx context.Context, n int) ([]domain.Task, error) {
	tasks, err := a.Tasks.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("candidates: %w", err)
	}
	services.SortByPriority(tasks)
	if n > 0 && n < len(tasks) {
		tasks = tasks[:n]
	}
	return tasks, nil
}

// ---- Recalculate ----

// RecalculatePriorities runs the priority engine over all active tasks and
// saves the updated scores.
func (a *App) RecalculatePriorities(ctx context.Context) error {
	if a.Priority == nil {
		return nil
	}
	tasks, err := a.Tasks.FindActive(ctx)
	if err != nil {
		return fmt.Errorf("recalculate_priorities: find: %w", err)
	}
	updated, err := a.Priority.Recalculate(ctx, tasks)
	if err != nil {
		return fmt.Errorf("recalculate_priorities: engine: %w", err)
	}
	for _, t := range updated {
		if err := a.Tasks.Save(ctx, t); err != nil {
			return fmt.Errorf("recalculate_priorities: save %s: %w", t.ID, err)
		}
	}
	return nil
}
