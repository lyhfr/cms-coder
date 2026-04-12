"use strict";

const { httpClient } = require("./http-client");
const { localCache } = require("./storage");

/**
 * Fetch bootstrap configuration from the server and cache it.
 * Called after login and optionally on startup.
 * @param {string} accessToken
 */
async function sync(accessToken) {
  const data = await httpClient.get("/api/plugin/bootstrap", accessToken);

  // Cache the raw bootstrap data.
  localCache.set("bootstrap_data", JSON.stringify(data));

  // Extract and cache default model if present.
  if (data.defaultModel) {
    localCache.set("default_model", data.defaultModel);
  }

  // Update server config with feature flags.
  const existingConfig = localCache.getServerConfig();
  if (data.featureFlags || existingConfig.backendUrl) {
    localCache.setServerConfig(
      existingConfig.backendUrl || "",
      existingConfig.defaultModel || data.defaultModel || "",
      data.featureFlags || existingConfig.featureFlags || {},
    );
  }
}

/**
 * Get cached bootstrap data.
 */
function getCached() {
  const raw = localCache.get("bootstrap_data");
  return raw ? JSON.parse(raw) : {};
}

module.exports = { sync, getCached };
