package untis

import "testing"

func TestTimeTableFromResponse_empty(t *testing.T) {
	got, err := TimetableFromResponse(TimetableResponse{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Days) != 0 {
		t.Errorf("got %d days, want 0", len(got.Days))
	}
}

func TestTimeTableFromResponse_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input TimetableResponse
	}{
		{
			name: "bad date",
			input: TimetableResponse{
				Days: []TimetableResponseDay{{Date: "not-a-date"}},
			},
		},
		{name: "bad start time",
			input: TimetableResponse{
				Days: []TimetableResponseDay{
					{
						Date: "2024-01-15",
						GridEntries: []GridEntry{
							{
								Duration: struct {
									Start string `json:"start"`
									End   string `json:"end"`
								}{Start: "bad", End: "2024-01-15T10:00"},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := TimetableFromResponse(tt.input)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}
