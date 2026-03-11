package diff

import (
	"testing"
	"time"
	"untis-notifier/untis"
)

func TestDiffLesson_StatusChange(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		old     untis.Lesson
		new     untis.Lesson
		wantLen int
	}{
		{
			name:    "cancelled",
			old:     untis.Lesson{Status: "REGULAR"},
			new:     untis.Lesson{Status: "CANCELLED"},
			wantLen: 1,
		},
		{
			name:    "rescheduled",
			old:     untis.Lesson{Start: now, End: now.Add(time.Hour)},
			new:     untis.Lesson{Start: now.Add(time.Hour), End: now.Add(2 * time.Hour)},
			wantLen: 2,
		},
		{
			name:    "changed",
			old:     untis.Lesson{Status: "regular", Teacher: untis.ChangableEntry{Current: untis.VariableString{Short: "ot", Long: "old teacher"}, Planned: untis.VariableString{Short: "ot", Long: "old teacher"}}, Notes: "old notes", Room: untis.ChangableEntry{Current: untis.VariableString{Short: "old", Long: "old room"}, Planned: untis.VariableString{Short: "old", Long: "old room"}}},
			new:     untis.Lesson{Status: "regular", Teacher: untis.ChangableEntry{Current: untis.VariableString{Short: "nt", Long: "new teacher"}, Planned: untis.VariableString{Short: "ot", Long: "old teacher"}}, Notes: "new notes", Room: untis.ChangableEntry{Current: untis.VariableString{Short: "new", Long: "new room"}, Planned: untis.VariableString{Short: "old", Long: "old room"}}},
			wantLen: 3,
		},
		{
			name:    "type change",
			old:     untis.Lesson{Type: "NORMAL_TEACHING_PERIOD"},
			new:     untis.Lesson{Type: "EXAM"},
			wantLen: 1,
		},
		{
			name:    "no change",
			old:     untis.Lesson{Start: now, End: now.Add(time.Hour)},
			new:     untis.Lesson{Start: now, End: now.Add(time.Hour)},
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diffLesson(tt.old, tt.new)
			if len(got) != tt.wantLen {
				t.Errorf("DiffLesson() = %v, want %v", got, tt.wantLen)
			}
		})
	}

}
