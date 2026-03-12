package services_test

import (
	"testing"

	"github.com/quest-one/quest-one/internal/domain"
	"github.com/quest-one/quest-one/internal/domain/services"
)

func TestCompare_ManualUrgencyWinsFirst(t *testing.T) {
	high := domain.PriorityScore{ManualUrgency: domain.UrgencyHigh}
	low := domain.PriorityScore{ManualUrgency: domain.UrgencyLow}

	if got := services.Compare(high, low); got >= 0 {
		t.Errorf("expected high urgency to beat low urgency, got %d", got)
	}
	if got := services.Compare(low, high); got <= 0 {
		t.Errorf("expected low urgency to lose to high urgency, got %d", got)
	}
}

func TestCompare_EqualManualUrgency_UsesDeadline(t *testing.T) {
	a := domain.PriorityScore{ManualUrgency: domain.UrgencyNone, DeadlineUrgency: 90}
	b := domain.PriorityScore{ManualUrgency: domain.UrgencyNone, DeadlineUrgency: 10}

	if got := services.Compare(a, b); got >= 0 {
		t.Errorf("expected higher deadline urgency to win, got %d", got)
	}
}

func TestCompare_BlocksCountBreaksTie(t *testing.T) {
	a := domain.PriorityScore{BlocksCount: 5}
	b := domain.PriorityScore{BlocksCount: 1}

	if got := services.Compare(a, b); got >= 0 {
		t.Errorf("task blocking more others should have higher priority")
	}
}

func TestCompare_EqualScores_OlderWins(t *testing.T) {
	older := domain.PriorityScore{CreationOrder: "01ARZ3NDEKTSV4RRFFQ69G5FAV"} // lower ULID
	newer := domain.PriorityScore{CreationOrder: "01ARZ3NDEKTSV4RRFFQ69G5FAW"} // higher ULID

	if got := services.Compare(older, newer); got >= 0 {
		t.Errorf("older task (lower ULID) should win tie, got %d", got)
	}
}

func TestCompare_Symmetry(t *testing.T) {
	a := domain.PriorityScore{ManualUrgency: domain.UrgencyMedium, DeadlineUrgency: 50}
	b := domain.PriorityScore{ManualUrgency: domain.UrgencyLow, DeadlineUrgency: 80}

	ab := services.Compare(a, b)
	ba := services.Compare(b, a)

	if ab == 0 || ba == 0 {
		t.Error("expected non-zero comparison")
	}
	if (ab < 0) == (ba < 0) {
		t.Errorf("expected opposite signs: ab=%d ba=%d", ab, ba)
	}
}

func TestCompare_AllStepsEqual_ReturnsZero(t *testing.T) {
	s := domain.PriorityScore{
		ManualUrgency:   domain.UrgencyMedium,
		DeadlineUrgency: 50,
		BlocksCount:     2,
		Impact:          domain.ImpactMedium,
		SourcePriority:  3,
		WaitingDays:     7,
		DependencyDepth: 1,
		RecencyScore:    40,
		CreationOrder:   "01ARZ3NDEKTSV4RRFFQ69G5FAV",
	}

	if got := services.Compare(s, s); got != 0 {
		t.Errorf("identical scores should compare as 0, got %d", got)
	}
}

func TestSortByPriority_OrderIsCorrect(t *testing.T) {
	tasks := []domain.Task{
		{ID: "c", Priority: domain.PriorityScore{ManualUrgency: domain.UrgencyLow}},
		{ID: "a", Priority: domain.PriorityScore{ManualUrgency: domain.UrgencyCritical}},
		{ID: "b", Priority: domain.PriorityScore{ManualUrgency: domain.UrgencyHigh}},
	}

	services.SortByPriority(tasks)

	if tasks[0].ID != "a" || tasks[1].ID != "b" || tasks[2].ID != "c" {
		t.Errorf("unexpected order: %v %v %v", tasks[0].ID, tasks[1].ID, tasks[2].ID)
	}
}

func TestTop_ReturnsHighestPriority(t *testing.T) {
	tasks := []domain.Task{
		{ID: "x", Priority: domain.PriorityScore{ManualUrgency: domain.UrgencyLow}},
		{ID: "y", Priority: domain.PriorityScore{ManualUrgency: domain.UrgencyCritical}},
	}

	top := services.Top(tasks)
	if top == nil {
		t.Fatal("expected non-nil top task")
	}
	if top.ID != "y" {
		t.Errorf("expected task y, got %s", top.ID)
	}
}

func TestTop_EmptySlice_ReturnsNil(t *testing.T) {
	if top := services.Top(nil); top != nil {
		t.Error("expected nil for empty slice")
	}
}
