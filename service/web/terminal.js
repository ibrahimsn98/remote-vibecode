// Tmux terminal client with xterm.js using gotty protocol
const API_BASE = window.location.origin;

// Terminal state
let terminals = {};  // sessionId -> { term, fitAddon, ws, writtenContent }
let currentSessionId = null;
let sessions = {};    // sessionId -> { id, name, created_at }
let unreadSessions = new Set();  // Track sessions with unread updates

// Gotty protocol constants
const GOTTY_OUTPUT = '1';
const GOTTY_INPUT = '1';
const GOTTY_PING = '2';
const GOTTY_PONG = '3';
const GOTTY_RESIZE = '4';

// Sessions WebSocket for real-time updates
let sessionsWs = null;

// Initialize
async function init() {
    // Initial load via REST API
    await loadSessions();

    // Connect to sessions WebSocket for real-time updates
    connectSessionsWebSocket();
}

// Connect to sessions WebSocket for real-time updates
function connectSessionsWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/v1/sessions/ws`;

    sessionsWs = new WebSocket(wsUrl);

    sessionsWs.onopen = () => {
        console.log('Sessions WebSocket connected');
    };

    sessionsWs.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            if (data.type === 'sessions' && data.sessions) {
                handleSessionsUpdate(data.sessions);
            }
        } catch (e) {
            console.error('Failed to parse sessions WebSocket message:', e);
        }
    };

    sessionsWs.onclose = () => {
        console.log('Sessions WebSocket disconnected, reconnecting...');
        // Reconnect after 2 seconds
        setTimeout(connectSessionsWebSocket, 2000);
    };

    sessionsWs.onerror = (error) => {
        console.error('Sessions WebSocket error:', error);
    };
}

// Handle sessions update from WebSocket
function handleSessionsUpdate(sessionList) {
    const newSessions = {};

    sessionList.forEach(s => {
        newSessions[s.id] = {
            id: s.id,
            name: s.session_name || 'Unknown',
            created_at: s.last_capture || s.created_at || Date.now() / 1000
        };
    });

    sessions = newSessions;
    updateSessionList();
}

// Load all tmux sessions
async function loadSessions() {
    try {
        const tmuxResp = await fetch('/api/v1/tmux/sessions');
        const tmuxData = await tmuxResp.json();

        const newSessions = {};

        if (tmuxData.sessions) {
            tmuxData.sessions.forEach(s => {
                newSessions[s.id] = {
                    id: s.id,
                    name: s.session_name || 'Unknown',
                    created_at: s.last_capture ? new Date(s.last_capture).getTime() / 1000 : Date.now() / 1000
                };
            });
        }

        sessions = newSessions;
        updateSessionList();
    } catch (e) {
        console.error('Failed to load sessions:', e);
    }
}

// Update session list in sidebar
function updateSessionList() {
    const list = document.getElementById('session-list');
    list.innerHTML = '';

    const sessionArray = Object.values(sessions).sort((a, b) => {
        // Primary sort: by created_at (newest first)
        if (b.created_at !== a.created_at) {
            return b.created_at - a.created_at;
        }
        // Secondary sort: by session name (alphabetical) for stability
        return a.name.localeCompare(b.name);
    });

    if (sessionArray.length === 0) {
        list.innerHTML = '<div style="padding: 16px; color: #a0a0a0;">No tmux sessions found</div>';
        return;
    }

    sessionArray.forEach(session => {
        const item = document.createElement('div');
        item.className = 'session-item';
        if (session.id === currentSessionId) {
            item.classList.add('active');
        }

        const hasUnread = unreadSessions.has(session.id);

        item.innerHTML = `
            <div class="session-item-header">
                <div class="session-item-name">
                    ${escapeHtml(session.name)}
                    ${hasUnread ? '<span class="unread-badge"></span>' : ''}
                </div>
            </div>
            <div class="session-item-id">${escapeHtml(session.id.substring(0, 8))}</div>
        `;

        item.onclick = () => selectSession(session.id);
        list.appendChild(item);
    });
}

// Select a tmux session
function selectSession(sessionId) {
    currentSessionId = sessionId;
    // Clear unread flag for this session
    unreadSessions.delete(sessionId);
    updateSessionList();

    const session = sessions[sessionId];
    if (!session) {
        console.warn('Session not found:', sessionId);
        return;
    }

    // Hide all terminals
    Object.values(terminals).forEach(t => {
        if (t.term && t.term.element) {
            t.term.element.style.display = 'none';
        }
    });

    // Create or show terminal for this session
    if (terminals[sessionId]) {
        // Show existing terminal
        terminals[sessionId].term.element.style.display = 'block';
        terminals[sessionId].fitAddon.fit();
        terminals[sessionId].term.focus();
    } else {
        // Create new terminal
        createTerminalForSession(sessionId, session.name);
    }
}

// Create terminal for a tmux session
function createTerminalForSession(sessionId, tmuxSessionName) {
    const container = document.getElementById('terminal-container');

    // Clear container if this is the first terminal
    if (Object.keys(terminals).length === 0) {
        container.innerHTML = '';
    }

    // Create xterm.js instance
    const term = new Terminal({
        cursorBlink: true,
        fontSize: 14,
        fontFamily: 'SF Mono, Monaco, Consolas, monospace',
        theme: {
            background: '#000000',
            foreground: '#ffffff',
            cursor: '#CC785C',
            selection: 'rgba(204, 120, 92, 0.3)',
            black: '#000000',
            red: '#ff6b6b',
            green: '#51cf66',
            yellow: '#ffd43b',
            blue: '#339af0',
            magenta: '#cc5de8',
            cyan: '#22b8cf',
            white: '#ffffff',
            brightBlack: '#495057',
            brightRed: '#ff8787',
            brightGreen: '#69db7c',
            brightYellow: '#ffe066',
            brightBlue: '#4dabf7',
            brightMagenta: '#da77f2',
            brightCyan: '#66d9e8',
            brightWhite: '#ffffff',
        },
        allowProposedApi: true,
        scrollback: 1000,
    });

    const fitAddon = new FitAddon.FitAddon();
    term.loadAddon(fitAddon);

    // Create wrapper div for this terminal
    const wrapper = document.createElement('div');
    wrapper.className = 'terminal-wrapper';
    wrapper.style.cssText = 'position: absolute; top: 0; left: 0; right: 0; bottom: 0; display: none;';
    container.appendChild(wrapper);
    term.open(wrapper);
    fitAddon.fit();

    // Focus the terminal initially
    term.focus();

    // Show this terminal
    wrapper.style.display = 'block';

    // Store terminal
    terminals[sessionId] = { term, fitAddon, ws: null, pingTimer: null };

    // Focus terminal on click
    wrapper.addEventListener('click', () => {
        term.focus();
    });

    // Handle keyboard input - send via gotty protocol
    term.onData((data) => {
        const terminal = terminals[sessionId];
        if (terminal && terminal.ws && terminal.ws.readyState === WebSocket.OPEN) {
            // Send input with gotty protocol prefix
            const encoder = new TextEncoder();
            const bytes = encoder.encode(data);
            const msg = GOTTY_INPUT + Array.from(bytes).map(b => String.fromCharCode(b)).join('');
            terminal.ws.send(msg);
        }
    });

    // Handle resize
    const resizeObserver = new ResizeObserver(() => {
        if (terminals[sessionId]) {
            terminals[sessionId].fitAddon.fit();
            // Send resize via gotty protocol
            sendResize(sessionId);
        }
    });
    resizeObserver.observe(wrapper);

    // Connect to gotty WebSocket
    connectGotty(sessionId, tmuxSessionName);
}

// Send terminal resize via gotty protocol
function sendResize(sessionId) {
    const terminal = terminals[sessionId];
    if (!terminal || !terminal.ws || terminal.ws.readyState !== WebSocket.OPEN) {
        return;
    }

    const dims = terminal.term;
    const cols = dims.cols;
    const rows = dims.rows;

    // Format: columns,rows as ASCII
    const resizePayload = `${cols},${rows}`;
    terminal.ws.send(GOTTY_RESIZE + resizePayload);
}

// Connect to gotty WebSocket
function connectGotty(sessionId, tmuxSessionName) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // Use the gotty endpoint with tmux session name
    const wsUrl = `${protocol}//${window.location.host}/gotty/${encodeURIComponent(tmuxSessionName)}`;

    const terminal = terminals[sessionId];
    if (!terminal) return;

    const ws = new WebSocket(wsUrl);
    terminal.ws = ws;

    ws.onopen = () => {
        console.log('Gotty connected:', tmuxSessionName);
        terminal.term.write('\x1b[38;5;215m*** Connected to tmux ***\x1b[0m\r\n');
        terminal.term.focus();

        // Start ping timer for keepalive
        if (terminal.pingTimer) {
            clearInterval(terminal.pingTimer);
        }
        terminal.pingTimer = setInterval(() => {
            if (ws.readyState === WebSocket.OPEN) {
                ws.send(GOTTY_PING);
            }
        }, 30000);
    };

    ws.onmessage = (event) => {
        if (!event.data || event.data.length === 0) {
            return;
        }

        // Parse gotty protocol
        const messageType = event.data[0];
        const payload = event.data.slice(1);

        switch (messageType) {
            case GOTTY_OUTPUT:
                // If this session is not currently active, mark as unread
                if (currentSessionId !== sessionId) {
                    unreadSessions.add(sessionId);
                    updateSessionList();
                }
                // Output is base64 encoded - need to handle UTF-8 properly
                try {
                    const decoded = base64ToUtf8(payload);
                    terminal.term.write(decoded);
                } catch (e) {
                    console.error('Failed to decode gotty output:', e);
                }
                break;

            case GOTTY_PING:
                // Respond with pong
                if (ws.readyState === WebSocket.OPEN) {
                    ws.send(GOTTY_PONG);
                }
                break;

            case GOTTY_PONG:
                // Pong received - connection alive
                break;

            case GOTTY_RESIZE:
                // Server-initiated resize (not typically used)
                break;

            default:
                console.log('Unknown gotty message type:', messageType);
        }
    };

    ws.onclose = () => {
        console.log('Gotty closed:', tmuxSessionName);
        terminal.term.write('\r\n\x1b[38;5;215m*** Connection closed ***\x1b[0m\r\n');

        // Clear ping timer
        if (terminal.pingTimer) {
            clearInterval(terminal.pingTimer);
            terminal.pingTimer = null;
        }
    };

    ws.onerror = (error) => {
        console.error('Gotty error:', error);
        terminal.term.write('\r\n\x1b[38;5;203m*** Connection error ***\x1b[0m\r\n');
    };
}

// Expose selectSession to window for mobile sidebar functionality
window.selectSession = selectSession;

// Auto-connect to first available session
window.addEventListener('DOMContentLoaded', () => {
    init().then(() => {
        const sessionArray = Object.values(sessions);
        if (sessionArray.length > 0) {
            selectSession(sessionArray[0].id);
        } else {
            document.getElementById('terminal-container').innerHTML =
                '<div style="display:flex;align-items:center;justify-content:center;height:100%;color:#a0a0a0;">No tmux sessions found.<br>Create one with: tmux new-session -s claude</div>';
        }
    });
});

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Decode base64 to UTF-8 string
// atob() returns Latin-1, so we need to convert to UTF-8 properly
function base64ToUtf8(base64) {
    const binaryString = atob(base64);
    const bytes = new Uint8Array(binaryString.length);
    for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i);
    }
    return new TextDecoder('utf-8').decode(bytes);
}
