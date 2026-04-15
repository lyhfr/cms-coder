#!/usr/bin/env node
"use strict";

// Thin CLI wrapper — delegates all logic to lib modules.
// Also exports functions for programmatic use: require('./cmscoder')

const {
  login,
  logout,
  refreshSilent,
  ensureSession,
  getAccessToken,
  getModelApiKey,
  getCompositeToken,
  status: cmdStatus,
} = require("./auth");

const { sync: bootstrapSync } = require("./bootstrap");
const { getModelToken } = require("./model-token");
const modelProxy = require("./model-proxy");

// Re-export for programmatic use.
module.exports = {
  login,
  logout,
  refreshSilent,
  ensureSession,
  getAccessToken,
  getModelApiKey,
  getCompositeToken,
  status: cmdStatus,
  bootstrapSync,
  getModelToken,
  modelProxy,
};

// CLI entry point.
if (require.main === module) {
  const command = process.argv[2];
  const commands = {
    login,
    logout,
    refresh: refreshSilent,
    status: cmdStatus,
    token: async () => {
      const token = await getAccessToken();
      if (!token) {
        console.error("No valid session. Please login.");
        process.exit(1);
      }
      console.log(token);
    },
    "model-token": async () => {
      const token = await getModelToken();
      console.log(token);
    },
    "model-proxy": async () => {
      const subcommand = process.argv[3] || "start";
      await modelProxy[subcommand]();
    },
    "ensure-session": async () => {
      if (await ensureSession()) process.exit(0);
      process.exit(1);
    },
  };

  const fn = commands[command];
  if (!fn) {
    console.error(`Usage: cmscoder <command>
Commands:
  login            Full login flow (opens browser)
  logout           Revoke session and clear local data
  refresh          Silent session refresh
  status           Show current session status
  token            Print current access token to stdout
  model-token      Get short-lived Model Token for API access (Claude Code apiKeyHelper)
  model-proxy      Start/stop local proxy for OpenCode (start|stop|status)
  ensure-session   Check and restore session (for hooks)`);
    process.exit(1);
  }

  fn().catch((e) => {
    console.error(`ERROR: ${e.message}`);
    process.exit(1);
  });
}
