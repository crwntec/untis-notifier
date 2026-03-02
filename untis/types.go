// untis/types.go
package untis

// --- Raw API types (what WebUntis actually returns) ---

type RawPeriodElement struct {
	Type    int    `json:"type"`
	ID      int    `json:"id"`
	OrgID   int    `json:"orgId"`
	Missing bool   `json:"missing"`
	State   string `json:"state"`
}

type RawUntisPeriod struct {
	ID                int                `json:"id"`
	LessonID          int                `json:"lessonId"`
	LessonNumber      int                `json:"lessonNumber"`
	LessonCode        string             `json:"lessonCode"`
	LessonText        string             `json:"lessonText"`
	PeriodText        string             `json:"periodText"`
	HasPeriodText     bool               `json:"hasPeriodText"`
	PeriodInfo        string             `json:"periodInfo"`
	PeriodAttachments []PeriodAttachment `json:"periodAttachments"`
	SubstText         string             `json:"substText"`
	Date              int                `json:"date"`
	StartTime         int                `json:"startTime"`
	EndTime           int                `json:"endTime"`
	Elements          []RawPeriodElement `json:"elements"`
	StudentGroup      string             `json:"studentGroup"`
	HasInfo           bool               `json:"hasInfo"`
	Code              int                `json:"code"`
	CellState         string             `json:"cellState"`
	Priority          int                `json:"priority"`
	Is                PeriodIs           `json:"is"`
	RoomCapacity      int                `json:"roomCapacity"`
	StudentCount      int                `json:"studentCount"`
}

type PeriodAttachment struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

type PeriodIs struct {
	Standard bool `json:"standard"`
	Event    bool `json:"event"`
}

type RawUntisElement struct {
	Type             int    `json:"type"`
	ID               int    `json:"id"`
	Name             string `json:"name"`
	LongName         string `json:"longName"`
	DisplayName      string `json:"displayname"`
	AlternateName    string `json:"alternatename"`
	CanViewTimetable bool   `json:"canViewTimetable"`
	RoomCapacity     int    `json:"roomCapacity"`
	ExternKey        string `json:"externKey"`
}

type RawUntisData struct {
	NoDetails      bool                        `json:"noDetails"`
	ElementIDs     []int                       `json:"elementIds"`
	ElementPeriods map[string][]RawUntisPeriod `json:"elementPeriods"`
	Elements       []RawUntisElement           `json:"elements"`
}

type LoginResponse struct {
	Result *struct {
		SessionID string `json:"sessionId"`
		Token     string `json:"token"`
	} `json:"result"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// --- Processed types (what your code works with) ---

type UntisElement struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	LongName string `json:"longName,omitempty"`
}

type Lesson struct {
	ID             int          `json:"id"`
	StartTime      string       `json:"startTime"`
	EndTime        string       `json:"endTime"`
	Date           string       `json:"date"`
	Teacher        string       `json:"teacher"`
	SubstTeacher   string       `json:"substTeacher"`
	Subject        string       `json:"subject"`
	Room           UntisElement `json:"room"`
	AdditionalInfo string       `json:"additionalInfo"`
	IsSubstitution bool         `json:"isSubstitution"`
	IsTeams        bool         `json:"isTeams"`
	IsEva          bool         `json:"isEva"`
	IsCancelled    bool         `json:"isCancelled"`
	IsFree         bool         `json:"isFree"`
}

type Block struct {
	Lesson
	IsInHoliday bool `json:"isInHoliday"`
}

// --- Session / auth types ---

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
	UserID    int
}
type UntisInfo struct {
	SessionID         string     `json:"sessionID"`
	UserID            int        `json:"userID"`
	SchoolID          string     `json:"schoolID"`
	AllowedClass      int        `json:"allowedClass"`
	CurrentSchoolYear SchoolYear `json:"currentSchoolYear"`
	Holidays          []any      `json:"holidays"`
}

type Config struct {
	BaseURL    string
	SchoolName string
}

// --- Response wrapper (getData return value) ---

type TimetableData struct {
	Lessons  []Block        `json:"lessons"`
	Teachers []UntisElement `json:"teachers"`
	Rooms    []UntisElement `json:"rooms"`
	Subjects []UntisElement `json:"subjects"`
}
