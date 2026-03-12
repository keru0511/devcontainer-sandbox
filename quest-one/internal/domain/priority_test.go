package domain_test

import (
	"testing"

	"github.com/quest-one/quest-one/internal/domain"
)

func TestDeadlineUrgencyScore_Overdue(t *testing.T) {
	if got := domain.DeadlineUrgencyScore(-1); got != 100 {
		t.Errorf("overdue should be 100, got %d", got)
	}
	if got := domain.DeadlineUrgencyScore(0); got != 100 {
		t.Errorf("due today should be 100, got %d", got)
	}
}

func TestDeadlineUrgencyScore_Far(t *testing.T) {
	if got := domain.DeadlineUrgencyScore(30); got != 0 {
		t.Errorf("30 days away should be 0, got %d", got)
	}
	if got := domain.DeadlineUrgencyScore(100); got != 0 {
		t.Errorf("100 days away should be 0, got %d", got)
	}
}

func TestDeadlineUrgencyScore_Interpolation(t *testing.T) {
	// 15 days out → roughly 50
	got := domain.DeadlineUrgencyScore(15)
	if got < 45 || got > 55 {
		t.Errorf("15-day score expected ~50, got %d", got)
	}
}

func TestTaskComplete_SetsStatus(t *testing.T) {
	task := domain.NewTask("test01", "Test task")
	event := task.Complete()

	if task.Status != domain.TaskStatusDone {
		t.Errorf("expected Done, got %s", task.Status)
	}
	if task.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
	if _, ok := event.(domain.TaskCompletedEvent); !ok {
		t.Error("expected TaskCompletedEvent")
	}
}

func TestTaskPromote_IncrementsUrgency(t *testing.T) {
	task := domain.NewTask("test02", "Test task")
	task.Priority.ManualUrgency = domain.UrgencyNone

	task.Promote()
	if task.Priority.ManualUrgency != domain.UrgencyLow {
		t.Errorf("expected urgency 1 after promote, got %d", task.Priority.ManualUrgency)
	}
}

func TestTaskPromote_MaxCaps(t *testing.T) {
	task := domain.NewTask("test03", "Test task")
	task.Priority.ManualUrgency = domain.UrgencyMax

	task.Promote()
	if task.Priority.ManualUrgency > domain.UrgencyMax {
		t.Errorf("urgency exceeded max after promote: %d", task.Priority.ManualUrgency)
	}
}

func TestTaskIsActive(t *testing.T) {
	active := []domain.TaskStatus{domain.TaskStatusTodo, domain.TaskStatusInProgress, domain.TaskStatusWaiting}
	for _, s := range active {
		task := domain.NewTask("x", "x")
		task.Status = s
		if !task.IsActive() {
			t.Errorf("status %s should be active", s)
		}
	}

	inactive := []domain.TaskStatus{domain.TaskStatusDone, domain.TaskStatusCancelled}
	for _, s := range inactive {
		task := domain.NewTask("x", "x")
		task.Status = s
		if task.IsActive() {
			t.Errorf("status %s should not be active", s)
		}
	}
}
