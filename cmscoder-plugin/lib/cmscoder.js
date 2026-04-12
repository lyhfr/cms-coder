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
  status: cmdStatus,
} = require("./auth");

const { sync: bootstrapSync } = require("./bootstrap");

// Re-export for programmatic use.
module.exports = {
  login,
  logout,
  refreshSilent,
  ensureSession,
  getAccessToken,
  status: cmdStatus,
  bootstrapSync,
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
  ensure-session   Check and restore session (for hooks)`);
    process.exit(1);
  }

  fn().catch((e) => {
    console.error(`ERROR: ${e.message}`);
    process.exit(1);
  });
}
