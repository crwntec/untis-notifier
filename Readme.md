# untis-notifier

Polls your WebUntis timetable on a schedule and sends a push notification via [ntfy](https://ntfy.sh) whenever something changes (room swap, cancellation, teacher change, etc).

## How it works

1. Logs in to WebUntis on startup
2. Fetches today's and tomorrow's timetable every `CHECK_INTERVAL`
3. Diffs the result against the previous fetch
4. Sends an ntfy notification if anything changed

---

## Quick start with Docker Compose

**1. Create a `.env` file:**

```env
UNTIS_USER=your.username
UNTIS_PASS=yourpassword
UNTIS_BASE_URL=https://your-school.webuntis.com
UNTIS_SCHOOL_NAME=your-school
NTFY_TOPIC=your-unique-topic-name
```

**2. Run it:**

```bash
docker compose up -d
```

**3. Subscribe to notifications:**

Open `https://ntfy.sh/<your-topic>` in a browser, or install the ntfy app and subscribe to your topic.

---

## Configuration

All configuration is via environment variables:

| Variable            | Required | Default           | Description                                                 |
| ------------------- | -------- | ----------------- | ----------------------------------------------------------- |
| `UNTIS_USER`        | ✅       | —                 | Your WebUntis username                                      |
| `UNTIS_PASS`        | ✅       | —                 | Your WebUntis password                                      |
| `UNTIS_BASE_URL`    | ✅       | -                 | Base URL of your school's WebUntis instance                 |
| `UNTIS_SCHOOL_NAME` | ✅       | -                 | School identifier (shown in the login URL)                  |
| `NTFY_BASE_URL`     |          | `https://ntfy.sh` | ntfy server URL (use your own for self-hosted)              |
| `NTFY_TOPIC`        |          | `untis-notifier`  | ntfy topic to publish to — make this unique                 |
| `LOOK_AHEAD`        |          | `7`               | Number of days to look ahead for changes                    |
| `CHECK_INTERVAL`    |          | `5m`              | How often to check. Accepts Go durations: `1m`, `30s`, `1h` |
| `LOG_FORMAT`        |          | `text`            | `text` for human-readable, `json` for log aggregators       |

---

## Self-hosted ntfy

If you want notifications to stay private, add a self-hosted ntfy instance to your compose file:

```yaml
services:
  ntfy:
    image: binwiederhier/ntfy
    command: serve
    ports:
      - "8080:80"
    volumes:
      - ntfy-data:/var/cache/ntfy

  untis-notifier:
    image: crwntec/untis-notifier:latest
    environment:
      NTFY_BASE_URL: http://ntfy:80
      # ... other vars

volumes:
  ntfy-data:
```

---

## Local development

```bash
# Copy and fill in credentials
cp .env.example .env

# Run (builds first, so Ctrl+C works correctly)
go build -o checker . && ./checker

# Or with live reload
go install github.com/air-verse/air@latest
air
```
