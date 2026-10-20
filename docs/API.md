# API reference

Base URL: `http://localhost:8080/api`

Every endpoint except `/health`, `/register` and `/login` needs a header:

```
Authorization: Bearer <token>
```

Errors always come back in the same shape, so the client has one thing to check:

```json
{ "error": "incorrect email or password" }
```

---

## Health

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/health` | No auth. Used to check the server is up. |

---

## Accounts

| Method | Path | Body | Notes |
| --- | --- | --- | --- |
| POST | `/register` | `name, email, password, role` | `role` is `student` or `alumni`. Returns `{token, user}`. |
| POST | `/login` | `email, password` | Returns `{token, user}`. |
| POST | `/logout` | — | Deletes the session row. |
| GET | `/me` | — | Returns `{user, profile}`. |

Password rules: at least 8 characters. Stored as a bcrypt hash.

---

## Profiles

| Method | Path | Body | Notes |
| --- | --- | --- | --- |
| PUT | `/me/profile` | all profile fields | Upsert. Bio is passed through the moderation filter. |
| GET | `/users/{id}` | — | One person's public card. |

Profile fields: `college, branch, graduation_year, company, job_title,
industry, location, skills, interests, bio, linkedin, open_to_mentor`.

`skills` and `interests` are comma separated text, e.g. `"Go, React, SQL"`.

---

## Directory

| Method | Path | Query | Notes |
| --- | --- | --- | --- |
| GET | `/directory` | see below | Search and filter everyone. |
| GET | `/directory/filters` | — | Distinct branches, industries, companies, locations, years. |

Query parameters, all optional and combinable:

| Param | Example | Effect |
| --- | --- | --- |
| `q` | `q=google` | Free text over name, company, job title, skills, industry, branch |
| `role` | `role=alumni` | Exact role |
| `branch` | `branch=Computer Science` | Exact branch |
| `year` | `year=2016` | Graduation year |
| `industry`, `company`, `location` | | Exact match |
| `mentor` | `mentor=true` | Only alumni open to mentoring |
| `verified` | `verified=true` | Only department-verified accounts |
| `limit`, `offset` | `limit=20&offset=20` | Paging, max 200 per page |

---

## Suggested connections

| Method | Path | Query | Notes |
| --- | --- | --- | --- |
| GET | `/recommendations` | `limit` (default 5, max 50) | Students get alumni, alumni get students. |

Each result is a directory entry plus:

```json
{
  "score": 77,
  "reasons": ["2 shared skills: Go, Python", "Same branch: Computer Science"]
}
```

---

## Mentorship

| Method | Path | Body | Who |
| --- | --- | --- | --- |
| POST | `/mentorships` | `alumni_id, message` | Students only |
| GET | `/mentorships` | — | Anyone; returns only your own requests |
| PATCH | `/mentorships/{id}` | `status` | Only the mentor named on the request |

`status` is `accepted` or `rejected`. A request can only be answered once.

---

## Events

| Method | Path | Body / Query | Who |
| --- | --- | --- | --- |
| GET | `/events` | `?upcoming=true` | Anyone |
| POST | `/events` | `title, description, kind, starts_at, location, link` | Alumni and admins |
| POST | `/events/{id}/rsvp` | — | Anyone |
| DELETE | `/events/{id}/rsvp` | — | Anyone |

`kind` is `meetup`, `webinar` or `panel`. `starts_at` is RFC3339,
e.g. `2026-11-15T10:00:00Z`.

---

## Forum

| Method | Path | Body / Query | Who |
| --- | --- | --- | --- |
| GET | `/posts` | `?category=career&q=interview` | Anyone |
| POST | `/posts` | `title, body, category` | Anyone |
| GET | `/posts/{id}` | — | Returns `{post, comments}` |
| DELETE | `/posts/{id}` | — | Author or admin |
| POST | `/posts/{id}/comments` | `body` | Anyone |

`category` is `general`, `career`, `academic` or `placement`.
Titles, bodies and comments all run through the moderation filter before
they are stored; the stored text is the masked version and `flagged` is set.

---

## Placements

| Method | Path | Body / Query | Who |
| --- | --- | --- | --- |
| GET | `/jobs` | `?kind=internship&q=google` | Anyone |
| POST | `/jobs` | `title, company, location, kind, description, apply_link` | Alumni and admins |
| DELETE | `/jobs/{id}` | — | Poster or admin |

`kind` is `full-time` or `internship`.
