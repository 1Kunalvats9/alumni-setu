let forumCategory = "";

pages.forum = async function (container) {
  const data = await api.get("/posts" + (forumCategory ? "?category=" + forumCategory : ""));

  container.innerHTML = `
    <div class="card">
      <div class="toolbar">
        <select id="forum-category">
          <option value="">All topics</option>
          <option value="general">General</option>
          <option value="career">Career</option>
          <option value="academic">Academic</option>
          <option value="placement">Placement</option>
        </select>
        <button id="btn-new-post" class="btn btn-primary">Start a discussion</button>
      </div>
    </div>
    <div class="list spaced">
      ${data.count === 0 ? emptyState("No discussions here yet.") : data.results.map(postRow).join("")}
    </div>`;

  document.getElementById("forum-category").value = forumCategory;
  document.getElementById("forum-category").addEventListener("change", async (event) => {
    forumCategory = event.target.value;
    await openPage("forum");
  });
  document.getElementById("btn-new-post").addEventListener("click", openPostForm);

  container.querySelectorAll("[data-post]").forEach((row) => {
    row.addEventListener("click", () => openPostModal(row.dataset.post));
  });
};

function postRow(post) {
  return `
    <div class="list-item clickable" data-post="${post.id}">
      <div class="avatar dark">${escapeHtml(initials(post.author_name))}</div>
      <div class="list-body">
        <div class="event-title">
          <strong>${escapeHtml(post.title)}</strong>
          <span class="pill">${escapeHtml(post.category)}</span>
          ${post.flagged ? '<span class="badge badge-flagged">flagged</span>' : ""}
        </div>
        <p>${escapeHtml(post.body)}</p>
        <small>
          ${escapeHtml(post.author_name)} ${roleBadge(post.author_role)}
          &middot; ${escapeHtml(formatDate(post.created_at))}
          &middot; ${escapeHtml(post.comment_count)} repl${post.comment_count === 1 ? "y" : "ies"}
        </small>
      </div>
    </div>`;
}

async function openPostModal(id) {
  const data = await api.get("/posts/" + id);
  const post = data.post;

  showModal(
    post.title,
    `<div class="post-detail">
       <small>${escapeHtml(post.author_name)} ${roleBadge(post.author_role)} &middot; ${escapeHtml(formatDate(post.created_at))}</small>
       <p class="bio">${escapeHtml(post.body)}</p>

       <h4>${escapeHtml(data.comments.length)} repl${data.comments.length === 1 ? "y" : "ies"}</h4>
       <div class="comments">
         ${data.comments.length === 0 ? emptyState("Be the first to reply.") : data.comments.map(commentRow).join("")}
       </div>

       <form id="comment-form" class="form request-form">
         <label>Your reply<textarea name="body" required></textarea></label>
         <button type="submit" class="btn btn-primary">Post reply</button>
       </form>
     </div>`
  );

  document.getElementById("comment-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const body = new FormData(event.target).get("body");

    try {
      await api.post("/posts/" + id + "/comments", { body });
      closeModal();
      toast("Reply posted", "success");
      await openPage("forum");
    } catch (error) {
      toast(error.message, "error");
    }
  });
}

function commentRow(comment) {
  return `
    <div class="comment">
      <div class="avatar small">${escapeHtml(initials(comment.author_name))}</div>
      <div>
        <small>${escapeHtml(comment.author_name)} ${roleBadge(comment.author_role)}</small>
        <p>${escapeHtml(comment.body)}</p>
      </div>
    </div>`;
}

function openPostForm() {
  showModal(
    "Start a discussion",
    `<form id="post-form" class="form">
       <label>Title<input name="title" required /></label>
       <label>Topic
         <select name="category">
           <option value="general">General</option>
           <option value="career">Career</option>
           <option value="academic">Academic</option>
           <option value="placement">Placement</option>
         </select>
       </label>
       <label>Your question<textarea name="body" required></textarea></label>
       <p class="muted">Posts are automatically screened for abusive or spam content.</p>
       <button type="submit" class="btn btn-primary">Publish</button>
     </form>`
  );

  document.getElementById("post-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = Object.fromEntries(new FormData(event.target));

    try {
      const post = await api.post("/posts", data);
      closeModal();
      toast(post.flagged ? "Posted, but flagged for review" : "Discussion published", "success");
      await openPage("forum");
    } catch (error) {
      toast(error.message, "error");
    }
  });
}
