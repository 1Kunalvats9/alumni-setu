const assistantStarters = [
  "How do I find a mentor?",
  "When is the next session?",
  "What jobs are available?",
  "How is the match score calculated?",
];

function setupAssistant() {
  document.getElementById("assistant-toggle").addEventListener("click", toggleAssistant);
  document.getElementById("assistant-close").addEventListener("click", toggleAssistant);

  document.getElementById("assistant-form").addEventListener("submit", (event) => {
    event.preventDefault();
    const input = document.getElementById("assistant-input");
    const question = input.value.trim();
    if (!question) return;
    input.value = "";
    askAssistant(question);
  });
}

function toggleAssistant() {
  const panel = document.getElementById("assistant");
  const opening = panel.classList.contains("hidden");
  panel.classList.toggle("hidden");

  if (opening && document.getElementById("assistant-log").children.length === 0) {
    addAssistantMessage(
      "bot",
      "Hello " + state.user.name.split(" ")[0] + ". Ask me how to use the platform.",
      assistantStarters
    );
  }
}

function addAssistantMessage(who, text, suggestions, page) {
  const log = document.getElementById("assistant-log");
  const bubble = document.createElement("div");
  bubble.className = "bubble " + who;

  bubble.innerHTML = `
    <p>${escapeHtml(text)}</p>
    ${page ? `<button class="btn btn-secondary btn-small" data-goto="${escapeHtml(page)}">Take me there</button>` : ""}
    ${
      suggestions && suggestions.length
        ? `<div class="chips">${suggestions
            .map((s) => `<button class="chip" data-ask="${escapeHtml(s)}">${escapeHtml(s)}</button>`)
            .join("")}</div>`
        : ""
    }`;

  log.appendChild(bubble);

  bubble.querySelectorAll("[data-ask]").forEach((button) => {
    button.addEventListener("click", () => askAssistant(button.dataset.ask));
  });

  const goButton = bubble.querySelector("[data-goto]");
  if (goButton) {
    goButton.addEventListener("click", () => {
      toggleAssistant();
      openPage(goButton.dataset.goto);
    });
  }

  log.scrollTop = log.scrollHeight;
}

async function askAssistant(question) {
  addAssistantMessage("user", question);

  try {
    const reply = await api.post("/assistant", { question });
    addAssistantMessage("bot", reply.answer, reply.suggestions, reply.page);
  } catch (error) {
    addAssistantMessage("bot", error.message);
  }
}
