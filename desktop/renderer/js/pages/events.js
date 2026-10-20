pages.events = async function (container) {
  const data = await api.get("/events");
  const canCreate = state.user.role === "alumni" || state.user.role === "admin";

  container.innerHTML = `
    ${canCreate ? `<button id="btn-new-event" class="btn btn-primary">Schedule a session</button>` : ""}
    <div class="list spaced">
      ${data.count === 0 ? emptyState("No sessions scheduled yet.") : data.results.map(eventCard).join("")}
    </div>`;

  container.querySelectorAll("[data-rsvp]").forEach((button) => {
    button.addEventListener("click", async () => {
      const going = button.dataset.going === "true";
      try {
        await api.request(going ? "DELETE" : "POST", "/events/" + button.dataset.rsvp + "/rsvp");
        await openPage("events");
      } catch (error) {
        toast(error.message, "error");
      }
    });
  });

  const newButton = document.getElementById("btn-new-event");
  if (newButton) newButton.addEventListener("click", openEventForm);
};

function eventCard(event) {
  const past = new Date(event.starts_at) < new Date();

  return `
    <div class="list-item event">
      <div class="list-body">
        <div class="event-title">
          <strong>${escapeHtml(event.title)}</strong>
          <span class="pill">${escapeHtml(event.kind)}</span>
          ${past ? '<span class="badge badge-pending">past</span>' : ""}
        </div>
        <p>${escapeHtml(event.description)}</p>
        <small>
          ${escapeHtml(formatDateTime(event.starts_at))}
          &middot; ${escapeHtml(event.location || "Online")}
          &middot; by ${escapeHtml(event.creator_name)}
          &middot; ${escapeHtml(event.rsvp_count)} attending
        </small>
      </div>
      <div class="list-actions">
        ${
          past
            ? ""
            : `<button class="btn ${event.is_attending ? "btn-danger" : "btn-secondary"} btn-small"
                 data-rsvp="${event.id}" data-going="${event.is_attending}">
                 ${event.is_attending ? "Cancel RSVP" : "Attend"}
               </button>`
        }
      </div>
    </div>`;
}

function openEventForm() {
  showModal(
    "Schedule a session",
    `<form id="event-form" class="form">
       <label>Title<input name="title" required /></label>
       <label>Description<textarea name="description"></textarea></label>
       <label>Type
         <select name="kind">
           <option value="webinar">Webinar</option>
           <option value="meetup">Alumni meetup</option>
           <option value="panel">Panel discussion</option>
         </select>
       </label>
       <label>Starts at<input name="starts_at" type="datetime-local" required /></label>
       <label>Location<input name="location" placeholder="Leave blank for online" /></label>
       <label>Joining link<input name="link" /></label>
       <button type="submit" class="btn btn-primary">Create session</button>
     </form>`
  );

  document.getElementById("event-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = Object.fromEntries(new FormData(event.target));
    data.starts_at = new Date(data.starts_at).toISOString();

    try {
      await api.post("/events", data);
      closeModal();
      toast("Session created", "success");
      await openPage("events");
    } catch (error) {
      toast(error.message, "error");
    }
  });
}
