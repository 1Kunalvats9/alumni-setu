const { app, BrowserWindow, dialog } = require("electron");
const { spawn } = require("child_process");
const path = require("path");
const fs = require("fs");
const http = require("http");

const API_PORT = 8080;
const API_HEALTH = `http://localhost:${API_PORT}/api/health`;

let serverProcess = null;
let mainWindow = null;

function serverBinaryPath() {
  if (app.isPackaged) {
    return path.join(process.resourcesPath, "server", "alumnisetu-server");
  }
  return path.join(__dirname, "..", "server", "bin", "alumnisetu-server");
}

function pingServer() {
  return new Promise((resolve) => {
    const request = http.get(API_HEALTH, (response) => {
      response.resume();
      resolve(response.statusCode === 200);
    });
    request.on("error", () => resolve(false));
    request.setTimeout(700, () => {
      request.destroy();
      resolve(false);
    });
  });
}

async function waitForServer(attempts) {
  for (let i = 0; i < attempts; i++) {
    if (await pingServer()) return true;
    await new Promise((resolve) => setTimeout(resolve, 300));
  }
  return false;
}

function startServer() {
  const binary = serverBinaryPath();

  if (!fs.existsSync(binary)) {
    console.log("no bundled server at", binary, "- expecting one to be running already");
    return false;
  }

  const databasePath = path.join(app.getPath("userData"), "alumnisetu.db");
  console.log("starting server", binary, "db", databasePath);

  serverProcess = spawn(binary, [], {
    env: { ...process.env, ALUMNISETU_DB: databasePath },
    stdio: ["ignore", "pipe", "pipe"],
  });

  serverProcess.stdout.on("data", (chunk) => console.log("[server]", String(chunk).trim()));
  serverProcess.stderr.on("data", (chunk) => console.error("[server]", String(chunk).trim()));
  serverProcess.on("exit", (code) => console.log("[server] exited with", code));

  return true;
}

function stopServer() {
  if (serverProcess && !serverProcess.killed) {
    serverProcess.kill();
    serverProcess = null;
  }
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 860,
    minWidth: 900,
    minHeight: 600,
    title: "AlumniSetu",
    backgroundColor: "#f4f6fb",
    webPreferences: {
      preload: path.join(__dirname, "preload.js"),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  mainWindow.loadFile(path.join(__dirname, "renderer", "index.html"));
}

app.whenReady().then(async () => {
  const alreadyRunning = await pingServer();
  if (!alreadyRunning) {
    startServer();
  }

  const ready = await waitForServer(20);
  if (!ready) {
    dialog.showErrorBox(
      "AlumniSetu cannot reach its server",
      `Nothing is answering on port ${API_PORT}.\n\n` +
        "If you are running from source, start it with:\n" +
        "    cd server && go run ."
    );
  }

  createWindow();

  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") app.quit();
});

app.on("before-quit", stopServer);
app.on("will-quit", stopServer);
process.on("exit", stopServer);
