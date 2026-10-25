const state = {
  user: null,
  profile: null,
};

const navItems = [
  { id: "home", label: "Home" },
  { id: "directory", label: "Alumni Directory" },
  { id: "mentors", label: "Suggested Mentors" },
  { id: "mentorships", label: "Mentorship Requests" },
  { id: "events", label: "Events & Webinars" },
  { id: "forum", label: "Discussion Forum" },
  { id: "jobs", label: "Placements" },
  { id: "admin", label: "Admin Console", roles: ["admin"] },
  { id: "analytics", label: "Analytics", roles: ["admin"] },
  { id: "profile", label: "My Profile" },
];

function visibleNavItems() {
  return navItems.filter((item) => !item.roles || item.roles.includes(state.user.role));
}

function showScreen(name) {
  document.getElementById("screen-login").classList.toggle("hidden", name !== "login");
  document.getElementById("screen-app").classList.toggle("hidden", name !== "app");

  if (name === "login") {
    document.getElementById("notif-panel").classList.add("hidden");
    document.getElementById("bell-count").classList.add("hidden");
    document.getElementById("notif-list").innerHTML = "";
    document.getElementById("assistant-toggle").classList.add("hidden");
    document.getElementById("assistant").classList.add("hidden");
    document.getElementById("assistant-log").innerHTML = "";
  }
}

function renderNav() {
  const nav = document.getElementById("nav");
  nav.innerHTML = visibleNavItems()
    .map((item) => `<button class="nav-item" data-page="${item.id}">${item.label}</button>`)
    .join("");

  nav.querySelectorAll(".nav-item").forEach((button) => {
    button.addEventListener("click", () => openPage(button.dataset.page));
  });
}

async function openPage(name) {
  const allowed = visibleNavItems();
  const item = allowed.find((entry) => entry.id === name) || allowed[0];

  document.querySelectorAll(".nav-item").forEach((button) => {
    button.classList.toggle("active", button.dataset.page === item.id);
  });

  document.getElementById("page-title").textContent = item.label;

  const container = document.getElementById("page");
  container.innerHTML = loading();

  try {
    await pages[item.id](container);
  } catch (error) {
    container.innerHTML = emptyState(error.message);
  }

  refreshNotifications(false);
}

function renderCurrentUser() {
  document.getElementById("current-user").innerHTML = `
    <div class="avatar">${escapeHtml(initials(state.user.name))}</div>
    <div>
      <strong>${escapeHtml(state.user.name)}</strong>
      <span>${escapeHtml(state.user.role)}</span>
    </div>`;
}

async function startApp() {
  const me = await api.get("/me");
  state.user = me.user;
  state.profile = me.profile;

  showScreen("app");
  document.getElementById("assistant-toggle").classList.remove("hidden");
  renderNav();
  renderCurrentUser();
  await openPage("home");
  await refreshNotifications(false);
}

function logout() {
  api.post("/logout").catch(() => {});
  api.clearToken();
  state.user = null;
  state.profile = null;
  showScreen("login");
}

async function boot() {
  setupAuthForms();
  setupModal();
  setupAssistant();
  setupNotifications();
  document.getElementById("btn-logout").addEventListener("click", logout);

  if (!api.token) {
    showScreen("login");
    return;
  }

  try {
    await startApp();
  } catch (error) {
    api.clearToken();
    showScreen("login");
    if (error.message.includes("Cannot reach")) {
      showLoginError(error.message);
    }
  }
}

boot();
