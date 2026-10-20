function showLoginError(message) {
  const box = document.getElementById("login-error");
  box.textContent = message;
  box.classList.remove("hidden");
}

function hideLoginError() {
  document.getElementById("login-error").classList.add("hidden");
}

function setupAuthForms() {
  const tabLogin = document.getElementById("tab-login");
  const tabRegister = document.getElementById("tab-register");
  const formLogin = document.getElementById("form-login");
  const formRegister = document.getElementById("form-register");

  tabLogin.addEventListener("click", () => {
    hideLoginError();
    tabLogin.classList.add("active");
    tabRegister.classList.remove("active");
    formLogin.classList.remove("hidden");
    formRegister.classList.add("hidden");
  });

  tabRegister.addEventListener("click", () => {
    hideLoginError();
    tabRegister.classList.add("active");
    tabLogin.classList.remove("active");
    formRegister.classList.remove("hidden");
    formLogin.classList.add("hidden");
  });

  formLogin.addEventListener("submit", async (event) => {
    event.preventDefault();
    hideLoginError();

    const data = Object.fromEntries(new FormData(formLogin));
    try {
      const result = await api.post("/login", data);
      api.setToken(result.token);
      await startApp();
    } catch (error) {
      showLoginError(error.message);
    }
  });

  formRegister.addEventListener("submit", async (event) => {
    event.preventDefault();
    hideLoginError();

    const data = Object.fromEntries(new FormData(formRegister));
    try {
      const result = await api.post("/register", data);
      api.setToken(result.token);
      await startApp();
    } catch (error) {
      showLoginError(error.message);
    }
  });
}
