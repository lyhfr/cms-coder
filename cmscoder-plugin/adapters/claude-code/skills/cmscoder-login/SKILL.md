# cmscoder-login

## Description
Authenticate with the cmscoder enterprise SSO system. Opens a browser for IAM login, receives the callback, and stores session credentials securely.

## When to use
- User explicitly requests to login to cmscoder
- User needs to authenticate before using enterprise model access
- Session has expired and re-authentication is required

## Instructions
1. Run: `node "$CMSCODER_PLUGIN_DIR/lib/cmscoder.js" login`
2. The login process will:
   - Start a local callback server on 127.0.0.1
   - Open the system browser with the SSO login URL
   - Wait for the user to complete authentication
   - Exchange the callback ticket for session credentials
   - Store tokens securely in the OS keychain
   - Sync bootstrap configuration from the server
3. Report the result to the user

## Error handling
- If the backend URL is not configured, inform the user to run cmscoder-init first
- If the browser callback times out after 5 minutes, report the timeout and suggest retrying
- If the token exchange fails, show the error message from the server
- If keychain access fails (e.g., in a headless environment), inform the user that secure storage is unavailable
