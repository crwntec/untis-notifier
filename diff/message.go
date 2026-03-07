// diff/message.go
package diff

import (
	"fmt"
	"strings"
	"untis-notifier/notifier"
)

func ToMessage(d TimetableDiff) notifier.Message {
	if len(d.Changes) == 0 {
		return notifier.Message{}
	}

	var sb strings.Builder
	for _, lesson := range d.Changes {
		fmt.Fprintf(&sb, "%s: %s (%s – %s)\n",
			lesson.Start.Format("02.01.2006"),
			lesson.Subject,
			lesson.Start.Format("15:04"),
			lesson.End.Format("15:04"),
		)
		for _, change := range lesson.Changes {
			fmt.Fprintf(&sb, "  • %s: %s → %s\n", change.Field, change.Before, change.After)
		}
		sb.WriteString("\n")
	}

	return notifier.Message{
		Title:    fmt.Sprintf("%d timetable change(s)", len(d.Changes)),
		Priority: 3,
		Tags:     []string{"calendar", "warning"},
		Text:     strings.TrimSpace(sb.String()),
	}
}
