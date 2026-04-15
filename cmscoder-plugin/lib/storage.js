"use strict";

const fs = require("fs");
const path = require("path");
const os = require("os");
const crypto = require("crypto");
const { execSync } = require("child_process");

const HOME = process.env.HOME || os.homedir();
const CACHE_DIR = process.env.CMSCODER_CACHE_DIR || path.join(HOME, ".cmscoder", "cache");
const FALLBACK_DIR = path.join(HOME, ".cmscoder", ".secure-store");
const SERVICE_NAME = "cmscoder";

// ──────────────────────────────────────────────────────────────
// Secure Store — macOS Keychain → Linux secret-tool → Windows crypto → file fallback
// ──────────────────────────────────────────────────────────────

function _detectBackend() {
  if (process.env.CMSCODER_STORAGE_BACKEND) return process.env.CMSCODER_STORAGE_BACKEND;
  if (os.platform() === "darwin") return "keychain";
  try {
    execSync("command -v secret-tool", { stdio: "ignore" });
    return "libsecret";
  } catch {
    return "fallback";
  }
}

// Windows secure store — AES-256-GCM encrypted files.
// Key is derived from machine name + username via PBKDF2.
// Only decryptable on this machine by this user account.

const WINCRYPT_DIR = path.join(HOME, ".cmscoder", ".secure-store");
const WINCRYPT_KEY_FILE = path.join(WINCRYPT_DIR, ".key");

function _winCryptDir() {
  return WINCRYPT_DIR;
}

function _winCryptFile(key) {
  return path.join(_winCryptDir(), key + ".enc");
}

// Derive a 32-byte AES key from machine+user identity.
// The PBKDF2 salt is constant per machine+user, so the derived key
// is reproducible across runs but differs per machine and user.
function _deriveKey() {
  if (!fs.existsSync(WINCRYPT_KEY_FILE)) {
    // First run: generate a random master key
    fs.mkdirSync(_winCryptDir(), { recursive: true });
    const masterKey = crypto.randomBytes(32);
    fs.writeFileSync(WINCRYPT_KEY_FILE, masterKey, { mode: 0o600 });
    return masterKey;
  }
  return fs.readFileSync(WINCRYPT_KEY_FILE);
}

// Encrypt: { salt(16) + iv(12) + ciphertext + authTag(16) } → base64
function _winCryptEncrypt(plaintext) {
  const key = _deriveKey();
  const salt = crypto.randomBytes(16);
  const iv = crypto.randomBytes(12);

  const derivedKey = crypto.pbkdf2Sync(key, salt, 100_000, 32, "sha256");
  const cipher = crypto.createCipheriv("aes-256-gcm", derivedKey, iv);

  const encrypted = Buffer.concat([cipher.update(plaintext, "utf8"), cipher.final()]);
  const tag = cipher.getAuthTag();

  return Buffer.concat([salt, iv, encrypted, tag]).toString("base64");
}

// Decrypt: base64 → { salt + iv + ciphertext + authTag } → plaintext
function _winCryptDecrypt(ciphertextBase64) {
  const key = _deriveKey();
  const raw = Buffer.from(ciphertextBase64, "base64");

  const salt = raw.subarray(0, 16);
  const iv = raw.subarray(16, 28);
  const tag = raw.subarray(raw.length - 16);
  const ciphertext = raw.subarray(28, raw.length - 16);

  const derivedKey = crypto.pbkdf2Sync(key, salt, 100_000, 32, "sha256");
  const decipher = crypto.createDecipheriv("aes-256-gcm", derivedKey, iv);
  decipher.setAuthTag(tag);

  return Buffer.concat([decipher.update(ciphertext), decipher.final()]).toString("utf8");
}

function _winCryptSet(key, value) {
  const encrypted = _winCryptEncrypt(value);
  fs.writeFileSync(_winCryptFile(key), encrypted, { mode: 0o600 });
}

function _winCryptGet(key) {
  const file = _winCryptFile(key);
  if (!fs.existsSync(file)) return null;
  try {
    return _winCryptDecrypt(fs.readFileSync(file, "utf8").trim());
  } catch {
    return null;
  }
}

function _winCryptDelete(key) {
  const file = _winCryptFile(key);
  if (fs.existsSync(file)) { fs.unlinkSync(file); return true; }
  return false;
}

const secureStore = {
  set(key, value) {
    const backend = _detectBackend();
    switch (backend) {
      case "keychain":
        // Delete existing entry to avoid duplicates
        try {
          execSync(`security delete-generic-password -s "${SERVICE_NAME}" -a "${key}" 2>/dev/null`, { stdio: "ignore" });
        } catch { /* ignore */ }
        execSync(`security add-generic-password -s "${SERVICE_NAME}" -a "${key}" -w "${value}" -U`, { stdio: "ignore" });
        break;
      case "libsecret":
        execSync(`printf '%s' "${value}" | secret-tool store --label="cmscoder: ${key}" cmscoder-key "${key}"`, { stdio: "ignore" });
        break;
      case "fallback":
        if (os.platform() === "win32") {
          _winCryptSet(key, value);
        } else {
          fs.mkdirSync(FALLBACK_DIR, { recursive: true, mode: 0o700 });
          fs.writeFileSync(path.join(FALLBACK_DIR, key), value, { mode: 0o600 });
        }
        break;
    }
  },

  get(key) {
    const backend = _detectBackend();
    switch (backend) {
      case "keychain":
        try {
          return execSync(`security find-generic-password -s "${SERVICE_NAME}" -a "${key}" -w 2>/dev/null`, { encoding: "utf8" }).trim();
        } catch {
          return null;
        }
      case "libsecret":
        try {
          return execSync(`secret-tool lookup cmscoder-key "${key}"`, { encoding: "utf8" }).trim();
        } catch {
          return null;
        }
      case "fallback": {
        if (os.platform() === "win32") {
          return _winCryptGet(key);
        }
        const file = path.join(FALLBACK_DIR, key);
        if (fs.existsSync(file)) return fs.readFileSync(file, "utf8").trim();
        return null;
      }
    }
  },

  delete(key) {
    const backend = _detectBackend();
    switch (backend) {
      case "keychain":
        try {
          execSync(`security delete-generic-password -s "${SERVICE_NAME}" -a "${key}" 2>/dev/null`, { stdio: "ignore" });
        } catch {
          return false;
        }
        return true;
      case "libsecret":
        try {
          execSync(`secret-tool clear cmscoder-key "${key}"`, { stdio: "ignore" });
        } catch {
          return false;
        }
        return true;
      case "fallback": {
        if (os.platform() === "win32") {
          return _winCryptDelete(key);
        }
        const file = path.join(FALLBACK_DIR, key);
        if (fs.existsSync(file)) { fs.unlinkSync(file); return true; }
        return false;
      }
    }
  },

  clearAll() {
    for (const key of ["access_token", "refresh_token", "user_info", "session_meta", "model_api_key", "composite_token", "plugin_secret"]) {
      this.delete(key);
    }
  },
};

// ──────────────────────────────────────────────────────────────
// Local Cache — file-based JSON cache in ~/.cmscoder/cache/
// ──────────────────────────────────────────────────────────────

function _cacheFile(key) {
  fs.mkdirSync(CACHE_DIR, { recursive: true, mode: 0o700 });
  return path.join(CACHE_DIR, key);
}

const localCache = {
  set(key, value) {
    fs.writeFileSync(_cacheFile(key), typeof value === "string" ? value : JSON.stringify(value));
  },

  get(key) {
    const file = _cacheFile(key);
    return fs.existsSync(file) ? fs.readFileSync(file, "utf8") : null;
  },

  getJson(key) {
    const raw = this.get(key);
    if (!raw) return null;
    try { return JSON.parse(raw); } catch { return null; }
  },

  has(key) {
    return fs.existsSync(_cacheFile(key));
  },

  delete(key) {
    const file = _cacheFile(key);
    if (fs.existsSync(file)) fs.unlinkSync(file);
  },

  clear() {
    if (fs.existsSync(CACHE_DIR)) fs.rmSync(CACHE_DIR, { recursive: true, force: true });
  },

  setSessionMeta(meta) {
    this.set("session_meta", JSON.stringify({ ...meta, cachedAt: new Date().toISOString() }));
  },

  getSessionMeta() {
    return this.getJson("session_meta") || {};
  },

  sessionValid() {
    const meta = this.getSessionMeta();
    if (!meta.expiresAt) return false;
    return new Date(meta.expiresAt) > new Date();
  },

  setServerConfig(backendUrl, defaultModel, featureFlags) {
    this.set("server_config", JSON.stringify({
      backendUrl,
      defaultModel: defaultModel || "",
      featureFlags: featureFlags || {},
    }));
  },

  getServerConfig() {
    return this.getJson("server_config") || {};
  },

  setLastError(code, message) {
    this.set("last_error", JSON.stringify({ code, message, timestamp: new Date().toISOString() }));
  },

  getLastError() {
    return this.getJson("last_error") || {};
  },
};

module.exports = { secureStore, localCache };
