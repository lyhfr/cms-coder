# cmscoder — Enterprise AI Coding Assistant

cmscoder provides enterprise-grade AI coding capabilities, integrating your organization's identity, policies, and development standards.

## Quick Commands

- `/cmscoder-login` — Authenticate with enterprise account
- `/cmscoder-status` — View current session status, user info, and configuration
- To log out, run: `node "$CMSCODER_PLUGIN_DIR/lib/cmscoder.js" logout`

## Authentication

cmscoder uses enterprise SSO (IAM) for authentication. If you are not logged in, you will be prompted to authenticate via your browser.

### Session Management

- Sessions are automatically checked on startup
- Access tokens are silently refreshed before expiration
- If refresh fails, you will be prompted to re-login
- Sessions expire after the configured duration (default: access token 15min, refresh token 7 days)

## Configuration

Backend endpoint and settings are managed by your organization administrator. You do not need to configure API keys or model endpoints manually.

## Development Standards

When working with cmscoder, follow these guidelines:

1. **YAGNI** — Implement only what is required. Do not add features or abstractions beyond the current scope.
2. **DRY** — Avoid duplicating logic. Reuse existing functions and patterns.
3. **KISS** — Prefer simple, readable solutions over clever or complex ones.
4. **Test-first** — Write tests before implementation when applicable.
5. **Commit incrementally** — Each logical change should be its own commit with a clear message.

## Security Reminders

- Never commit secrets, API keys, or credentials to the repository
- Use environment variables or secret management for sensitive configuration
- Review all generated code for security vulnerabilities before committing
- Follow OWASP Top 10 guidelines when writing code
