# cmscoder-status

## Description
Display the current cmscoder session status, including authentication state, user info, tenant, session expiry, default model, and any recent errors.

## When to use
- User asks about their login status
- User wants to see current session information
- User wants to check which tenant or project they are on
- User wants to verify model access

## Instructions
1. Run: `node "$CMSCODER_PLUGIN_DIR/lib/cmscoder.js" status`
2. Display the output to the user

## Output format
The output includes:
- Authentication state (logged in / not logged in)
- User display name and email
- Tenant ID
- Session remaining time
- Default model (if configured)
- Last error (if any)
- Recommended actions (e.g., re-login if expired)
