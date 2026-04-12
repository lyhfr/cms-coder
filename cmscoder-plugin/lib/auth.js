"use strict";

const { execSync } = require("child_process");
const os = require("os");
const { secureStore, localCache } = require("./storage");
const { httpClient, _getBaseUrl, _getPluginDir } = require("./http-client");
const { startCallbackServer } = require("./callback-server");
const { sync: bootstrapSync } = require("./bootstrap");

const AGENT_TYPE = process.env.CMSCODER_AGENT_TYPE || "claude-code";
const PLUGIN_INSTANCE_ID = process.env.CMSCODER_PLUGIN_INSTANCE_ID || "cmscoder-default";

// ── Browser helpers ───────────────────────────────────────────

function _openBrowser(url) {
  const platform = os.platform();
  const cmd = platform === "darwin" ? "open" : platform === "win32" ? "start" : "xdg-open";
  try {
    execSync(`${cmd} "${url}"`, { stdio: "ignore" });
  } catch {
    console.error(`Please open this URL in your browser: ${url}`);
  }
}

// ── Login ─────────────────────────────────────────────────────

async function login() {
  // Validate backend URL.
  try {
    _getBaseUrl();
  } catch (e) {
    console.error(`ERROR: ${e.message}`);
    process.exit(1);
  }

  console.error("Starting cmscoder login...");

  // 1. Start callback server.
  const { port, waitForTicket, close } = startCallbackServer();
  const callbackPort = await port;
  console.error(`Callback server started on port ${callbackPort}`);

  try {
    // 2. Create login session on backend.
    const loginData = await httpClient.post("/api/auth/login", {
      localPort: callbackPort,
      agentType: AGENT_TYPE,
      pluginInstanceId: PLUGIN_INSTANCE_ID,
    });

    console.error("Login session created");

    const browserUrl = loginData.browserUrl || loginData.data?.browserUrl;
    if (!browserUrl) {
      throw new Error(`Failed to get browser URL from response: ${JSON.stringify(loginData)}`);
    }

    // 3. Open browser.
    console.error("Opening browser for authentication...");
    _openBrowser(browserUrl);

    // 4. Wait for callback.
    console.error("Waiting for browser callback (timeout: 5 min)...");
    const loginTicket = await waitForTicket;
    console.error("Received login ticket, exchanging for session...");

    // 5. Exchange ticket for formal session.
    const exchangeData = await httpClient.post("/api/auth/exchange", {
      loginTicket,
      pluginInstanceId: PLUGIN_INSTANCE_ID,
    });

    const accessToken = exchangeData.accessToken || exchangeData.data?.accessToken;
    const refreshToken = exchangeData.refreshToken || exchangeData.data?.refreshToken;
    const modelApiKey = exchangeData.modelApiKey || exchangeData.data?.modelApiKey;
    const expiresIn = exchangeData.expiresIn || exchangeData.data?.expiresIn || 900;
    const user = exchangeData.user || exchangeData.data?.user || {};

    if (!accessToken || !refreshToken) {
      throw new Error(`Failed to extract tokens from response: ${JSON.stringify(exchangeData)}`);
    }

    // 6. Store tokens.
    secureStore.set("access_token", accessToken);
    secureStore.set("refresh_token", refreshToken);
    secureStore.set("user_info", JSON.stringify(user));
    if (modelApiKey) {
      secureStore.set("model_api_key", modelApiKey);
    }

    // 7. Store session metadata.
    const expiresAt = new Date(Date.now() + expiresIn * 1000).toISOString();
    localCache.setSessionMeta({
      userId: user.userId || "unknown",
      email: user.email || "",
      displayName: user.displayName || "",
      tenantId: user.tenantId || "",
      sessionId: user.sessionId || "",
      expiresAt,
    });

    // 8. Bootstrap sync.
    try {
      await bootstrapSync(accessToken);
    } catch (e) {
      console.error(`Warning: Bootstrap sync failed, will retry on next startup: ${e.message}`);
    }

    console.error(`Login successful! Welcome, ${user.displayName || user.userId}`);
    if (modelApiKey) {
      console.error(`Model API key generated: ${modelApiKey.substring(0, 10)}...`);
    }
  } finally {
    close();
  }
}

// ── Logout ────────────────────────────────────────────────────

async function logout() {
  console.error("Logging out...");
  const refreshToken = secureStore.get("refresh_token");

  if (refreshToken) {
    try {
      await httpClient.post("/api/auth/logout", { refreshToken });
    } catch {
      // Ignore server errors during logout.
    }
  }

  secureStore.clearAll();
  localCache.clear();
  console.error("Logged out successfully");
}

// ── Silent Refresh ────────────────────────────────────────────

async function refreshSilent() {
  const refreshToken = secureStore.get("refresh_token");
  if (!refreshToken) return false;

  try {
    const data = await httpClient.post("/api/auth/refresh", { refreshToken });
    const accessToken = data.accessToken || data.data?.accessToken;
    const newRefreshToken = data.refreshToken || data.data?.refreshToken;
    const expiresIn = data.expiresIn || data.data?.expiresIn || 900;

    if (!accessToken) return false;

    secureStore.set("access_token", accessToken);
    if (newRefreshToken) secureStore.set("refresh_token", newRefreshToken);

    const expiresAt = new Date(Date.now() + expiresIn * 1000).toISOString();
    const userInfo = JSON.parse(secureStore.get("user_info") || "{}");
    localCache.setSessionMeta({
      userId: userInfo.userId || "unknown",
      email: userInfo.email || "",
      displayName: userInfo.displayName || "",
      tenantId: userInfo.tenantId || "",
      sessionId: userInfo.sessionId || "",
      expiresAt,
    });

    return true;
  } catch {
    return false;
  }
}

// ── Ensure Session ────────────────────────────────────────────

async function ensureSession() {
  const accessToken = secureStore.get("access_token");
  if (!accessToken) return false;

  if (localCache.sessionValid()) return true;

  // Try silent refresh.
  const refreshed = await refreshSilent();
  if (refreshed) return true;

  console.error("Session expired. Please login again: run /cmscoder-login in Claude Code");
  return false;
}

// ── Get Access Token ──────────────────────────────────────────

async function getAccessToken() {
  const token = secureStore.get("access_token");
  if (!token) return null;

  if (!localCache.sessionValid()) {
    const refreshed = await refreshSilent();
    if (!refreshed) return null;
    return secureStore.get("access_token");
  }

  return token;
}

// ── CLI Commands ──────────────────────────────────────────────

async function status() {
  const accessToken = secureStore.get("access_token");
  if (!accessToken) {
    console.log("Not logged in. Run: cmscoder login");
    return;
  }

  const meta = localCache.getSessionMeta();
  const valid = localCache.sessionValid();

  console.log(`Status: ${valid ? "Active" : "Expired"}`);
  if (meta.userId) console.log(`User: ${meta.displayName || meta.userId}`);
  if (meta.email) console.log(`Email: ${meta.email}`);
  if (meta.tenantId) console.log(`Tenant: ${meta.tenantId}`);
  if (meta.expiresAt) console.log(`Expires: ${meta.expiresAt}`);
  if (meta.cachedAt) console.log(`Cached at: ${meta.cachedAt}`);

  try {
    const url = _getBaseUrl();
    console.log(`Backend: ${url}`);
  } catch { /* ignore */ }
}

async function cmdToken() {
  const token = await getAccessToken();
  if (!token) {
    console.error("No valid session. Please login.");
    process.exit(1);
  }
  console.log(token);
}

async function cmdRefresh() {
  console.error("Refreshing session...");
  if (await refreshSilent()) {
    console.error("Session refreshed successfully");
  } else {
    console.error("Session refresh failed. Please login again.");
    process.exit(1);
  }
}

async function cmdEnsureSession() {
  if (await ensureSession()) {
    process.exit(0);
  }
  process.exit(1);
}

module.exports = {
  login,
  logout,
  refreshSilent,
  ensureSession,
  getAccessToken,
  getModelApiKey,
  status,
};

/**
 * Get the stored model API key.
 */
function getModelApiKey() {
  return secureStore.get("model_api_key");
}

// CLI entry point.
if (require.main === module) {
  const command = process.argv[2];
  const commands = {
    login,
    logout,
    refresh: cmdRefresh,
    status,
    token: cmdToken,
    "ensure-session": cmdEnsureSession,
  };

  const fn = commands[command];
  if (!fn) {
    console.error(`Usage: cmscoder.js <command>
Commands:
  login            Full login flow (opens browser)
  logout           Revoke session and clear local data
  refresh          Silent session refresh
  status           Show current session status
  token            Print current access token to stdout
  ensure-session   Check and restore session (for hooks)`);
    process.exit(1);
  }

  fn().catch((e) => {
    console.error(`ERROR: ${e.message}`);
    process.exit(1);
  });
}
