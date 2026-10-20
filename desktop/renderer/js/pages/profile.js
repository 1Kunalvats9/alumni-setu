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

function field(label, name, value, type) {
  return `
    <label>${escapeHtml(label)}
      <input type="${type || "text"}" name="${name}" value="${escapeHtml(value)}" />
    </label>`;
}
