const directoryFilters = {
  q: "",
  role: "alumni",
  branch: "",
  year: "",
  mentor: false,
};

pages.directory = async function (container) {
  const options = await api.get("/directory/filters");

  container.innerHTML = `
    <div class="card">
      <div class="toolbar">
        <input id="dir-q" type="search" placeholder="Search name, company, skill, branch..." value="${escapeHtml(directoryFilters.q)}" />
        <select id="dir-role">
          <option value="">Everyone</option>
          <option value="alumni">Alumni</option>
          <option value="student">Students</option>
        </select>
        <select id="dir-branch">${selectOptions("All branches", options.branches, directoryFilters.branch)}</select>
        <select id="dir-year">${selectOptions("All batches", options.years, directoryFilters.year)}</select>
        <label class="checkbox">
          <input id="dir-mentor" type="checkbox" ${directoryFilters.mentor ? "checked" : ""} />
          Open to mentoring
        </label>
      </div>
    </div>
    <div id="dir-results">${loading()}</div>`;

  document.getElementById("dir-role").value = directoryFilters.role;

  const run = () => {
    directoryFilters.q = document.getElementById("dir-q").value;
    directoryFilters.role = document.getElementById("dir-role").value;
    directoryFilters.branch = document.getElementById("dir-branch").value;
    directoryFilters.year = document.getElementById("dir-year").value;
    directoryFilters.mentor = document.getElementById("dir-mentor").checked;
    loadDirectoryResults();
  };

  document.getElementById("dir-q").addEventListener("input", debounce(run, 300));
  ["dir-role", "dir-branch", "dir-year", "dir-mentor"].forEach((id) => {
    document.getElementById(id).addEventListener("change", run);
  });

  await loadDirectoryResults();
};

async function loadDirectoryResults() {
  const box = document.getElementById("dir-results");
  const params = new URLSearchParams();

  if (directoryFilters.q) params.set("q", directoryFilters.q);
  if (directoryFilters.role) params.set("role", directoryFilters.role);
  if (directoryFilters.branch) params.set("branch", directoryFilters.branch);
  if (directoryFilters.year) params.set("year", directoryFilters.year);
  if (directoryFilters.mentor) params.set("mentor", "true");

  try {
    const data = await api.get("/directory?" + params.toString());
    box.innerHTML =
      data.count === 0
        ? emptyState("No one matches these filters.")
        : `<p class="result-count">${data.count} result${data.count === 1 ? "" : "s"}</p>
           <div class="grid">${data.results.map(personCard).join("")}</div>`;
    wirePersonCards(box);
  } catch (error) {
    box.innerHTML = emptyState(error.message);
  }
}

function personCard(person) {
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
      </div>
      <p class="person-role">${escapeHtml(line || "Profile not completed yet")}</p>
      <p class="person-sub">${escapeHtml([p.branch, p.graduation_year || ""].filter(Boolean).join(" &middot; "))}</p>
      <div class="tags">${tagList(p.skills)}</div>
      <button class="btn btn-secondary btn-small" data-person="${person.id}">View profile</button>
    </div>`;
}

function wirePersonCards(scope) {
  scope.querySelectorAll("[data-person]").forEach((button) => {
    button.addEventListener("click", () => openPersonModal(button.dataset.person));
  });
}

async function openPersonModal(id) {
  const person = await api.get("/users/" + id);
  const p = person.profile;

  const rows = [
    ["Role", [p.job_title, p.company].filter(Boolean).join(" at ")],
    ["Industry", p.industry],
    ["College", p.college],
    ["Branch", p.branch],
    ["Batch", p.graduation_year || ""],
    ["Location", p.location],
    ["LinkedIn", p.linkedin],
  ]
    .filter(([, value]) => value)
    .map(([key, value]) => `<div class="row"><span>${key}</span><strong>${escapeHtml(value)}</strong></div>`)
    .join("");

  const canRequest = state.user.role === "student" && person.role === "alumni";

  showModal(
    person.name,
    `<div class="person-detail">
       <div class="person-meta">${roleBadge(person.role)}${person.verified ? '<span class="badge badge-verified">verified</span>' : ""}</div>
       ${p.bio ? `<p class="bio">${escapeHtml(p.bio)}</p>` : ""}
       ${rows}
       ${p.skills ? `<div class="row stacked"><span>Skills</span><div class="tags">${tagList(p.skills)}</div></div>` : ""}
       ${p.interests ? `<div class="row stacked"><span>Interests</span><div class="tags">${tagList(p.interests)}</div></div>` : ""}
       ${
         canRequest
           ? `<form id="mentor-form" class="form request-form">
                <label>Message to mentor
                  <textarea name="message" placeholder="Introduce yourself and say what guidance you need." required></textarea>
                </label>
                <button type="submit" class="btn btn-primary">Send mentorship request</button>
              </form>`
           : ""
       }
     </div>`
  );

  const form = document.getElementById("mentor-form");
  if (!form) return;

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const message = new FormData(form).get("message");

    try {
      await api.post("/mentorships", { alumni_id: Number(id), message });
      closeModal();
      toast("Mentorship request sent to " + person.name, "success");
    } catch (error) {
      toast(error.message, "error");
    }
  });
}

function selectOptions(allLabel, values, selected) {
  const options = values
    .map((value) => `<option value="${escapeHtml(value)}" ${value === selected ? "selected" : ""}>${escapeHtml(value)}</option>`)
    .join("");
  return `<option value="">${escapeHtml(allLabel)}</option>` + options;
}

function debounce(fn, wait) {
  let timer;
  return function () {
    clearTimeout(timer);
    timer = setTimeout(fn, wait);
  };
}
