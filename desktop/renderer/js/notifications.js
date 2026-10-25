function setupNotifications() {
  document.getElementById("bell").addEventListener("click", toggleNotifications);
  document.getElementById("notif-read-all").addEventListener("click", async () => {
    await api.post("/notifications/read");
    await refreshNotifications(true);
  });
}

function toggleNotifications() {
  const panel = document.getElementById("notif-panel");
  const opening = panel.classList.contains("hidden");
  panel.classList.toggle("hidden");
  if (opening) refreshNotifications(true);
}

async function refreshNotifications(renderList) {
  let data;
  try {
    data = await api.get("/notifications");
  } catch (error) {
    return;
  }

  const badge = document.getElementById("bell-count");
  badge.textContent = data.unseen > 9 ? "9+" : data.unseen;
  badge.classList.toggle("hidden", data.unseen === 0);

  if (!renderList) return;

  const list = document.getElementById("notif-list");
  list.innerHTML =
    data.count === 0
      ? emptyState("Nothing yet.")
      : data.results.map(notificationRow).join("");

  list.querySelectorAll("[data-notif-page]").forEach((row) => {
    row.addEventListener("click", (event) => {
      if (event.target.hasAttribute("data-notif-remove")) return;
      toggleNotifications();
      openPage(row.dataset.notifPage);
    });
  });

  list.querySelectorAll("[data-notif-remove]").forEach((button) => {
    button.addEventListener("click", async () => {
      await api.delete("/notifications/" + button.dataset.notifRemove);
      await refreshNotifications(true);
    });
  });
}

function notificationRow(item) {
  return `
    <div class="notif ${item.seen ? "" : "unseen"}" data-notif-page="${escapeHtml(item.page || "home")}">
      <div class="notif-body">
        <strong>${escapeHtml(item.title)}</strong>
        ${item.body ? `<p>${escapeHtml(item.body)}</p>` : ""}
        <small>${escapeHtml(formatDateTime(item.created_at))}</small>
      </div>
      <button class="icon-btn small" data-notif-remove="${item.id}" title="Remove">&times;</button>
    </div>`;
}
