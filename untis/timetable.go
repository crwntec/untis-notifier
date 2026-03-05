package untis

import (
	"fmt"
	"time"
)

func positionItemToChangableEntry(items []PositionItem) (ChangableEntry, error) {
	if len(items) == 0 {
		return ChangableEntry{}, fmt.Errorf("no position items provided")
	}
	item := items[0]

	var current VariableString
	if item.Current != nil {
		current = VariableString{item.Current.ShortName, item.Current.LongName}
	}

	planned := current
	if item.Removed != nil {
		planned = VariableString{item.Removed.ShortName, item.Removed.LongName}
	}

	return ChangableEntry{
		Current: current,
		Planned: planned,
	}, nil
}

func TimetableFromResponse(resp TimetableResponse) (Timetable, error) {
	timetable := Timetable{}
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return Timetable{}, fmt.Errorf("setting up date location %w", err)
	}
	for _, d := range resp.Days {
		date, err := time.ParseInLocation(time.DateOnly, d.Date, loc)
		if err != nil {
			return Timetable{}, fmt.Errorf("parsing date %q: %w", d.Date, err)
		}
		day := TimetableDay{
			Date:   date,
			Status: d.Status,
		}
		for _, ge := range d.GridEntries {
			start, err := time.ParseInLocation(Layout, ge.Duration.Start, loc)
			if err != nil {
				return Timetable{}, fmt.Errorf("parsing date %q: %w", ge.Duration.Start, err)
			}
			end, err := time.ParseInLocation(Layout, ge.Duration.End, loc)
			if err != nil {
				return Timetable{}, fmt.Errorf("parsing date %q: %w", ge.Duration.End, err)
			}
			teacher, err := positionItemToChangableEntry(ge.Position1)
			if err != nil {
				return Timetable{}, fmt.Errorf("parsing teacher position: %w", err)
			}
			subject, err := positionItemToChangableEntry(ge.Position2)
			if err != nil {
				return Timetable{}, fmt.Errorf("parsing subject position: %w", err)
			}
			room, err := positionItemToChangableEntry(ge.Position3)
			if err != nil {
				return Timetable{}, fmt.Errorf("parsing room position: %w", err)
			}

			lesson := Lesson{
				IDs:     ge.IDs,
				Start:   start,
				End:     end,
				Type:    ge.Type,
				Notes:   ge.NotesAll,
				Status:  ge.Status,
				Teacher: teacher,
				Subject: subject,
				Room:    room,
			}
			day.Lessons = append(day.Lessons, lesson)
		}
		timetable.Days = append(timetable.Days, day)
	}
	return timetable, nil
}
