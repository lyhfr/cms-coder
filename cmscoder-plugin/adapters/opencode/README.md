# cmscoder — Enterprise AI Coding Assistant (OpenCode)

cmscoder provides enterprise-grade AI coding capabilities for OpenCode, integrating your organization's identity, policies, and development standards.

## Quick Commands

- `/cmscoder-login` — Authenticate with enterprise account
- `/cmscoder-status` — View current session status
- To log out, run: `node "$CMSCODER_PLUGIN_DIR/lib/cmscoder.js" logout`

## Authentication

cmscoder uses enterprise SSO (IAM) for authentication. Configuration is shared with the Claude Code adapter.

## Development Standards

1. **YAGNI** — Implement only what is required.
2. **DRY** — Avoid duplicating logic.
3. **KISS** — Prefer simple, readable solutions.
4. **Test-first** — Write tests before implementation.
5. **Commit incrementally** — Each logical change should be its own commit.
