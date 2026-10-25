AlumniSetu — Alumni & Student Interaction Platform for the Technical Education Department, Govt. of Rajasthan.

## Download

**`AlumniSetu-1.0.0.dmg`** — 7 MB, macOS 11 or later, runs on both Apple Silicon and Intel Macs.

## Install

1. Open the `.dmg` and drag **AlumniSetu** into Applications.
2. The build is not code-signed, so the first launch needs **right-click → Open**, then confirm. (Or System Settings → Privacy & Security → Open Anyway.)
3. Launch it. The Go API server is bundled inside the app and starts automatically — there is nothing else to install and no terminal to run.

Data is stored at `~/Library/Application Support/AlumniSetu/alumnisetu.db`. Delete that file to reset to fresh demo data.

## Demo accounts

All use the password `password123`.

| Role | Email |
| --- | --- |
| Student | `riya@student.ac.in` |
| Alumni | `arjun@alumni.ac.in` |
| Department admin | `admin@rtu.ac.in` |

## What is in this build

Role-based accounts, centralized alumni directory with search and filters, explainable mentor matching, mentorship requests, events and RSVPs, discussion forum, automatic content moderation, placement board, in-app notifications, skill extraction from free-text bios, a tamper-evident verification ledger, an admin console, a department analytics dashboard, and a help assistant.

Built with a Go standard-library API over SQLite and a native macOS shell (Swift + WKWebView) around an interface written in plain HTML, CSS and JavaScript. The whole app is 7 MB because it uses the WebKit already present in macOS rather than bundling a browser engine.
