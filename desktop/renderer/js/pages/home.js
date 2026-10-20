const pages = {};

pages.home = async function (container) {
  const [directory, events, mentorships, jobs] = await Promise.all([
    api.get("/directory?role=alumni&limit=200"),
    api.get("/events?upcoming=true"),
    api.get("/mentorships"),
    api.get("/jobs"),
  ]);

  const pending = mentorships.results.filter((m) => m.status === "pending").length;

  container.innerHTML = `
    <div class="welcome">
      <h3>Welcome back, ${escapeHtml(state.user.name.split(" ")[0])}</h3>
      <p>${escapeHtml(welcomeLine(state.user.role))}</p>
    </div>

    <div class="stats">
      ${statCard("Alumni on the platform", directory.count)}
      ${statCard("Upcoming sessions", events.count)}
      ${statCard("Pending mentorship requests", pending)}
      ${statCard("Open job postings", jobs.count)}
    </div>

    <div class="card">
      <h3>Next session</h3>
      ${events.count === 0 ? emptyState("No upcoming sessions yet.") : nextEvent(events.results[0])}
    </div>`;
};

function welcomeLine(role) {
  if (role === "student") {
    return "Find a mentor, join a session and ask your questions in the forum.";
  }
  if (role === "alumni") {
    return "Students are looking for guidance. Check your mentorship requests.";
  }
  return "Department overview of alumni engagement.";
}

function statCard(label, value) {
  return `
    <div class="stat">
      <span class="stat-value">${escapeHtml(value)}</span>
      <span class="stat-label">${escapeHtml(label)}</span>
    </div>`;
}

function nextEvent(event) {
  return `
    <div class="event-row">
      <div>
        <strong>${escapeHtml(event.title)}</strong>
        <p>${escapeHtml(event.description)}</p>
        <small>${escapeHtml(formatDateTime(event.starts_at))} &middot; ${escapeHtml(event.location || "Online")}</small>
      </div>
      <span class="pill">${escapeHtml(event.kind)}</span>
    </div>`;
}
