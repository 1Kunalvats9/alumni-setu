pages.admin = async function (container) {
  const [stats, pending, flagged] = await Promise.all([
    api.get("/admin/stats"),
    api.get("/admin/pending"),
    api.get("/admin/flagged"),
  ]);

  container.innerHTML = `
    <div class="card">
      <h3>Engagement overview</h3>
      <div class="stats">
        ${statCard("Students", stats.students)}
        ${statCard("Alumni", stats.alumni)}
        ${statCard("Verified alumni", stats.verified_alumni)}
        ${statCard("Open to mentoring", stats.open_to_mentor)}
        ${statCard("Active mentorships", stats.mentorships_active)}
        ${statCard("Pending requests", stats.mentorships_pending)}
        ${statCard("Sessions held", stats.events)}
        ${statCard("Session RSVPs", stats.rsvps)}
        ${statCard("Forum posts", stats.posts)}
        ${statCard("Forum replies", stats.comments)}
        ${statCard("Job postings", stats.jobs)}
        ${statCard("Awaiting review", stats.flagged_posts + stats.flagged_comments)}
      </div>
    </div>

    <div class="card">
      <h3>Alumni awaiting verification (${pending.count})</h3>
      <p class="muted">Verifying an account marks it as genuine in the directory and boosts it in mentor matching.</p>
      <div class="list spaced">
        ${pending.count === 0 ? emptyState("No accounts waiting.") : pending.results.map(pendingRow).join("")}
      </div>
    </div>

    <div class="card">
      <h3>Content flagged by the filter (${flagged.posts.length + flagged.comments.length})</h3>
      <p class="muted">The filter already masked the offending words. Approve clears a false positive, Delete removes the content.</p>
      <div class="list spaced">
        ${
          flagged.posts.length + flagged.comments.length === 0
            ? emptyState("Nothing is waiting for review.")
            : flagged.posts.map(flaggedPostRow).join("") + flagged.comments.map(flaggedCommentRow).join("")
        }
      </div>
    </div>`;

  wireAdminActions(container);
};

function pendingRow(person) {
  const p = person.profile;
  const line = [p.job_title, p.company].filter(Boolean).join(" at ");

  return `
    <div class="list-item">
      <div class="avatar dark">${escapeHtml(initials(person.name))}</div>
      <div class="list-body">
        <strong>${escapeHtml(person.name)}</strong>
        <p>${escapeHtml(line || "No employment details given")}</p>
        <small>
          ${escapeHtml(person.email)}
          ${p.branch ? "&middot; " + escapeHtml(p.branch) : ""}
          ${p.graduation_year ? "&middot; batch of " + escapeHtml(p.graduation_year) : ""}
          ${p.college ? "&middot; " + escapeHtml(p.college) : ""}
        </small>
      </div>
      <div class="list-actions">
        <button class="btn btn-primary btn-small" data-verify="${person.id}">Verify</button>
      </div>
    </div>`;
}

function flaggedPostRow(post) {
  return `
    <div class="list-item">
      <div class="list-body">
        <div class="event-title">
          <strong>${escapeHtml(post.title)}</strong>
          <span class="badge badge-flagged">post</span>
        </div>
        <p>${escapeHtml(post.body)}</p>
        <small>${escapeHtml(post.author_name)} ${roleBadge(post.author_role)} &middot; ${escapeHtml(formatDate(post.created_at))}</small>
      </div>
      <div class="list-actions">
        <button class="btn btn-secondary btn-small" data-approve-post="${post.id}">Approve</button>
        <button class="btn btn-danger btn-small" data-delete-post="${post.id}">Delete</button>
      </div>
    </div>`;
}

function flaggedCommentRow(comment) {
  return `
    <div class="list-item">
      <div class="list-body">
        <div class="event-title">
          <strong>Reply on thread #${escapeHtml(comment.post_id)}</strong>
          <span class="badge badge-flagged">reply</span>
        </div>
        <p>${escapeHtml(comment.body)}</p>
        <small>${escapeHtml(comment.author_name)} ${roleBadge(comment.author_role)} &middot; ${escapeHtml(formatDate(comment.created_at))}</small>
      </div>
      <div class="list-actions">
        <button class="btn btn-danger btn-small" data-delete-comment="${comment.id}">Delete</button>
      </div>
    </div>`;
}

function wireAdminActions(container) {
  const actions = [
    ["data-verify", (id) => api.patch("/admin/users/" + id + "/verify", { verified: true }), "Account verified"],
    ["data-approve-post", (id) => api.patch("/admin/posts/" + id + "/approve"), "Post approved"],
    ["data-delete-post", (id) => api.delete("/posts/" + id), "Post deleted"],
    ["data-delete-comment", (id) => api.delete("/comments/" + id), "Reply deleted"],
  ];

  actions.forEach(([attribute, call, message]) => {
    container.querySelectorAll("[" + attribute + "]").forEach((button) => {
      button.addEventListener("click", async () => {
        try {
          await call(button.getAttribute(attribute));
          toast(message, "success");
          await openPage("admin");
        } catch (error) {
          toast(error.message, "error");
        }
      });
    });
  });
}
