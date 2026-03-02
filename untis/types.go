// untis/types.go
package untis

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

type TimetableResponse struct {
	Days   []TimetableDay `json:"days"`
	Errors []any          `json:"errors"`
}

type TimetableDay struct {
	Date        string      `json:"date"`
	Status      string      `json:"status"` // REGULAR | NO_DATA
	GridEntries []GridEntry `json:"gridEntries"`
}

type GridEntry struct {
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
	Current PositionCurrent `json:"current"`
}

type PositionCurrent struct {
	Type      string `json:"type"`
	Status    string `json:"status"`
	ShortName string `json:"shortName"`
	LongName  string `json:"longName"`
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
type Session struct {
	SessionID string
	Token     string
	SchoolID  string
}
type Config struct {
	BaseURL    string
	SchoolName string
}
