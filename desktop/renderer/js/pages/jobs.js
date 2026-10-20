pages.jobs = async function (container) {
  const data = await api.get("/jobs");
  const canPost = state.user.role === "alumni" || state.user.role === "admin";

  container.innerHTML = `
    ${canPost ? `<button id="btn-new-job" class="btn btn-primary">Post an opening</button>` : ""}
    <div class="list spaced">
      ${data.count === 0 ? emptyState("No openings posted yet.") : data.results.map(jobCard).join("")}
    </div>`;

  const newButton = document.getElementById("btn-new-job");
  if (newButton) newButton.addEventListener("click", openJobForm);

  container.querySelectorAll("[data-delete-job]").forEach((button) => {
    button.addEventListener("click", async () => {
      try {
        await api.delete("/jobs/" + button.dataset.deleteJob);
        toast("Opening removed", "success");
        await openPage("jobs");
      } catch (error) {
        toast(error.message, "error");
      }
    });
  });
};

function jobCard(job) {
  const mine = job.posted_by === state.user.id;

  return `
    <div class="list-item">
      <div class="list-body">
        <div class="event-title">
          <strong>${escapeHtml(job.title)}</strong>
          <span class="pill">${escapeHtml(job.kind)}</span>
        </div>
        <p>${escapeHtml(job.description)}</p>
        <small>
          ${escapeHtml(job.company)}
          ${job.location ? "&middot; " + escapeHtml(job.location) : ""}
          &middot; posted by ${escapeHtml(job.poster_name)}
          &middot; ${escapeHtml(formatDate(job.created_at))}
        </small>
        ${job.apply_link ? `<small class="link">${escapeHtml(job.apply_link)}</small>` : ""}
      </div>
      <div class="list-actions">
        ${mine ? `<button class="btn btn-danger btn-small" data-delete-job="${job.id}">Remove</button>` : ""}
      </div>
    </div>`;
}

function openJobForm() {
  showModal(
    "Post an opening",
    `<form id="job-form" class="form">
       <label>Job title<input name="title" required /></label>
       <label>Company<input name="company" required /></label>
       <label>Location<input name="location" /></label>
       <label>Type
         <select name="kind">
           <option value="full-time">Full time</option>
           <option value="internship">Internship</option>
         </select>
       </label>
       <label>Description<textarea name="description"></textarea></label>
       <label>Application link<input name="apply_link" /></label>
       <button type="submit" class="btn btn-primary">Publish opening</button>
     </form>`
  );

  document.getElementById("job-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = Object.fromEntries(new FormData(event.target));

    try {
      await api.post("/jobs", data);
      closeModal();
      toast("Opening published", "success");
      await openPage("jobs");
    } catch (error) {
      toast(error.message, "error");
    }
  });
}
