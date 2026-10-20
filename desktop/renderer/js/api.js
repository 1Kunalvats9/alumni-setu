const API_BASE = window.alumniSetu.apiBase;
const TOKEN_KEY = "alumnisetu.token";

const api = {
  token: localStorage.getItem(TOKEN_KEY) || "",

  setToken(token) {
    this.token = token;
    localStorage.setItem(TOKEN_KEY, token);
  },

  clearToken() {
    this.token = "";
    localStorage.removeItem(TOKEN_KEY);
  },

  async request(method, path, body) {
    const options = { method, headers: {} };

    if (this.token) {
      options.headers["Authorization"] = "Bearer " + this.token;
    }
    if (body !== undefined) {
      options.headers["Content-Type"] = "application/json";
      options.body = JSON.stringify(body);
    }

    let response;
    try {
      response = await fetch(API_BASE + path, options);
    } catch (error) {
      throw new Error("Cannot reach the server. Is it running on port 8080?");
    }

    const text = await response.text();
    const data = text ? JSON.parse(text) : {};

    if (!response.ok) {
      throw new Error(data.error || "Request failed");
    }
    return data;
  },

  get(path) {
    return this.request("GET", path);
  },
  post(path, body) {
    return this.request("POST", path, body);
  },
  put(path, body) {
    return this.request("PUT", path, body);
  },
  patch(path, body) {
    return this.request("PATCH", path, body);
  },
  delete(path) {
    return this.request("DELETE", path);
  },
};
