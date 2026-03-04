package main

import (
	"context"
	"fmt"
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
	// .env is optional — in Docker, vars come from the environment directly
	_ = godotenv.Load()

	setupLogger()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := configFromEnv()
	if err != nil {
		slog.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	client, err := untis.NewClient(cfg.untis)
	if err != nil {
		slog.Error("failed to create untis client", "err", err)
		os.Exit(1)
	}

	slog.Info("logging in to WebUntis",
		"url", cfg.untis.BaseURL,
		"school", cfg.untis.SchoolName,
		"user", cfg.username,
	)
	if err = client.Login(cfg.username, cfg.password); err != nil {
		slog.Error("login failed", "err", err)
		os.Exit(1)
	}
	slog.Info("login successful")

	info, err := client.GetStaticInfo()
	if err != nil {
		slog.Error("failed to fetch user info", "err", err)
		os.Exit(1)
	}
	slog.Info("fetched user info", "userID", info.UserID, "schoolID", info.SchoolID)

	ntfy := notifier.NewClient(notifier.Config{BaseURL: cfg.ntfyBaseURL})

	slog.Info("starting timetable checker",
		"interval", cfg.interval,
		"ntfy_topic", cfg.ntfyTopic,
	)

	var prev untis.Timetable
	first := true
	for {
		prev, first = runCheck(ctx, client, ntfy, cfg.ntfyTopic, info, prev, first)

		select {
		case <-ctx.Done():
			slog.Info("shutting down gracefully")
			return
		case <-time.After(cfg.interval):
		}
	}
}

func runCheck(
	ctx context.Context,
	client *untis.Client,
	ntfy *notifier.Client,
	topic string,
	info untis.UntisInfo,
	old untis.Timetable,
	first bool,
) (untis.Timetable, bool) {
	now := time.Now()
	start := now.Format(time.DateOnly)
	end := now.Add(24 * time.Hour).Format(time.DateOnly)

	slog.Info("fetching timetable", "start", start, "end", end)

	timetable, err := client.GetTimetable(ctx, info, start, end)
	if err != nil {
		if ctx.Err() != nil {
			return old, first
		}
		slog.Error("failed to fetch timetable", "err", err)
		return old, first
	}

	slog.Info("timetable fetched",
		"days", len(timetable.Days),
	)

	if first {
		slog.Info("first run — baseline stored, changes will be reported from next check")
		return timetable, false
	}

	d := diff.Compare(old, timetable)
	if len(d.Changes) == 0 {
		slog.Info("no timetable changes detected")
		return timetable, false
	}

	slog.Info("timetable changes detected", "count", len(d.Changes))
	for _, change := range d.Changes {
		slog.Info("changed lesson",
			"subject", change.Subject,
			"start", change.Start.Format("15:04"),
			"end", change.End.Format("15:04"),
			"fields_changed", len(change.Changes),
		)
	}

	msg := diff.ToMessage(d)
	if err := ntfy.SendMessage(ctx, topic, msg); err != nil {
		if ctx.Err() != nil {
			return timetable, false
		}
		slog.Error("failed to send notification", "err", err)
	} else {
		slog.Info("notification sent", "topic", topic, "title", msg.Title)
	}

	return timetable, false
}

// appConfig holds all runtime configuration
type appConfig struct {
	untis       untis.Config
	username    string
	password    string
	ntfyBaseURL string
	ntfyTopic   string
	interval    time.Duration
}

func configFromEnv() (appConfig, error) {
	username := os.Getenv("UNTIS_USER")
	password := os.Getenv("UNTIS_PASS")
	if username == "" || password == "" {
		return appConfig{}, fmt.Errorf("UNTIS_USER and UNTIS_PASS must be set")
	}

	baseURL := envOr("UNTIS_BASE_URL", "https://st-bernhard-gym.webuntis.com")
	schoolName := envOr("UNTIS_SCHOOL_NAME", "st-bernhard-gym")
	ntfyBaseURL := envOr("NTFY_BASE_URL", "https://ntfy.sh")
	ntfyTopic := envOr("NTFY_TOPIC", "untis-alerts")

	intervalStr := envOr("CHECK_INTERVAL", "5m")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return appConfig{}, fmt.Errorf("invalid CHECK_INTERVAL %q: %w", intervalStr, err)
	}

	return appConfig{
		untis: untis.Config{
			BaseURL:    baseURL,
			SchoolName: schoolName,
		},
		username:    username,
		password:    password,
		ntfyBaseURL: ntfyBaseURL,
		ntfyTopic:   ntfyTopic,
		interval:    interval,
	}, nil
}

func setupLogger() {
	format := os.Getenv("LOG_FORMAT")
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}

	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
