"use strict";

const http = require("http");
const fs = require("fs");
const path = require("path");
const os = require("os");

const TIMEOUT_MS = 5 * 60 * 1000; // 5 minutes

/**
 * Start a local loopback HTTP server that waits for a single /callback request
 * containing `login_ticket`.
 *
 * Returns { port: Promise<number>, waitForTicket: Promise<string>, close: () => void }
 */
function startCallbackServer() {
  const cacheDir = process.env.CMSCODER_CACHE_DIR || path.join(os.homedir(), ".cmscoder", "cache");
  const ticketFile = path.join(cacheDir, "callback_ticket");

  fs.mkdirSync(cacheDir, { recursive: true, mode: 0o700 });
  try { fs.unlinkSync(ticketFile); } catch { /* ignore */ }

  let ticketResolve = null;
  let ticketReject = null;
  let closed = false;
  let server = null;

  const waitForTicket = new Promise((resolve, reject) => {
    ticketResolve = resolve;
    ticketReject = reject;
  });

  // Timeout for the ticket.
  const ticketTimer = setTimeout(() => {
    if (!closed) {
      close();
      if (ticketReject) ticketReject(new Error("Callback server timed out after 5 minutes"));
    }
  }, TIMEOUT_MS);

  server = http.createServer((req, res) => {
    const url = new URL(req.url, "http://127.0.0.1");

    if (url.pathname === "/callback") {
      const ticket = url.searchParams.get("login_ticket");
      if (ticket) {
        fs.writeFileSync(ticketFile, ticket);
        res.writeHead(200, { "Content-Type": "text/html; charset=utf-8" });
        res.end(
          '<html><head><title>cmscoder Login</title></head>' +
          '<body style="font-family:system-ui,sans-serif;text-align:center;padding:60px 20px;">' +
          '<h2>Login Successful</h2>' +
          '<p>You can close this tab and return to your editor.</p>' +
          '<script>setTimeout(function(){window.close();}, 2000);</script>' +
          '</body></html>'
        );
        clearTimeout(ticketTimer);
        if (ticketResolve) {
          ticketResolve(ticket);
          ticketResolve = null;
          ticketReject = null;
        }
        close();
      } else {
        res.writeHead(400, { "Content-Type": "text/plain" });
        res.end("Missing login_ticket parameter");
      }
    } else if (url.pathname === "/health") {
      res.writeHead(200, { "Content-Type": "text/plain" });
      res.end("ok");
    } else {
      res.writeHead(404);
      res.end();
    }
  });

  function close() {
    if (!closed && server) {
      server.close();
      closed = true;
    }
  }

  const portPromise = new Promise((resolvePort, rejectPort) => {
    server.on("error", (err) => {
      clearTimeout(ticketTimer);
      rejectPort(err);
      close();
    });
    server.listen(0, "127.0.0.1", () => {
      const addr = server.address();
      if (addr && typeof addr.port === "number") {
        resolvePort(addr.port);
      } else {
        clearTimeout(ticketTimer);
        rejectPort(new Error("Failed to get server address"));
        close();
      }
    });
  });

  return { port: portPromise, waitForTicket, close };
}

module.exports = { startCallbackServer };
