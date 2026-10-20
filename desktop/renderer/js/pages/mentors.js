pages.mentors = async function (container) {
  const data = await api.get("/recommendations?limit=10");

  const heading = state.user.role === "alumni" ? "Students you could mentor" : "Alumni matched to your profile";

  if (data.count === 0) {
    container.innerHTML = `
      <div class="card"><h3>${escapeHtml(heading)}</h3>
        ${emptyState("Add your branch, skills and interests in My Profile to get suggestions.")}
      </div>`;
    return;
  }

  container.innerHTML = `
    <div class="card">
      <h3>${escapeHtml(heading)}</h3>
      <p class="muted">Matched on shared skills, interests, branch, industry and college. Every suggestion shows why it was made.</p>
    </div>
    <div class="grid">${data.results.map(recommendationCard).join("")}</div>`;

  wirePersonCards(container);
};

function recommendationCard(person) {
  const p = person.profile;
  const line = [p.job_title, p.company].filter(Boolean).join(" at ");

  return `
    <div class="person">
      <div class="person-head">
        <div class="avatar dark">${escapeHtml(initials(person.name))}</div>
        <div>
          <strong>${escapeHtml(person.name)}</strong>
          <div class="person-meta">
            ${roleBadge(person.role)}
            ${person.verified ? '<span class="badge badge-verified">verified</span>' : ""}
          </div>
        </div>
        <span class="score" title="Match score">${escapeHtml(person.score)}</span>
      </div>
      <p class="person-role">${escapeHtml(line || "Profile not completed yet")}</p>
      <ul class="reasons">
        ${person.reasons.map((reason) => `<li>${escapeHtml(reason)}</li>`).join("")}
      </ul>
      <button class="btn btn-secondary btn-small" data-person="${person.id}">View profile</button>
    </div>`;
}
