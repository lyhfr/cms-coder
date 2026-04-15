"use strict";

const crypto = require("crypto");
const { secureStore, localCache } = require("./storage");
const { httpClient, _getBaseUrl } = require("./http-client");

const AGENT_TYPE = process.env.CMSCODER_AGENT_TYPE || "claude-code";
const PLUGIN_INSTANCE_ID = process.env.CMSCODER_PLUGIN_INSTANCE_ID || "cmscoder-default";

// ── HMAC Signature ────────────────────────────────────────────

/**
 * Generates HMAC-SHA256 signature for Model Token request.
 * @param {string} accessToken
 * @param {number} timestamp
 * @param {string} nonce
 * @param {string} pluginSecret
 * @returns {string} Hex-encoded signature
 */
function generateSignature(accessToken, timestamp, nonce, pluginSecret) {
  const message = accessToken + timestamp + nonce;
  return crypto.createHmac("sha256", pluginSecret).update(message).digest("hex");
}

/**
 * Generates a random nonce string.
 * @param {number} length
 * @returns {string}
 */
function generateNonce(length = 16) {
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  let result = "";
  const randomBytes = crypto.randomBytes(length);
  for (let i = 0; i < length; i++) {
    result += chars[randomBytes[i] % chars.length];
  }
  return result;
}

// ── Token Refresh ─────────────────────────────────────────────

/**
 * Attempts to silently refresh the access token.
 * @returns {Promise<boolean>}
 */
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

// ── Model Token ───────────────────────────────────────────────

/**
 * Gets a valid access token, refreshing if necessary.
 * @returns {Promise<string|null>}
 */
async function getValidAccessToken() {
  let accessToken = secureStore.get("access_token");
  if (!accessToken) return null;

  // Check if token is still valid (with 60s buffer)
  const meta = localCache.getSessionMeta();
  if (meta.expiresAt) {
    const expiresAt = new Date(meta.expiresAt);
    const bufferMs = 60 * 1000; // 60 seconds buffer
    if (expiresAt.getTime() - bufferMs > Date.now()) {
      return accessToken;
    }
  }

  // Try to refresh
  const refreshed = await refreshSilent();
  if (refreshed) {
    return secureStore.get("access_token");
  }

  return null;
}

/**
 * Requests a Model Token from the server.
 * @returns {Promise<string|null>} The Model Token (JWT) or null on failure
 */
async function getModelToken() {
  // 1. Get valid access token
  const accessToken = await getValidAccessToken();
  if (!accessToken) {
    console.error("ERROR: No valid session. Please login again.");
    process.exit(1);
  }

  // 2. Get plugin_secret
  const pluginSecret = secureStore.get("plugin_secret");
  if (!pluginSecret) {
    console.error("ERROR: Plugin secret not found. Please login again.");
    process.exit(1);
  }

  // 3. Generate signature parameters
  const timestamp = Math.floor(Date.now() / 1000);
  const nonce = generateNonce(16);
  const signature = generateSignature(accessToken, timestamp, nonce, pluginSecret);

  // 4. Request Model Token
  try {
    const response = await httpClient.post("/api/auth/model-token", {
      accessToken,
      timestamp,
      nonce,
      signature,
      pluginInstanceId: PLUGIN_INSTANCE_ID,
    });

    const modelToken = response.modelToken || response.data?.modelToken;
    if (!modelToken) {
      console.error("ERROR: Failed to get Model Token from server.");
      process.exit(1);
    }

    return modelToken;
  } catch (error) {
    console.error(`ERROR: Failed to get Model Token: ${error.message}`);
    process.exit(1);
  }
}

// ── CLI Entry Point ───────────────────────────────────────────

/**
 * CLI entry point for model-token command.
 * Outputs the Model Token to stdout for apiKeyHelper consumption.
 */
async function main() {
  const modelToken = await getModelToken();
  console.log(modelToken);
}

module.exports = {
  getModelToken,
  generateSignature,
  generateNonce,
};

// Run if called directly
if (require.main === module) {
  main().catch((e) => {
    console.error(`ERROR: ${e.message}`);
    process.exit(1);
  });
}
