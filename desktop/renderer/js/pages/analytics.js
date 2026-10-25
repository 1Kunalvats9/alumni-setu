const analyticsCharts = [
  ["members_by_branch", "Members by branch", "people"],
  ["alumni_by_industry", "Alumni by industry", "alumni"],
  ["alumni_by_batch", "Alumni by graduating batch", "alumni"],
  ["top_mentors", "Most active mentors", "accepted mentees"],
  ["sessions_by_kind", "Sessions by type", "sessions"],
  ["posts_by_category", "Forum activity by topic", "threads"],
  ["openings_by_location", "Openings by location", "openings"],
  ["signups_by_month", "Sign-ups by month", "accounts"],
];

pages.analytics = async function (container) {
  const data = await api.get("/admin/analytics");

  container.innerHTML = `
    <p class="muted page-intro">
      Engagement across the platform. Each chart counts one thing, so every bar
      uses the same colour and length alone carries the value.
    </p>
    <div class="chart-grid">
      ${analyticsCharts.map(([key, title, unit]) => barChart(title, data[key], unit)).join("")}
    </div>`;

  applyBarWidths(container);
};

function applyBarWidths(container) {
  container.querySelectorAll(".bar-fill").forEach((fill) => {
    fill.style.width = fill.dataset.width + "%";
  });
}

function barChart(title, buckets, unit) {
  if (!buckets || buckets.length === 0) {
    return `<div class="card chart"><h3>${escapeHtml(title)}</h3>${emptyState("No data yet.")}</div>`;
  }

  const max = Math.max(...buckets.map((b) => b.value));
  const total = buckets.reduce((sum, b) => sum + b.value, 0);

  const rows = buckets
    .map((bucket) => {
      const percent = max === 0 ? 0 : (bucket.value / max) * 100;
      const label = escapeHtml(bucket.label) + ": " + bucket.value + " " + escapeHtml(unit);

      return `
        <div class="bar-row" title="${label}">
          <span class="bar-label">${escapeHtml(bucket.label)}</span>
          <span class="bar-track">
            <span class="bar-fill" data-width="${percent.toFixed(1)}"></span>
          </span>
          <span class="bar-value">${escapeHtml(bucket.value)}</span>
        </div>`;
    })
    .join("");

  return `
    <div class="card chart">
      <h3>${escapeHtml(title)}</h3>
      <p class="muted chart-total">${escapeHtml(total)} total across ${buckets.length} group${buckets.length === 1 ? "" : "s"}</p>
      <div class="bars">${rows}</div>
    </div>`;
}
