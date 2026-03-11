# untis-notifier

![Build](https://github.com/crwntec/untis-notifier/actions/workflows/docker-image.yml/badge.svg)
![Docker Pulls](https://img.shields.io/docker/pulls/crwntec/untis-notifier)
![Go Version](https://img.shields.io/github/go-mod/go-version/crwntec/untis-notifier)
![License](https://img.shields.io/github/license/crwntec/untis-notifier)

> Stop manually checking WebUntis every morning. Get an instant push notification the moment your timetable changes — room swaps, cancellations, teacher changes, and more.

**untis-notifier** polls your [WebUntis](https://www.webuntis.com/) timetable on a schedule and sends a push notification via [ntfy](https://ntfy.sh) whenever anything changes. Works with any school that uses WebUntis (common across Germany, Austria, Switzerland, and other European countries).

---

## 📱 Example notification

```
📅 2 timetable change(s)

14.03.2025: M-LK (08:00 – 08:45)
  • status: REGULAR → CANCELLED

15.03.2025: PH-GK3 (10:00 – 10:45)
  • room: R101 → R204
  • teacher: VV → KT
```

Delivered to your phone via the [ntfy app](https://ntfy.sh) (Android / iOS / web).

---

## How it works

1. **Logs in** to WebUntis using your school credentials on startup
2. **Fetches** your timetable for today + the next `LOOK_AHEAD` days every `CHECK_INTERVAL`
3. **Diffs** the result against the previous fetch
4. **Sends a push notification** via ntfy if anything changed (room, teacher, status, notes, time)

---

## Quick start

### Prerequisites

- Docker & Docker Compose
- A WebUntis account (ask your school if you don't have one)
- Your school's WebUntis URL — find it by logging in via browser and copying the base URL

### 1. Find your school details

Log in to WebUntis in your browser. The URL will look like:

```
https://<server>.webuntis.com/WebUntis/?school=<schoolname>#/basic/main
```

- `UNTIS_BASE_URL` → `https://<server>.webuntis.com`
- `UNTIS_SCHOOL_NAME` → the value after `school=`

### 2. Create a `.env` file

```env
UNTIS_USER=your.username
UNTIS_PASS=yourpassword
UNTIS_BASE_URL=https://arche.webuntis.com
UNTIS_SCHOOL_NAME=your-school
NTFY_TOPIC=your-unique-secret-topic
```

> **Tip:** make `NTFY_TOPIC` something unguessable (e.g. `timetable-mustermann-xk92`) so others can't subscribe to your notifications.

### 3. Run

```bash
docker compose up -d
```

### 4. Subscribe to notifications

- **Browser:** open `https://ntfy.sh/<your-topic>`
- **App:** install [ntfy for Android](https://play.google.com/store/apps/details?id=io.heckel.ntfy) or [iOS](https://apps.apple.com/app/ntfy/id1625396347) and subscribe to your topic

---

## Configuration

All configuration is via environment variables:

| Variable            | Required | Default            | Description                                                  |
| ------------------- | -------- | ------------------ | ------------------------------------------------------------ |
| `UNTIS_USER`        | ✅       | —                  | Your WebUntis username                                       |
| `UNTIS_PASS`        | ✅       | —                  | Your WebUntis password                                       |
| `UNTIS_BASE_URL`    | ✅       | —                  | Base URL of your school's WebUntis instance                  |
| `UNTIS_SCHOOL_NAME` | ✅       | —                  | School identifier (from the login URL, see Quick Start)      |
| `NTFY_BASE_URL`     |          | `https://ntfy.sh`  | ntfy server URL (use your own for self-hosted)               |
| `NTFY_TOPIC`        |          | `untis-notifier`   | ntfy topic to publish to — **make this unique and secret**   |
| `LOOK_AHEAD`        |          | `7`                | Number of days ahead to monitor for changes                  |
| `CHECK_INTERVAL`    |          | `5m`               | How often to poll. Accepts Go durations: `30s`, `5m`, `1h`  |
| `LOG_FORMAT`        |          | `text`             | `text` for human-readable logs, `json` for log aggregators  |

---

## Self-hosted ntfy (recommended for privacy)

If you want notifications to stay fully private, run your own ntfy instance alongside the notifier:

```yaml
# docker-compose.yml
services:
  ntfy:
    image: binwiederhier/ntfy
    command: serve
    restart: unless-stopped
    ports:
      - "8080:80"
    volumes:
      - ntfy-data:/var/cache/ntfy

  untis-notifier:
    image: crwntec/untis-notifier:latest
    restart: unless-stopped
    env_file: .env
    environment:
      NTFY_BASE_URL: http://ntfy:80
    depends_on:
      - ntfy

volumes:
  ntfy-data:
```

Then point the ntfy app at your server's IP/domain instead of `ntfy.sh`.

---

## Troubleshooting

**Login fails immediately**
- Double-check `UNTIS_BASE_URL` — it should not have a trailing slash and should not include `/WebUntis`
- Verify `UNTIS_SCHOOL_NAME` matches exactly what appears in the browser URL after `school=`
- Try logging in manually in an incognito browser window to confirm credentials work

**No notifications arriving**
- Check that you subscribed to the exact topic name in `NTFY_TOPIC`
- Run `docker compose logs untis-notifier` and look for `notification sent` or any errors
- If using self-hosted ntfy, ensure the `NTFY_BASE_URL` is reachable from the container

**"first run — baseline stored" then nothing**
- This is expected. On the first run, the current timetable is saved as a baseline. Notifications are only sent when something *changes* from that baseline. Wait for the next poll after a real change occurs.

**Seeing too many false positives**
- Lower `LOOK_AHEAD` to reduce the window being monitored (e.g. `LOOK_AHEAD=2` for today and tomorrow only)

---

## Local development

```bash
# Copy and fill in credentials
cp .env.example .env

# Build and run
go build -o untis-notifier . && ./untis-notifier

# Run tests
go test ./...

# Live reload (requires air)
go install github.com/air-verse/air@latest
air
```

---

## Supported schools

WebUntis is widely used across **Germany, Austria, Switzerland, South Tyrol, and the Netherlands**. If your school uses WebUntis for its timetable, this tool should work without any changes.

Not sure if your school uses WebUntis? Check if your school's timetable URL contains `webuntis.com`.

---

## Contributing

Contributions are welcome! Please open an issue first for significant changes so we can discuss the approach.

- Bug reports: use the issue tracker
- Feature requests: open a discussion or issue
- Pull requests: please include tests for new logic

---

## License

MIT
