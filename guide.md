# Learning Go: WebUntis Timetable Notifier

A project-based guide that teaches Go by porting your existing TypeScript/n8n WebUntis notifier.
You'll end up with a single ~6MB Docker image that polls WebUntis and pushes to ntfy.sh.

---

## 1. Why This Project Works Well for Learning Go

Your existing TypeScript code already covers the main areas Go excels at:

- **Structs & interfaces** — you have rich TypeScript interfaces (`Lesson`, `RawUntisPeriod`, etc.)
- **HTTP client** — login, session cookies, parallel week fetching
- **JSON parsing** — deeply nested API responses
- **Concurrency** — `axios.all()` parallel requests become goroutines
- **State persistence** — the n8n static data becomes a JSON file on disk

---

## 2. Environment Setup

### Install Go
Download from [go.dev/dl](https://go.dev/dl) — pick the latest stable (1.23+).

```bash
go version  # should print go version go1.23.x
```

### Create Your Project

```bash
mkdir untis-notifier
cd untis-notifier
go mod init untis-notifier   # creates go.mod — like package.json
```

No external dependencies are needed. Everything you need is in the standard library:
- `net/http` — HTTP client (replaces axios)
- `encoding/json` — JSON marshal/unmarshal
- `net/http/cookiejar` — cookie handling (replaces qs + manual cookie headers)
- `time` — date/time (replaces moment)
- `os` — file I/O for state persistence
- `sync` — WaitGroup for parallel requests (replaces axios.all)
- `log` — logging

### Suggested File Structure

```
untis-notifier/
├── go.mod
├── main.go               # entry point + scheduler loop
├── untis/
│   ├── client.go         # login, session management
│   ├── timetable.go      # fetching + parsing lessons
│   └── types.go          # all structs (your interfaces)
├── notifier/
│   └── ntfy.go           # ntfy.sh notification
├── state/
│   └── state.go          # JSON file persistence (replaces n8n static data)
└── Dockerfile
```

---

## 3. Core Go Concepts You'll Need

### 3.1 Structs (your interfaces become these)

```go
// TypeScript:
// interface Lesson { id: number; startTime: string; ... }

type Lesson struct {
    ID          int    `json:"id"`
    StartTime   string `json:"startTime"`
    EndTime     string `json:"endTime"`
    Date        string `json:"date"`
    Subject     string `json:"subject"`
    Teacher     string `json:"teacher"`
    IsSubst     bool   `json:"isSubstitution"`
    IsCancelled bool   `json:"isCancelled"`
    // etc.
}
```

The backtick tags (`json:"..."`) control how the struct maps to/from JSON.
They're how you handle the snake_case vs camelCase differences in the API response.

### 3.2 Errors (no try/catch — errors are return values)

```go
// Functions that can fail return (value, error)
func login(username, password string) (Session, error) {
    // ...
    if resp.StatusCode != 200 {
        return Session{}, fmt.Errorf("login failed: status %d", resp.StatusCode)
    }
    return session, nil
}

// Callers check the error immediately
session, err := login(user, pass)
if err != nil {
    log.Fatal(err)  // or handle gracefully
}
```

This pattern will appear hundreds of times in your code. Get comfortable with it early.

### 3.3 HTTP Client

```go
client := &http.Client{}

// POST with form data (your j_spring_security_check login)
resp, err := client.PostForm(baseURL+"/WebUntis/j_spring_security_check", url.Values{
    "school":      {schoolName},
    "j_username":  {username},
    "j_password":  {password},
    "token":       {""},
})
defer resp.Body.Close()  // always close the body

// Read response body
body, err := io.ReadAll(resp.Body)

// Parse JSON
var result LoginResponse
json.Unmarshal(body, &result)
```

For cookie handling across requests (your JSESSIONID + schoolname cookies),
use `http.CookieJar` — attach it to your client and it handles cookies automatically,
just like a browser would.

### 3.4 Goroutines (your axios.all parallel fetches)

```go
// TypeScript:
// const promises = dates.map(d => makeRequest(d, ...))
// const responses = await axios.all(promises)

results := make([]WeekData, len(dates))
var wg sync.WaitGroup
var mu sync.Mutex  // protect shared slice from race conditions

for i, date := range dates {
    wg.Add(1)
    go func(i int, date string) {
        defer wg.Done()
        data, err := fetchWeek(date)
        if err != nil { return }
        mu.Lock()
        results[i] = data
        mu.Unlock()
    }(i, date)  // pass i and date as arguments — important to avoid closure capture bugs
}
wg.Wait()  // blocks until all goroutines finish
```

### 3.5 Maps (your lesson change detection)

```go
// Build a lookup map — same as your previousLessonsMap
previousMap := make(map[string]Lesson)
for _, lesson := range previousLessons {
    key := fmt.Sprintf("%d%s", lesson.ID, lesson.Date)
    previousMap[key] = lesson
}

// Check existence
if prev, exists := previousMap[key]; exists {
    // compare prev vs current
} else {
    // new lesson
}
```

### 3.6 Time

```go
// Parse a date string
t, err := time.Parse("02.01.2006", lesson.Date)  // Go uses this specific reference time

// Format
formatted := t.Format("2006-01-02")  // YYYY-MM-DD

// Current time
now := time.Now()

// Check if lesson is in the past
lessonEnd, _ := time.Parse("02.01.2006 15:04", lesson.Date+" "+lesson.EndTime)
isPast := lessonEnd.Before(time.Now())
```

Go's time format is unusual — it always uses the reference date `Mon Jan 2 15:04:05 MST 2006`.
Think of it as "1-2-3-4-5-6-7" for month-day-hour-min-sec-year-timezone. You'll get used to it.

---

## 4. Porting Roadmap

Work through these in order. Each phase produces something runnable.

### Phase 1 — Structs (`untis/types.go`)

Start here. No HTTP, no logic — just translate your TypeScript interfaces into Go structs.

Key interfaces to port:
- `RawUntisPeriod` → be careful with nested types like `elements []struct{ ... }`
- `RawUntisElement`
- `Lesson`
- `Block` (embeds Lesson — use struct embedding in Go)
- Session info (sessionID, userID, schoolID, etc.)

Hint for struct embedding (your `Block extends Lesson`):
```go
type Block struct {
    Lesson              // embedded — Block gets all Lesson fields
    IsInHoliday bool    `json:"isInHoliday"`
}
```

Run `go build ./...` after each file — the compiler catches type errors immediately.

### Phase 2 — Login (`untis/client.go`)

Port `getInfo()`. This is the most complex single function because it:
1. POSTs form data to get a session
2. Extracts cookies from the response headers
3. Makes two more GET requests with that session
4. Returns assembled session info

Focus areas:
- Parsing `set-cookie` headers manually (or use a CookieJar)
- `url.Values` for form encoding (replaces `qs.stringify`)
- Struct tags for JSON parsing the login response

Test by printing the session ID and user ID to stdout before moving on.

### Phase 3 — Single Week Fetch (`untis/timetable.go`)

Port `makeRequest()` and `getLessonsForWeek()`.

The parsing logic in `getLessonsForWeek` is the trickiest part — the time formatting
(`lesson.startTime.toString().length == 3 ? ...`) should be simplified using `fmt.Sprintf`:
```go
// Instead of the string length gymnastics in your TS:
startTime := fmt.Sprintf("%04d", rawLesson.StartTime)  // zero-pad to 4 digits
formatted := startTime[:2] + ":" + startTime[2:]       // "0800" → "08:00"
```

### Phase 4 — Parallel Fetches

Port `getData()` — specifically the `axios.all` part using goroutines + WaitGroup (see 3.4 above).

This is where Go really shines. Your TypeScript version needed a library for this;
in Go it's 10 lines of stdlib.

### Phase 5 — State Persistence (`state/state.go`)

Replace n8n's `$getWorkflowStaticData('global')` with a JSON file:

```go
const stateFile = "/data/lessons.json"  // mount this as a Docker volume

func LoadState() ([]Lesson, error) {
    data, err := os.ReadFile(stateFile)
    if os.IsNotExist(err) {
        return []Lesson{}, nil  // first run — no previous state
    }
    var lessons []Lesson
    err = json.Unmarshal(data, &lessons)
    return lessons, err
}

func SaveState(lessons []Lesson) error {
    data, _ := json.Marshal(lessons)
    return os.WriteFile(stateFile, data, 0644)
}
```

### Phase 6 — Change Detection

Port your big `Change Detection` code node. It maps almost 1:1 to Go.

The field comparison loop (your `fieldsToCompare.forEach`) becomes a manual field-by-field
comparison in Go — there's no generic `obj[field]` access. That's intentional; Go is explicit.
You'll write something like:

```go
if prev.IsSubstitution != curr.IsSubstitution {
    changes = append(changes, Change{Field: "isSubstitution", From: prev.IsSubstitution, To: curr.IsSubstitution})
}
// repeat for each field
```

It's more verbose but immediately clear.

### Phase 7 — Notification + Main Loop (`notifier/ntfy.go`, `main.go`)

Port the ntfy.sh POST (simplest part — one `http.Post` call).

Then write `main.go` with a ticker loop:

```go
func main() {
    // run immediately on start, then every 15 minutes
    run()
    ticker := time.NewTicker(15 * time.Minute)
    for range ticker.C {
        run()
    }
}
```

---

## 5. Docker Setup

```dockerfile
# Build stage — full Go toolchain
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o notifier .

# Final stage — empty image
FROM scratch
COPY --from=builder /app/notifier /notifier
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
VOLUME ["/data"]
CMD ["/notifier"]
```

`CGO_ENABLED=0` produces a fully static binary. `FROM scratch` is literally an empty image —
no shell, no OS, just your binary. The `ca-certificates` copy is needed for HTTPS to work.

Build and run:
```bash
docker build -t untis-notifier .
docker run -v $(pwd)/data:/data \
  -e UNTIS_USER=youruser \
  -e UNTIS_PASS=yourpass \
  untis-notifier
```

Use environment variables for credentials — read them with `os.Getenv("UNTIS_USER")`.

Expected final image size: **~6MB** vs n8n's ~500MB+.

---

## 6. Useful Go Commands

```bash
go build ./...          # compile everything, catch errors
go run main.go          # run without building a binary
go vet ./...            # catch common mistakes (run this often)
go fmt ./...            # auto-format all files (do this before committing)
```

---

## 7. Resources

- **[go.dev/tour](https://go.dev/tour)** — interactive tour, worth doing the first 3 sections
- **[pkg.go.dev](https://pkg.go.dev)** — stdlib documentation (especially `net/http`, `encoding/json`, `time`)
- **[gobyexample.com](https://gobyexample.com)** — short, focused examples for every concept
- **[go.dev/doc/effective_go](https://go.dev/doc/effective_go)** — read after you're comfortable with basics

The standard library docs are excellent. When in doubt, read them before looking for a package.

---

## 8. Things That Will Trip You Up

| TypeScript habit | Go equivalent |
|---|---|
| `obj?.field` optional chaining | check for nil/zero value explicitly |
| `arr.filter(...)` | manual loop with `if` + `append` |
| `arr.map(...)` | manual loop building a new slice |
| `console.log` | `fmt.Println` or `log.Printf` |
| `JSON.parse` | `json.Unmarshal(&target)` — note the pointer |
| `typeof x === 'undefined'` | `x == nil` (for pointers/interfaces) |
| Implicit returns | must always explicitly `return` |
| `const` for objects | `var` — Go's `const` is only for primitives |

The biggest mental shift: **there is no exception handling**. Every function that can fail
returns an error. Check it every time. Don't use `_` to ignore errors until you understand why.