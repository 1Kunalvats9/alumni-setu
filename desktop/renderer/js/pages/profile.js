pages.profile = async function (container) {
  const me = await api.get("/me");
  state.user = me.user;
  state.profile = me.profile;

  const p = state.profile;
  const isAlumni = state.user.role === "alumni";

  container.innerHTML = `
    <div class="card">
      <h3>Account</h3>
      <div class="row"><span>Name</span><strong>${escapeHtml(state.user.name)}</strong></div>
      <div class="row"><span>Email</span><strong>${escapeHtml(state.user.email)}</strong></div>
      <div class="row"><span>Role</span><strong>${roleBadge(state.user.role)}</strong></div>
      <div class="row"><span>Verification</span><strong>${
        state.user.verified
          ? '<span class="badge badge-verified">verified</span>'
          : '<span class="badge badge-pending">pending department review</span>'
      }</strong></div>
    </div>

    <form id="profile-form" class="card">
      <h3>Profile details</h3>
      <p class="muted">This is what other people see in the directory, and what mentor matching uses.</p>

      <div class="form-grid">
        ${field("College", "college", p.college)}
        ${field("Branch", "branch", p.branch)}
        ${field("Graduation year", "graduation_year", p.graduation_year || "", "number")}
        ${field("Location", "location", p.location)}
        ${field("Company", "company", p.company)}
        ${field("Job title", "job_title", p.job_title)}
        ${field("Industry", "industry", p.industry)}
        ${field("LinkedIn", "linkedin", p.linkedin)}
      </div>

      ${field("Skills (comma separated)", "skills", p.skills)}
      ${field("Interests (comma separated)", "interests", p.interests)}

      <label>About you
        <textarea name="bio">${escapeHtml(p.bio)}</textarea>
      </label>

      <div class="extract">
        <button type="button" id="btn-extract" class="btn btn-secondary btn-small">
          Scan my bio for skills
        </button>
        <span class="muted">Reads the text above and suggests skills to add.</span>
      </div>
      <div id="extract-result"></div>

      ${
        isAlumni
          ? `<label class="checkbox">
               <input type="checkbox" name="open_to_mentor" ${p.open_to_mentor ? "checked" : ""} />
               I am open to mentoring students
             </label>`
          : ""
      }

      <button type="submit" class="btn btn-primary">Save profile</button>
    </form>`;

  document.getElementById("btn-extract").addEventListener("click", runSkillScan);

  document.getElementById("profile-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.target;
    const data = Object.fromEntries(new FormData(form));

    data.graduation_year = Number(data.graduation_year) || 0;
    data.open_to_mentor = form.open_to_mentor ? form.open_to_mentor.checked : false;

    try {
      state.profile = await api.put("/me/profile", data);
      toast("Profile saved", "success");
    } catch (error) {
      toast(error.message, "error");
    }
  });
};

async function runSkillScan() {
  const form = document.getElementById("profile-form");
  const box = document.getElementById("extract-result");
  box.innerHTML = loading();

  try {
    const result = await api.post("/me/extract-skills", { text: form.bio.value });

    if (result.found.length === 0) {
      box.innerHTML = emptyState(
        "No known skills found in your bio. Try naming tools and subjects directly, for example 'I work with Python and SQL'."
      );
      return;
    }

    box.innerHTML = `
      <div class="extract-result">
        <p class="muted">
          Found ${result.found.length} skill${result.found.length === 1 ? "" : "s"}
          ${result.already.length ? "&middot; " + result.already.length + " already on your profile" : ""}
        </p>
        ${
          result.new.length === 0
            ? `<p class="muted">Everything found is already listed.</p>`
            : `<div class="tags">${result.new
                .map((s) => `<button type="button" class="chip" data-add-skill="${escapeHtml(s)}">+ ${escapeHtml(s)}</button>`)
                .join("")}</div>
               <button type="button" id="btn-add-all" class="btn btn-secondary btn-small">Add all</button>`
        }
      </div>`;

    box.querySelectorAll("[data-add-skill]").forEach((button) => {
      button.addEventListener("click", () => {
        addSkillToField(button.dataset.addSkill);
        button.remove();
      });
    });

    const addAll = document.getElementById("btn-add-all");
    if (addAll) {
      addAll.addEventListener("click", () => {
        result.new.forEach(addSkillToField);
        box.innerHTML = emptyState("Added. Remember to save your profile.");
      });
    }
  } catch (error) {
    box.innerHTML = emptyState(error.message);
  }
}

function addSkillToField(skill) {
  const field = document.getElementById("profile-form").skills;
  const current = field.value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);

  if (current.some((item) => item.toLowerCase() === skill.toLowerCase())) return;

  current.push(skill);
  field.value = current.join(", ");
}

function field(label, name, value, type) {
  return `
    <label>${escapeHtml(label)}
      <input type="${type || "text"}" name="${name}" value="${escapeHtml(value)}" />
    </label>`;
}
