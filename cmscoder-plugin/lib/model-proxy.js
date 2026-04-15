"use strict";

const http = require("http");
const https = require("https");
const fs = require("fs");
const path = require("path");
const os = require("os");
const crypto = require("crypto");
const { getModelToken } = require("./model-token");
const { _getBaseUrl } = require("./http-client");

// ── Configuration ─────────────────────────────────────────────

const PROXY_CONFIG_DIR = path.join(os.homedir(), ".cmscoder");
const PROXY_PID_FILE = path.join(PROXY_CONFIG_DIR, "model-proxy.pid");
const PROXY_SECRET_FILE = path.join(PROXY_CONFIG_DIR, "model-proxy.secret");

// Unix Domain Socket path (macOS/Linux)
const UDS_PATH = path.join(PROXY_CONFIG_DIR, "model-proxy.sock");

// Windows Named Pipe name
const NAMED_PIPE_NAME = `\\\\.\\pipe\\cmscoder-model-proxy-${process.getuid?.() || process.pid}`;

// ── State ─────────────────────────────────────────────────────

let cachedModelToken = null;
let tokenExpiresAt = 0;
let proxyServer = null;
let proxySecret = null;

// ── Platform Detection ────────────────────────────────────────

function isWindows() {
  return process.platform === "win32";
}

function isMacOrLinux() {
  return process.platform === "darwin" || process.platform === "linux";
}

// ── Secret Management ─────────────────────────────────────────

function generateProxySecret() {
  return crypto.randomBytes(32).toString("hex");
}

function saveProxySecret(secret) {
  fs.mkdirSync(PROXY_CONFIG_DIR, { recursive: true, mode: 0o700 });
  fs.writeFileSync(PROXY_SECRET_FILE, secret, { mode: 0o600 });
}

function loadProxySecret() {
  if (fs.existsSync(PROXY_SECRET_FILE)) {
    return fs.readFileSync(PROXY_SECRET_FILE, "utf8").trim();
  }
  return null;
}

function clearProxySecret() {
  if (fs.existsSync(PROXY_SECRET_FILE)) {
    fs.unlinkSync(PROXY_SECRET_FILE);
  }
}

// ── PID File Management ───────────────────────────────────────

function savePid(pid) {
  fs.mkdirSync(PROXY_CONFIG_DIR, { recursive: true, mode: 0o700 });
  fs.writeFileSync(PROXY_PID_FILE, String(pid), { mode: 0o600 });
}

function loadPid() {
  if (fs.existsSync(PROXY_PID_FILE)) {
    const pid = fs.readFileSync(PROXY_PID_FILE, "utf8").trim();
    return parseInt(pid, 10);
  }
  return null;
}

function clearPid() {
  if (fs.existsSync(PROXY_PID_FILE)) {
    fs.unlinkSync(PROXY_PID_FILE);
  }
}

function isProcessRunning(pid) {
  try {
    process.kill(pid, 0);
    return true;
  } catch {
    return false;
  }
}

// ── Model Token Cache ─────────────────────────────────────────

async function getCachedModelToken() {
  // Refresh token if it expires in less than 60 seconds
  const now = Date.now();
  const bufferMs = 60 * 1000;

  if (!cachedModelToken || tokenExpiresAt - bufferMs < now) {
    try {
      cachedModelToken = await getModelToken();
      // Token TTL is 5 minutes (300 seconds)
      tokenExpiresAt = now + 5 * 60 * 1000;
    } catch (error) {
      console.error("Failed to get Model Token:", error.message);
      throw error;
    }
  }

  return cachedModelToken;
}

// ── HTTP Request Forwarding ───────────────────────────────────

function forwardRequest(req, res, modelToken) {
  const baseUrl = _getBaseUrl();
  const parsed = new URL(req.url, baseUrl);
  const isHttps = parsed.protocol === "https:";
  const lib = isHttps ? https : http;

  const options = {
    hostname: parsed.hostname,
    port: parsed.port || (isHttps ? 443 : 80),
    path: parsed.pathname + parsed.search,
    method: req.method,
    headers: {
      ...req.headers,
      "Authorization": `Bearer ${modelToken}`,
      "Host": parsed.hostname,
    },
  };

  // Remove hop-by-hop headers
  delete options.headers["proxy-connection"];
  delete options.headers["transfer-encoding"];

  const proxyReq = lib.request(options, (proxyRes) => {
    res.writeHead(proxyRes.statusCode, proxyRes.headers);
    proxyRes.pipe(res);
  });

  proxyReq.on("error", (err) => {
    console.error("Proxy request error:", err.message);
    res.writeHead(502);
    res.end(JSON.stringify({ error: "Bad Gateway", message: err.message }));
  });

  req.pipe(proxyReq);
}

// ── Request Handler ───────────────────────────────────────────

async function handleRequest(req, res) {
  // 1. Validate proxy secret from header
  const providedSecret = req.headers["x-proxy-secret"];
  if (!providedSecret || providedSecret !== proxySecret) {
    res.writeHead(401);
    res.end(JSON.stringify({ error: "Unauthorized", message: "Invalid or missing proxy secret" }));
    return;
  }

  // 2. Validate User-Agent contains OpenCode
  const userAgent = req.headers["user-agent"] || "";
  if (!userAgent.includes("OpenCode")) {
    // Log warning but don't block (for compatibility)
    console.error("Warning: Request User-Agent does not contain 'OpenCode':", userAgent);
  }

  // 3. Get Model Token (with caching)
  let modelToken;
  try {
    modelToken = await getCachedModelToken();
  } catch (error) {
    res.writeHead(401);
    res.end(JSON.stringify({ error: "Unauthorized", message: "Failed to get Model Token. Please login." }));
    return;
  }

  // 4. Forward request to upstream with Model Token
  forwardRequest(req, res, modelToken);
}

// ── Server Creation ───────────────────────────────────────────

function createServer() {
  return http.createServer((req, res) => {
    // Set CORS headers
    res.setHeader("Access-Control-Allow-Origin", "*");
    res.setHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS");
    res.setHeader("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Proxy-Secret");

    if (req.method === "OPTIONS") {
      res.writeHead(200);
      res.end();
      return;
    }

    handleRequest(req, res);
  });
}

// ── Server Start/Stop ─────────────────────────────────────────

function startServer() {
  return new Promise((resolve, reject) => {
    // Generate or load proxy secret
    proxySecret = loadProxySecret();
    if (!proxySecret) {
      proxySecret = generateProxySecret();
      saveProxySecret(proxySecret);
    }

    proxyServer = createServer();

    if (isMacOrLinux()) {
      // Use Unix Domain Socket
      // Clean up old socket file if exists
      if (fs.existsSync(UDS_PATH)) {
        fs.unlinkSync(UDS_PATH);
      }

      proxyServer.listen(UDS_PATH, () => {
        // Set socket permissions to 600 (owner only)
        fs.chmodSync(UDS_PATH, 0o600);
        console.error(`Model Proxy listening on Unix Domain Socket: ${UDS_PATH}`);
        savePid(process.pid);
        resolve({ type: "uds", path: UDS_PATH, secret: proxySecret });
      });
    } else {
      // Windows: Use TCP on localhost with dynamic port
      proxyServer.listen(0, "127.0.0.1", () => {
        const port = proxyServer.address().port;
        console.error(`Model Proxy listening on 127.0.0.1:${port}`);
        savePid(process.pid);
        resolve({ type: "tcp", port, secret: proxySecret });
      });
    }

    proxyServer.on("error", (err) => {
      reject(err);
    });
  });
}

function stopServer() {
  return new Promise((resolve) => {
    if (proxyServer) {
      proxyServer.close(() => {
        console.error("Model Proxy stopped");
        resolve();
      });
      proxyServer = null;
    } else {
      resolve();
    }

    // Clean up
    clearPid();
    clearProxySecret();

    // Clean up socket file on Unix
    if (isMacOrLinux() && fs.existsSync(UDS_PATH)) {
      fs.unlinkSync(UDS_PATH);
    }
  });
}

// ── OpenCode Config ───────────────────────────────────────────

function getOpenCodeConfig(type, address, secret) {
  const baseURL = type === "uds"
    ? `http+unix://${encodeURIComponent(address)}`
    : `http://127.0.0.1:${address}`;

  return {
    provider: {
      "cmscoder-local": {
        baseURL,
        apiKey: secret,
        models: {
          "gpt-4": { id: "gpt-4" },
          "gpt-4o": { id: "gpt-4o" },
        },
      },
    },
    model: "cmscoder-local/gpt-4",
  };
}

function writeOpenCodeConfig(config) {
  const opencodeDir = path.join(os.homedir(), ".opencode");
  fs.mkdirSync(opencodeDir, { recursive: true });

  const configPath = path.join(opencodeDir, "config.json");

  // Read existing config if present
  let existingConfig = {};
  if (fs.existsSync(configPath)) {
    try {
      existingConfig = JSON.parse(fs.readFileSync(configPath, "utf8"));
    } catch {
      // Ignore parse errors
    }
  }

  // Merge with new provider config
  const newConfig = {
    ...existingConfig,
    ...config,
    provider: {
      ...(existingConfig.provider || {}),
      ...config.provider,
    },
  };

  fs.writeFileSync(configPath, JSON.stringify(newConfig, null, 2));
  console.error(`OpenCode config written to: ${configPath}`);
}

// ── CLI Commands ──────────────────────────────────────────────

async function start() {
  // Check if already running
  const existingPid = loadPid();
  if (existingPid && isProcessRunning(existingPid)) {
    console.error(`Model Proxy is already running (PID: ${existingPid})`);
    console.error("Use 'model-proxy stop' to stop it first.");
    process.exit(1);
  }

  console.error("Starting Model Proxy...");

  const { type, path: address, port, secret } = await startServer();

  // Write OpenCode config
  const opencodeConfig = getOpenCodeConfig(type, address || port, secret);
  writeOpenCodeConfig(opencodeConfig);

  console.error("Model Proxy started successfully");
  console.error(`Provider: cmscoder-local`);
  console.error(`Secret: ${secret.substring(0, 8)}...`);

  // Keep process running
  process.on("SIGINT", async () => {
    console.error("\nReceived SIGINT, stopping proxy...");
    await stopServer();
    process.exit(0);
  });

  process.on("SIGTERM", async () => {
    console.error("\nReceived SIGTERM, stopping proxy...");
    await stopServer();
    process.exit(0);
  });
}

async function stop() {
  const pid = loadPid();
  if (!pid) {
    console.error("Model Proxy is not running (no PID file found)");
    process.exit(1);
  }

  if (!isProcessRunning(pid)) {
    console.error("Model Proxy is not running (stale PID file)");
    clearPid();
    clearProxySecret();
    process.exit(1);
  }

  console.error(`Stopping Model Proxy (PID: ${pid})...`);

  try {
    process.kill(pid, "SIGTERM");
    console.error("Model Proxy stopped");
  } catch (err) {
    console.error(`Failed to stop proxy: ${err.message}`);
    process.exit(1);
  }
}

async function status() {
  const pid = loadPid();
  const secret = loadProxySecret();

  if (!pid) {
    console.log("Status: Not running");
    return;
  }

  if (isProcessRunning(pid)) {
    console.log("Status: Running");
    console.log(`PID: ${pid}`);
    if (isMacOrLinux()) {
      console.log(`Socket: ${UDS_PATH}`);
    }
    if (secret) {
      console.log(`Secret: ${secret.substring(0, 8)}...`);
    }
  } else {
    console.log("Status: Not running (stale PID file)");
    clearPid();
    clearProxySecret();
  }
}

// ── CLI Entry Point ───────────────────────────────────────────

async function main() {
  const subcommand = process.argv[3] || "start";

  switch (subcommand) {
    case "start":
      await start();
      break;
    case "stop":
      await stop();
      break;
    case "status":
      await status();
      break;
    default:
      console.error(`Usage: cmscoder model-proxy [start|stop|status]`);
      process.exit(1);
  }
}

module.exports = {
  start,
  stop,
  status,
};

if (require.main === module) {
  main().catch((e) => {
    console.error(`ERROR: ${e.message}`);
    process.exit(1);
  });
}
