// main.go
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"untis-notifier/diff"
	"untis-notifier/notifier"
	"untis-notifier/untis"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := untis.Config{
		BaseURL:    "https://st-bernhard-gym.webuntis.com",
		SchoolName: "st-bernhard-gym",
	}

	client, err := untis.NewClient(cfg)
	if err != nil {
		slog.Error("failed to create client", "err", err)
		os.Exit(1)
	}

	username := os.Getenv("UNTIS_USER")
	password := os.Getenv("UNTIS_PASS")

	slog.Info("logging in", "user", username)
	if err = client.Login(username, password); err != nil {
		slog.Error("login failed", "err", err)
		os.Exit(1)
	}
	slog.Info("login successful")

	info, err := client.GetStaticInfo()
	if err != nil {
		slog.Error("failed to fetch static info", "err", err)
		os.Exit(1)
	}
	slog.Info("fetched user info", "userID", info.UserID, "schoolID", info.SchoolID)

	ntfy := notifier.NewClient(notifier.Config{BaseURL: "https://ntfy.sh"})
	const interval = 1 * time.Minute
	slog.Info("starting timetable checker", "interval", interval)
	var prev untis.Timetable
	first := true
	for {
		prev, first = runCheck(client, ntfy, ctx, info, prev, first)
		select {
		case <-ctx.Done():
			slog.Info("shutting down")
			return
		case <-time.After(interval):
		}
	}
}

func runCheck(client *untis.Client, ntfy *notifier.Client, ctx context.Context, info untis.UntisInfo, old untis.Timetable, first bool) (untis.Timetable, bool) {
	now := time.Now()
	start := now.Format(time.DateOnly)
	end := now.Add(24 * time.Hour).Format(time.DateOnly)

	slog.Info("fetching timetable", "start", start, "end", end)

	timetable, err := client.GetTimetable(info, start, end)
	if err != nil {
		slog.Error("failed to fetch timetable", "err", err)
		return old, first
	}

	if !first {
		d := diff.Compare(old, timetable)
		slog.Info("timetable diff", "changes", d)
		if len(d.Changes) > 0 {
			msg := diff.ToMessage(d)
			if err := ntfy.SendMessage(ctx, "untis-alert", msg); err != nil {
				slog.Error("failed to send notification", "err", err)
			}
		} else {
			slog.Info("no timetable changes")
		}
	}
	return timetable, false

}
