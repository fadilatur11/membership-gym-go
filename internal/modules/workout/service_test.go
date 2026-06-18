package workout

import "testing"

func TestAdminWorkoutSource(t *testing.T) {
	tests := map[string]string{
		"owner":   "admin",
		"admin":   "admin",
		"trainer": "trainer",
	}

	for role, expected := range tests {
		if actual := adminWorkoutSource(role); actual != expected {
			t.Errorf("adminWorkoutSource(%q) = %q, want %q", role, actual, expected)
		}
	}
}
