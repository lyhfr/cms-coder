#!/usr/bin/env bash
# session-start hook — Check cmscoder authentication on session start
# This hook runs when a new Claude Code session begins

if [[ -z "${CMSCODER_PLUGIN_DIR}" ]]; then
    # Try to auto-detect plugin directory
    if [[ -d "${HOME}/.cmscoder/plugin" ]]; then
        CMSCODER_PLUGIN_DIR="${HOME}/.cmscoder/plugin"
    else
        exit 0
    fi
fi

# Check session validity via JS plugin
if ! node "$CMSCODER_PLUGIN_DIR/lib/cmscoder.js" ensure-session 2>/dev/null; then
    echo ""
    echo "cmscoder: Not authenticated. Run /cmscoder-login to sign in with your enterprise account."
    echo ""
fi
