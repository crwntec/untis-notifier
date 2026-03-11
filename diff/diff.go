package diff

import (
	"slices"
	"time"
	"untis-notifier/untis"
)

type TimetableDiff struct {
	Changes []LessonDiff
}

type LessonDiff struct {
	Start   time.Time
	End     time.Time
	Subject string
	Changes []LessonChange
}

type LessonChange struct {
	Field  LessonField
	Before string
	After  string
}

type LessonField string

const (
	FieldStatus    LessonField = "status"
	FieldTeacher   LessonField = "teacher"
	FieldSubject   LessonField = "subject"
	FieldRoom      LessonField = "room"
	FieldNotes     LessonField = "notes"
	FieldStartTime LessonField = "startTime"
	FieldEndTime   LessonField = "endTime"
	FieldType      LessonField = "type"
)

func Compare(old, new untis.Timetable) TimetableDiff {
	var d TimetableDiff
	for _, day := range old.Days {
		for _, l := range day.Lessons {
			if other, ok := findLesson(new, l.IDs...); ok {
				changes := diffLesson(l, other)
				if len(changes) > 0 {
					d.Changes = append(d.Changes, LessonDiff{
						Subject: l.Subject.Planned.Long,
						Start:   l.Start,
						End:     l.End,
						Changes: changes,
					})
				}
			}
		}
	}
	return d
}

func findLesson(t untis.Timetable, ids ...int) (untis.Lesson, bool) {
	for _, day := range t.Days {
		for _, lesson := range day.Lessons {
			if slices.ContainsFunc(ids, func(id int) bool {
				return slices.Contains(lesson.IDs, id)
			}) {
				return lesson, true
			}
		}
	}
	return untis.Lesson{}, false
}

func diffLesson(l, other untis.Lesson) []LessonChange {
	var changes []LessonChange
	check := func(field LessonField, before, after string) {
		if before != after {
			changes = append(changes, LessonChange{Field: field, Before: before, After: after})
		}
	}
	check(FieldStatus, l.Status, other.Status)
	check(FieldTeacher, l.Teacher.Current.Long, other.Teacher.Current.Long)
	check(FieldSubject, l.Subject.Current.Long, other.Subject.Current.Long)
	check(FieldRoom, l.Room.Current.Short, other.Room.Current.Short)
	check(FieldNotes, l.Notes, other.Notes)
	check(FieldType, l.Type, other.Type)

	if !l.Start.Equal(other.Start) {
		check(FieldStartTime, l.Start.Format(untis.Layout), other.Start.Format(untis.Layout))
	}
	if !l.End.Equal(other.End) {
		check(FieldEndTime, l.End.Format(untis.Layout), other.End.Format(untis.Layout))
	}
	return changes
}
