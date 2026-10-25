pages.admin = async function (container) {
  const [stats, pending, flagged, ledger] = await Promise.all([
    api.get("/admin/stats"),
    api.get("/admin/pending"),
    api.get("/admin/flagged"),
    api.get("/admin/ledger"),
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

  container.insertAdjacentHTML("beforeend", ledgerSection(ledger));
  wireAdminActions(container);
};

function ledgerSection(ledger) {
  return `
    <div class="card">
      <h3>Verification ledger (${ledger.count})</h3>
      <p class="muted">
        Every verification decision is appended to a hash chain. Each record carries a
        fingerprint of its own contents plus the previous record's fingerprint, so editing
        or deleting history breaks the chain and shows up here.
      </p>

      <div class="chain-status ${ledger.intact ? "ok" : "broken"}">
        ${
          ledger.intact
            ? "Chain intact &mdash; all " + escapeHtml(ledger.count) + " records verify"
            : "Tampering detected at record #" + escapeHtml(ledger.broken_at) + " &mdash; " + escapeHtml(ledger.reason)
        }
      </div>

      <div class="list spaced">
        ${
          ledger.count === 0
            ? emptyState("No verification decisions recorded yet.")
            : ledger.results.slice().reverse().map(ledgerRow).join("")
        }
      </div>
    </div>`;
}

function ledgerRow(entry) {
  const granted = entry.action === "verification_granted";

  return `
    <div class="list-item">
      <div class="list-body">
        <div class="event-title">
          <strong>#${escapeHtml(entry.id)} ${escapeHtml(entry.subject_name)}</strong>
          <span class="badge ${granted ? "badge-accepted" : "badge-rejected"}">
            ${granted ? "verified" : "revoked"}
          </span>
        </div>
        <p>${escapeHtml(entry.details)}</p>
        <small>by ${escapeHtml(entry.actor_name)} &middot; ${escapeHtml(formatDateTime(entry.created_at))}</small>
        <small class="hash">prev ${escapeHtml(entry.prev_hash.slice(0, 16))}… &rarr; this ${escapeHtml(entry.hash.slice(0, 16))}…</small>
      </div>
    </div>`;
}

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
