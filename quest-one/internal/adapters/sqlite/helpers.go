package sqlite

import "time"

// boolToInt converts a bool to SQLite integer (0/1).
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// nullableTimeToStr converts an optional *time.Time to a *string for SQLite storage (RFC3339).
func nullableTimeToStr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

// parseNullableTime parses an optional *string from SQLite into a *time.Time.
func parseNullableTime(s *string) *time.Time {
	if s == nil {
		return nil
	}
	t, _ := time.Parse(time.RFC3339, *s)
	return &t
}

// parseTime parses a non-null RFC3339 string from SQLite.
func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
