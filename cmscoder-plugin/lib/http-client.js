"use strict";

const http = require("http");
const https = require("https");
const url = require("url");
const { localCache } = require("./storage");

function _getBaseUrl() {
  // 1. Environment variable
  if (process.env.CMSCODER_BACKEND_URL) return process.env.CMSCODER_BACKEND_URL;
  // 2. Cached server config
  const config = localCache.getServerConfig();
  if (config.backendUrl) return config.backendUrl;
  // 3. Config file
  const fs = require("fs");
  const path = require("path");
  const pluginDir = _getPluginDir();
  if (pluginDir) {
    const file = path.join(pluginDir, "config", "backend_url");
    if (fs.existsSync(file)) return fs.readFileSync(file, "utf8").trim();
  }
  throw new Error("Backend URL not configured. Please run cmscoder-init first.");
}

function _getPluginDir() {
  if (process.env.CMSCODER_PLUGIN_DIR) return process.env.CMSCODER_PLUGIN_DIR;
  const path = require("path");
  // Walk up from __dirname to find the plugin root
  let dir = __dirname;
  for (let i = 0; i < 5; i++) {
    if (fs.existsSync(path.join(dir, "lib"))) return dir;
    dir = path.dirname(dir);
  }
  return null;
}

function _request(method, path, body, headers = {}) {
  return new Promise((resolve, reject) => {
    const baseUrl = _getBaseUrl();
    const parsed = new url.URL(path, baseUrl);
    const isHttps = parsed.protocol === "https:";
    const lib = isHttps ? https : http;

    const options = {
      hostname: parsed.hostname,
      port: parsed.port,
      path: parsed.pathname + parsed.search,
      method,
      headers: {
        "Content-Type": "application/json",
        "X-Trace-Id": `cmscoder-node-${Date.now()}`,
        ...headers,
      },
    };

    const req = lib.request(options, (res) => {
      const chunks = [];
      res.on("data", (chunk) => chunks.push(chunk));
      res.on("end", () => {
        const body = Buffer.concat(chunks).toString("utf8");
        if (res.statusCode >= 200 && res.statusCode < 300) {
          try {
            const parsed = JSON.parse(body);
            // GoFrame standard response: {code, data}
            if (parsed.code !== undefined && parsed.code !== 0) {
              return reject(new Error(`API error: code=${parsed.code}`));
            }
            resolve(parsed.data || parsed);
          } catch (e) {
            if (e.message.startsWith("API error")) return reject(e);
            resolve(body);
          }
        } else {
          reject(new Error(`HTTP ${res.statusCode}: ${body}`));
        }
      });
    });

    req.on("error", reject);

    if (body) {
      req.write(typeof body === "string" ? body : JSON.stringify(body));
    }

    req.end();
  });
}

const httpClient = {
  post(path, body, token) {
    const headers = {};
    if (token) headers["Authorization"] = `Bearer ${token}`;
    return _request("POST", path, body, headers);
  },

  get(path, token) {
    const headers = {};
    if (token) headers["Authorization"] = `Bearer ${token}`;
    return _request("GET", path, undefined, headers);
  },
};

module.exports = { httpClient, _getBaseUrl, _getPluginDir };
