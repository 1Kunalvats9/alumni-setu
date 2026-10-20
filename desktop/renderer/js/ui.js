function escapeHtml(value) {
  return String(value === null || value === undefined ? "" : value)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function formatDate(value) {
  if (!value) return "";
  const date = new Date(value);
  if (isNaN(date)) return value;
  return date.toLocaleDateString("en-IN", {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
}

function formatDateTime(value) {
  if (!value) return "";
  const date = new Date(value);
  if (isNaN(date)) return value;
  return date.toLocaleString("en-IN", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

function initials(name) {
  return String(name || "?")
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0].toUpperCase())
    .join("");
}

function toast(message, kind) {
  const box = document.getElementById("toast");
  box.textContent = message;
  box.className = "toast show " + (kind || "info");
  setTimeout(() => {
    box.className = "toast";
  }, 3200);
}

function tagList(csv) {
  const items = String(csv || "")
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);

  if (items.length === 0) return "";
  return items.map((item) => `<span class="tag">${escapeHtml(item)}</span>`).join("");
}

function roleBadge(role) {
  return `<span class="badge badge-${escapeHtml(role)}">${escapeHtml(role)}</span>`;
}

function emptyState(message) {
  return `<div class="empty">${escapeHtml(message)}</div>`;
}

function loading() {
  return `<div class="empty">Loading...</div>`;
}
