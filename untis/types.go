// untis/types.go
package untis

import "time"

type AppData struct {
	CurrentSchoolYear SchoolYear  `json:"currentSchoolYear"`
	Holidays          []Holiday   `json:"holidays"`
	Tenant            Tenant      `json:"tenant"`
	User              AppDataUser `json:"user"`
	Permissions       []string    `json:"permissions"`
	Settings          []string    `json:"settings"`
}

type Tenant struct {
	DisplayName string `json:"displayName"`
	ID          string `json:"id"`
	Name        string `json:"name"`
}

type Holiday struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Bookable bool   `json:"bookable"`
}

type AppDataUser struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Email  string   `json:"email"`
	Locale string   `json:"locale"`
	Person Person   `json:"person"`
	Roles  []string `json:"roles"`
}

type Person struct {
	DisplayName string `json:"displayName"`
	ID          int    `json:"id"`
}

type UntisInfo struct {
	UserID            int        `json:"userID"`
	SchoolID          string     `json:"schoolID"`
	AllowedClass      int        `json:"allowedClass"`
	CurrentSchoolYear SchoolYear `json:"currentSchoolYear"`
	Holidays          []Holiday  `json:"holidays"`
}

type SchoolYear struct {
	DateRange DateRange `json:"dateRange"`
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	TimeGrid  TimeGrid  `json:"timeGrid"`
}

type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type TimeGrid struct {
	SchoolYearID int   `json:"schoolyearId"`
	Units        []any `json:"units"`
}

type TimetableResponse struct {
	Days   []TimetableResponseDay `json:"days"`
	Errors []any                  `json:"errors"`
}

type TimetableResponseDay struct {
	Date        string      `json:"date"`
	Status      string      `json:"status"` // REGULAR | NO_DATA
	GridEntries []GridEntry `json:"gridEntries"`
}

type GridEntry struct {
	IDs      []int `json:"ids"`
	Duration struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"duration"`
	Type      string         `json:"type"`
	Status    string         `json:"status"` // REGULAR | CHANGED | CANCELLED
	NotesAll  string         `json:"notesAll"`
	Position1 []PositionItem `json:"position1"` // teacher
	Position2 []PositionItem `json:"position2"` // subject
	Position3 []PositionItem `json:"position3"` // room
}

type PositionItem struct {
	Current *PositionItemEntry `json:"current"`
	Removed *PositionItemEntry `json:"removed"`
}

type PositionItemEntry struct {
	Type      string `json:"type"`
	Status    string `json:"status"`
	ShortName string `json:"shortName"`
	LongName  string `json:"longName"`
}

type Timetable struct {
	Days []TimetableDay
}
type TimetableDay struct {
	Date    time.Time
	Status  string // REGULAR | NO_DATA
	Lessons []Lesson
}
type Lesson struct {
	IDs     []int `json:"ids"`
	Start   time.Time
	End     time.Time
	Type    string
	Status  string
	Notes   string
	Teacher ChangableEntry
	Subject ChangableEntry
	Room    ChangableEntry
}
type ChangableEntry struct {
	Current VariableString
	Planned VariableString
}

type VariableString struct {
	Short string
	Long  string
}

type Session struct {
	SessionID string
	Token     string
	SchoolID  string
}
type Config struct {
	BaseURL    string
	SchoolName string
}

const Layout = "2006-01-02T15:04"
