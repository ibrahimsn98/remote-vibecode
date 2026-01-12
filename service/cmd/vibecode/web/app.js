const API_BASE = window.location.origin;
const WS_BASE = API_BASE.replace('http', 'ws');

let ws = null;
let currentSession = null;
let sessions = {};
let lastMessageContent = ''; // For deduplication

function connectToSession(sessionId) {
    if (ws) ws.close();
    currentSession = sessionId;
    lastMessageContent = ''; // Reset deduplication tracker

    const session = sessions[sessionId];
    const projectName = session.project || 'Unknown Project';
    const sessionIdShort = sessionId ? sessionId.substring(0, 8) : '????????';

    document.getElementById('session-title').textContent =
        `${projectName} (${sessionIdShort})`;

    const messagesDiv = document.getElementById('messages');
    messagesDiv.innerHTML = '<div class="loading">Loading messages...</div>';

    ws = new WebSocket(`${WS_BASE}/ws?session_id=${sessionId}`);

    ws.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        switch(msg.type) {
            case 'history':
                displayMessages(msg.payload?.messages || []);
                break;
            case 'update':
                appendMessage(msg.payload?.message);
                break;
            case 'session_message':
                appendMessage(msg.payload);
                break;
            case 'status':
                updateStatus(sessionId, msg.payload?.status);
                break;
        }
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
        console.log('WebSocket closed');
    };
}

function displayMessages(messages) {
    const messagesDiv = document.getElementById('messages');
    messagesDiv.innerHTML = '';

    if (messages.length === 0) {
        messagesDiv.innerHTML = '<div class="empty-state">No messages yet</div>';
        return;
    }

    messages.forEach(msg => {
        appendMessage(msg, false);
    });
    scrollToBottom();
}

function updateStatus(sessionId, status) {
    if (sessions[sessionId]) {
        sessions[sessionId].status = status;
        updateSessionList();
        if (currentSession === sessionId) {
            updateInputVisibility();
        }
    }
}

function updateSessionList() {
    const list = document.getElementById('session-list');
    list.innerHTML = '';

    const sessionArray = Object.values(sessions).sort((a, b) => {
        const aTime = getLastActivityTime(a);
        const bTime = getLastActivityTime(b);
        return bTime - aTime;
    });

    sessionArray.forEach(session => {
        const item = document.createElement('div');
        item.className = 'session-item';
        if (session.id === currentSession) {
            item.classList.add('active');
        }

        const projectName = session.project || 'Unknown Project';
        const sessionId = session.id ? session.id.substring(0, 8) : '????????';
        const status = session.status || 'active';
        const isTmux = session.type === 'tmux';

        item.innerHTML = `
            <div class="session-item-header">
                <div class="session-item-name">
                    ${isTmux ? '<span class="tmux-badge">tmux</span> ' : ''}
                    ${escapeHtml(projectName)}
                </div>
                <div class="session-status status-${status}">${status}</div>
            </div>
            <div class="session-item-id">${escapeHtml(sessionId)}</div>
        `;

        item.onclick = () => selectSession(session.id);
        list.appendChild(item);
    });
}

function getLastActivityTime(session) {
    if (session.messages && session.messages.length > 0) {
        const lastMessage = session.messages[session.messages.length - 1];
        return lastMessage.timestamp || 0;
    }
    return 0;
}

function selectSession(sessionId) {
    connectToSession(sessionId);
    updateSessionList();
    updateInputVisibility();
}

function appendMessage(msg, shouldScroll = true) {
    // Support both old format (role, content) and new format (message object)
    let role, content;
    if (typeof msg === 'string') {
        role = arguments[0];
        content = arguments[1];
        shouldScroll = arguments[2] !== undefined ? arguments[2] : true;
    } else {
        role = msg.role;
        content = msg.content;
    }

    // Skip empty or duplicate messages
    const strippedContent = stripAnsi(content);
    if (!strippedContent || strippedContent.trim() === '') {
        return;
    }
    if (strippedContent === lastMessageContent) {
        return; // Skip duplicate
    }
    lastMessageContent = strippedContent;

    const messagesDiv = document.getElementById('messages');

    // Remove empty state if present
    const emptyState = messagesDiv.querySelector('.empty-state');
    if (emptyState) {
        emptyState.remove();
    }

    // Remove loading state if present
    const loadingState = messagesDiv.querySelector('.loading');
    if (loadingState) {
        loadingState.remove();
    }

    const div = document.createElement('div');
    div.className = `message ${role}`;

    const label = document.createElement('div');
    label.className = 'message-header';
    label.textContent = role === 'user' ? '👤 User' : '🤖 Claude';

    const contentDiv = document.createElement('div');
    contentDiv.className = 'message-content';
    contentDiv.textContent = stripAnsi(content);

    div.appendChild(label);
    div.appendChild(contentDiv);
    messagesDiv.appendChild(div);

    if (shouldScroll) {
        scrollToBottom();
    }
}

function scrollToBottom() {
    const messagesDiv = document.getElementById('messages');
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Strip ANSI escape codes from text
function stripAnsi(text) {
    // Remove ANSI escape sequences
    return text.replace(/\x1b\[[0-9;]*m/g, '');
}

// Input form handler
const inputForm = document.getElementById('input-form');
const userInput = document.getElementById('user-input');
const inputContainer = document.getElementById('input-container');

inputForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    const input = userInput.value.trim();
    if (!input || !currentSession) {
        return;
    }

    // Check if this is a tmux session
    const session = sessions[currentSession];
    const isTmux = session && session.type === 'tmux';

    // Use appropriate endpoint based on session type
    const endpoint = isTmux
        ? `${API_BASE}/api/v1/tmux/sessions/${currentSession}/input`
        : `${API_BASE}/api/v1/sessions/${currentSession}/input`;

    const body = isTmux
        ? JSON.stringify({ input, send_enter: true })
        : JSON.stringify({ input });

    // Send to service
    try {
        const resp = await fetch(endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body
        });

        if (resp.ok) {
            userInput.value = '';
        } else {
            console.error('Failed to send input');
        }
    } catch (e) {
        console.error('Error sending input:', e);
    }
});

function updateInputVisibility() {
    const session = sessions[currentSession];
    const isTmux = session && session.type === 'tmux';

    // Show input for tmux sessions (always interactive) or when awaiting input
    if (session && (isTmux || session.status === 'awaiting_input')) {
        inputContainer.style.display = 'block';
        userInput.focus();
    } else {
        inputContainer.style.display = 'none';
    }
}

// Initial load of existing sessions
async function loadSessions() {
    try {
        // Load tmux sessions
        const tmuxResp = await fetch(`${API_BASE}/api/v1/tmux/sessions`);
        const tmuxData = await tmuxResp.json();

        if (tmuxData.sessions && tmuxData.sessions.length > 0) {
            tmuxData.sessions.forEach(session => {
                if (session.id) {
                    sessions[session.id] = {
                        id: session.id,
                        type: 'tmux',
                        project: session.session_name || 'Unknown',
                        messages: session.messages || [],
                        status: session.status || 'active',
                        session_name: session.session_name,
                        pane_id: session.pane_id
                    };
                }
            });
            updateSessionList();
        }
    } catch (e) {
        console.error('Failed to load sessions:', e);
    }
}

// Initialize
loadSessions();
// Auto-refresh sessions every 2 seconds for tmux auto-discovery
setInterval(loadSessions, 2000);
