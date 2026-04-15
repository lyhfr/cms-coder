# cmscoder-login (OpenCode)

OpenCode adapter for cmscoder login skill. Uses the shared JS plugin layer.

## Instructions
1. Run: `node "$CMSCODER_PLUGIN_DIR/lib/cmscoder.js" login`
2. Complete SSO authentication in browser
3. After login, start the local proxy for model access:
   ```bash
   node "$CMSCODER_PLUGIN_DIR/lib/cmscoder.js" model-proxy start
   ```

## Model Access for OpenCode

OpenCode does not support apiKeyHelper like Claude Code. Instead, use the local proxy:

1. **Start proxy**: `cmscoder model-proxy start`
   - macOS/Linux: Uses Unix Domain Socket at `~/.cmscoder/model-proxy.sock`
   - Windows: Uses TCP on `127.0.0.1` with dynamic port
   - Automatically writes OpenCode config to `~/.opencode/config.json`

2. **Check status**: `cmscoder model-proxy status`

3. **Stop proxy**: `cmscoder model-proxy stop`

The proxy automatically:
- Obtains short-lived Model Tokens (5 min TTL)
- Refreshes tokens before expiry
- Adds Authorization header to model requests
