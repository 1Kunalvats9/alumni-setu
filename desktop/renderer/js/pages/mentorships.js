pages.mentorships = async function (container) {
  const data = await api.get("/mentorships");

  if (data.count === 0) {
    container.innerHTML = emptyState(
      state.user.role === "student"
        ? "You have not requested a mentor yet. Try Suggested Mentors."
        : "No student has requested you as a mentor yet."
    );
    return;
  }

  container.innerHTML = `<div class="list">${data.results.map(mentorshipRow).join("")}</div>`;

  container.querySelectorAll("[data-answer]").forEach((button) => {
    button.addEventListener("click", async () => {
      try {
        await api.patch("/mentorships/" + button.dataset.id, { status: button.dataset.answer });
        toast("Request " + button.dataset.answer, "success");
        await openPage("mentorships");
      } catch (error) {
        toast(error.message, "error");
      }
    });
  });
};

function mentorshipRow(item) {
  const amMentor = state.user.id === item.alumni_id;
  const other = amMentor ? item.student_name : item.alumni_name;

  return `
    <div class="list-item">
      <div class="avatar dark">${escapeHtml(initials(other))}</div>
      <div class="list-body">
        <strong>${escapeHtml(other)}</strong>
        <p>${escapeHtml(item.message)}</p>
        <small>${escapeHtml(formatDate(item.created_at))}</small>
      </div>
      <div class="list-actions">
        <span class="badge badge-${escapeHtml(item.status)}">${escapeHtml(item.status)}</span>
        ${
          amMentor && item.status === "pending"
            ? `<button class="btn btn-primary btn-small" data-answer="accepted" data-id="${item.id}">Accept</button>
               <button class="btn btn-danger btn-small" data-answer="rejected" data-id="${item.id}">Decline</button>`
            : ""
        }
      </div>
    </div>`;
}
