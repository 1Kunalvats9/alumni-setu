const { contextBridge } = require("electron");

contextBridge.exposeInMainWorld("alumniSetu", {
  apiBase: "http://localhost:8080/api",
});
