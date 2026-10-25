# Code walkthrough

A guide for explaining this project out loud. Read it top to bottom and you
can answer "how does it work?" for any part of the app.

---

## 1. The big picture

There are exactly two programs:

1. **The server** (`server/`) — a Go program that owns the database and
   answers HTTP requests with JSON. It knows nothing about screens.
2. **The desktop app** (`desktop/`) — an Electron window showing an HTML page.
   It knows nothing about SQL. It only calls the server over HTTP.

They talk over `http://localhost:8080`. That is the whole architecture.

> If someone asks "why split it?" — because the same server could later power a
> web app or a mobile app without a single change.

---

## 2. The server, file by file

| File | One-line job |
| --- | --- |
| `main.go` | Lists every URL and which function answers it |
| `db.go` | Opens the SQLite file and creates the tables |
| `models.go` | Go structs matching the tables, with JSON names |
| `helpers.go` | Small shared utilities (JSON in/out, list handling) |
| `auth.go` | Register, login, logout, and "who is calling?" |
| `profiles.go` | Read and update a person's details |
| `directory.go` | Search and filter the alumni database |
| `recommend.go` | Score and rank suggested connections |
| `mentorship.go` | Mentorship requests |
| `events.go` | Meetups, webinars, panels and RSVPs |
| `forum.go` | Discussion threads and replies |
| `jobs.go` | Placement board |
| `moderation.go` | Blocks and masks harmful content |
| `seed.go` | Demo data for the first run |

### 2.1 How a request flows

Take `GET /api/directory?q=google`:

1. `main.go` matched the pattern `"GET /api/directory"` and calls
   `requireAuth(handleDirectory)`.
2. `requireAuth` (in `auth.go`) reads the `Authorization: Bearer ...` header,
   looks the token up in the `sessions` table joined to `users`, and either
   returns `401` or calls the handler with the logged-in user.
3. `handleDirectory` builds a SQL `WHERE` clause, runs the query, and
   `writeJSON` sends the rows back.

Every protected endpoint follows that same three-step shape.

### 2.2 How login works

Three steps, no JWT library:

1. **Register** — `bcrypt.GenerateFromPassword` turns the password into a
   hash. Only the hash is stored. The original password is never saved.
2. **Login** — `bcrypt.CompareHashAndPassword` checks the typed password
   against the hash. On success we generate 64 random hex characters and
   insert a row into `sessions`.
3. **Every later request** — the client sends that token; we join `sessions`
   to `users` to find out who it is.

**Why not JWT?** A JWT cannot be revoked without extra machinery. Here,
logging out is one `DELETE` and the token is dead immediately.

**Why is the login error always "incorrect email or password"?** So nobody
can use the login form to find out which emails are registered.

### 2.3 How the search avoids SQL injection

`handleDirectory` never pastes user text into the SQL string. It appends
a `?` to the query and the value to a separate `args` slice:

```go
where = append(where, "p.graduation_year = ?")
args  = append(args, year)
```

The driver sends the SQL and the values separately, so a search for
`' OR 1=1--` is treated as ordinary text to look for. There is a test for
exactly this in the commit history.

### 2.4 How matching works (`recommend.go`)

This is the "suggest connections" requirement. It is a **weighted overlap
score**, not a trained model:

| Signal | Points |
| --- | --- |
| Each shared skill | 10 |
| Each shared interest | 6 |
| Same branch | 15 |
| Works in an industry you follow | 8 |
| Same college | 5 |
| Open to mentoring | 12 |
| Verified profile | 5 |

Everyone with a score above zero is sorted highest-first and the top N
returned. Alongside the number, each match carries the **reasons** that
produced it — `"2 shared skills: Go, Python"`.

**Be ready for "why not machine learning?"** Three honest answers:

1. A brand new platform has no training data. This works on day one.
2. A student trusts a reason far more than a number from a black box.
3. The weights are in one block of constants — that is exactly where a
   trained model would later plug in, with no other code changing.

Say it is a **content-based recommender**, which is the accurate name for it.

### 2.5 How moderation works (`moderation.go`)

`moderate(text)` returns the cleaned text and whether it was flagged.

- A single regular expression is built from the blocked word list, with
  `\b` word boundaries — so `"scam"` does **not** match `"Scandinavian"`.
- Matching ignores case; matched words are replaced with asterisks, keeping
  the surrounding text and newlines intact.
- It also flags more than three links (spam) and text that is over 70%
  capitals (shouting).

It is applied to profile bios, forum posts, comments and job descriptions.
The **masked** text is what gets stored, and the `flagged` column marks the
row for an admin to review.

`moderation_test.go` covers 10 cases. Run them with:

```bash
cd server && go test ./...
```

---

## 3. The desktop app, file by file

| File | One-line job |
| --- | --- |
| `main.js` | Creates the Electron window |
| `preload.js` | Safely hands the page the API address |
| `renderer/index.html` | The whole page skeleton |
| `renderer/styles.css` | All styling |
| `renderer/js/api.js` | One `fetch` wrapper used by everything |
| `renderer/js/ui.js` | Escaping, dates, badges, modal, toast |
| `renderer/js/auth.js` | The login and register forms |
| `renderer/js/app.js` | Navigation and app state |
| `renderer/js/pages/*.js` | One file per screen |

### 3.1 Why there is no React

Every page is a function that takes a container element and sets its
`innerHTML`. That is the entire framework:

```js
pages.jobs = async function (container) {
  const data = await api.get("/jobs");
  container.innerHTML = data.results.map(jobCard).join("");
};
```

No build step, no bundler, no node_modules in the UI. Edit a file, reload
the window, see the change.

### 3.2 How navigation works

`app.js` has a `navItems` array. `openPage(name)` looks up
`pages[name]` and calls it. Adding a screen is two lines: one entry in
`navItems`, one file in `pages/`.

### 3.3 The three security details worth mentioning

1. **`contextIsolation: true` and `nodeIntegration: false`** in `main.js`.
   The page cannot touch Node or the filesystem. `preload.js` exposes one
   thing — the API address — and nothing else.
2. **A Content-Security-Policy meta tag** in `index.html` restricts the page
   to its own files plus connections to `localhost:8080`.
3. **`escapeHtml()` on every value** before it goes into `innerHTML`, so a
   user who types `<script>` into their bio gets text, not code.

### 3.4 Where the login token lives

`api.js` keeps it in `localStorage`, so closing and reopening the app keeps
you signed in. On startup `app.js` calls `/me`; if the token was revoked,
the call fails and the app falls back to the login screen.

---

## 4. The database

Nine tables. The interesting parts:

- **`users.verified`** — the anti-fake-profile flag. An admin flips it to 1
  after checking the person's degree or employment. The directory can filter
  on it and matching gives verified profiles a bonus.
- **`profiles`** is separate from `users` because login data and directory
  data change for different reasons and at different times.
- **`event_rsvps`** has a composite primary key `(event_id, user_id)`, so
  clicking Attend twice cannot create a duplicate row.
- **`ON DELETE CASCADE`** everywhere, with `PRAGMA foreign_keys = ON` in
  `db.go` — delete a post and its comments go with it, automatically.

---

## 5. Likely questions

**"Where is the AI?"**
The matching engine. Be straight about it: it is a content-based recommender
with transparent weights, not a neural network. That is a deliberate choice
for a cold-start platform, and `recommend.go` is the single place a trained
model would replace. An NLP layer to pull skills out of free-text bios is the
next step and is listed on the roadmap.

**"How do you stop fake profiles?"**
Today: role-based accounts, a `verified` flag only an admin can set, and
bcrypt passwords. The directory can show verified profiles only. Blockchain
credential verification is on the roadmap, not built.

**"What happens if the server is not running?"**
`api.js` catches the failed fetch and shows "Cannot reach the server. Is it
running on port 8080?" instead of a blank screen.

**"How much is actually built?"**
Roughly a third of the full platform: the database, accounts, directory,
matching, mentorship, events, forum, moderation and placements, end to end
from SQL to screen. `docs/ROADMAP.md` lists exactly what is not built yet.

---

## 6. The admin side (`admin.go`)

This is where the fake-profile and harmful-content requirements are actually
closed out, because a filter only helps if a human acts on what it catches.

Every route is wrapped in `requireRole([]string{"admin"}, ...)`, so a student
or alumni token gets `403` before the handler ever runs.

| Endpoint | What it does |
| --- | --- |
| `/admin/stats` | Counts across accounts, events, forum, placements, mentorships |
| `/admin/pending` | Alumni whose `verified` flag is still 0 |
| `/admin/users/{id}/verify` | Flips that flag |
| `/admin/flagged` | Everything `moderate()` marked |
| `/admin/posts/{id}/approve` | Clears a false positive |

**The anti-fake-profile answer, in one sentence:** accounts start unverified,
an admin checks the claimed degree and employer against the profile, and only
then does the account get a verified badge, rank higher in matching, and pass
the directory's `verified=true` filter.

One guard worth pointing out: `handleAdminVerify` refuses to touch an account
whose role is `admin`, so the endpoint cannot be used to tamper with the
admin accounts themselves.

---

## 7. The help assistant (`assistant.go`)

The problem statement asks for a chatbot. This is **keyword retrieval over a
knowledge base of 15 intents** — say that plainly rather than calling it AI.

Each intent is a struct with keywords, an answer and the page it relates to:

```go
{
    Name:     "find_mentor",
    Keywords: []string{"mentor", "guidance", "guide", "senior", "advice"},
    Text:     "Open Suggested Mentors...",
    Page:     "mentors",
}
```

`bestIntent(question)` lowercases the question and scores every intent by how
many of its keywords appear in it. A multi-word keyword scores its word count,
so `"host a session"` (3 points) beats a bare `"session"` (1 point) — that is
how a specific phrase wins over a general one. Highest score wins, and a score
of zero returns a list of what the assistant can help with instead of guessing.

Six intents set `Live` instead of `Text` — a function that queries the database
— so "when is the next session" and "do I have pending requests" return the
real answer for the person asking, not a canned sentence.

**If asked "why not a real LLM?"**: this needs no API key, no network, no cost
and no risk of inventing an answer about a feature that does not exist. `Live`
and `Text` are the only two things a model would replace, and it is one
function. That trade is a reasonable one to defend.

`assistant_test.go` covers 15 phrasings, including two that must *not* match.

---

## 8. Running the tests

```bash
cd server
go test ./...     # 25 cases: 10 moderation, 15 assistant intents
go vet ./...
```

---

## 9. Notifications (`notifications.go`)

Two helpers do all the work: `notify()` for one person, and
`notifyEveryoneExcept()` which inserts a row per user with a single
`INSERT ... SELECT` rather than looping in Go. They are called from the
handlers that already exist, so no feature had to be restructured.

One detail worth knowing because it was a real bug: ordering is
`created_at DESC, id DESC`. Timestamps are second-resolution, so several
notifications written in the same second came back in an arbitrary order
and the list looked oldest-first. The `id` tiebreaker fixes it.

Delete is scoped in SQL — `WHERE id = ? AND user_id = ?` — so guessing
another person's notification id does nothing.

---

## 10. Skill extraction (`nlp.go`)

This is the NLP piece, and the accurate name for it is **gazetteer matching**:
a dictionary of 52 canonical skills, each with aliases, compiled into one
regular expression per skill. `golang` resolves to Go, `k8s` to Kubernetes,
`deep learning` to Machine Learning.

The whole difficulty is word boundaries, and the tests are the specification:

| Input | Must | Must not |
| --- | --- | --- |
| `I had to google the algorithm` | | yield Go |
| `We kept the code pythonic` | | yield Python |
| `Competitive programming in C++.` | yield C++ | also yield bare C |
| `Stack: React, Node.js, SQL.` | yield all three | |

The boundary is asymmetric, which is the non-obvious part. On the left we
exclude `.` so `js` inside `Node.js` is not read as JavaScript. On the right
we must **not** exclude `.`, because a skill at the end of a sentence is
followed by a full stop. The first version excluded it on both sides and
silently missed every skill that ended a sentence.

---

## 11. The verification ledger (`ledger.go`)

The problem statement says "blockchain or any authenticated technique". This
is the data structure underneath a blockchain — a SHA-256 hash chain —
without the network, the consensus or the distribution, none of which would
do anything useful for a single department database. Call it a hash chain.

Each record stores a fingerprint of its own contents **plus the previous
record's fingerprint**. `auditLedger()` walks the chain and recomputes
everything. That makes three attacks visible:

| Attack | How it shows up |
| --- | --- |
| Edit a record | its stored fingerprint no longer matches its contents |
| Edit it *and* recompute its fingerprint | the next record's `prev_hash` no longer matches |
| Delete a record | the chain jumps and the link breaks |

**This is the best thing to demo live.** Open the Admin Console, show
"Chain intact". Then, in a terminal:

```bash
sqlite3 ~/Library/Application\ Support/AlumniSetu/alumnisetu.db \
  "UPDATE ledger SET details = 'forged approval' WHERE id = 2;"
```

Restart and reload: the console reads "Tampering detected at record #2".

Be ready for the honest limitation: the server operator could rewrite the
whole chain from a given point. Defending against that needs the hashes
anchored somewhere outside the server, which is on the roadmap and is
exactly what a real blockchain buys you.

---

## 12. Analytics (`analytics.go` and `pages/analytics.js`)

Eight breakdowns, each a `SELECT ... GROUP BY` returning label/value pairs.

The chart choice is worth being able to defend. Every one of these counts a
single thing, so the form is a ranked horizontal bar with a zero baseline,
**one hue for every bar**, and no legend — the title already names the
series. Shading bars darker-where-bigger is a common instinct and is wrong:
it encodes bar length twice and spends the only free channel on information
the length already carries.

A CSP detail caught by screenshotting rather than by assertions: the page
sets `style-src 'self'`, which blocks `style="width: 40%"` in HTML. DOM
tests passed while every bar rendered at zero width. Widths are therefore
set with `element.style.width`, which CSP does not restrict.

---

## 13. Packaging

`npm run dist` runs `build-server.sh` (cross-compiles the Go server for both
architectures and `lipo`s them into one universal binary) and then
electron-builder.

Two things make the bundle actually work:

1. **The app bundle is read-only.** The server takes its database path from
   `ALUMNISETU_DB`, and `main.js` points it at `app.getPath("userData")`.
2. **`main.js` health-checks port 8080 before spawning.** If you are running
   from source with `go run .`, the packaged app will not start a second
   copy, and development is unchanged.
