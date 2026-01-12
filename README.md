```
.................................................
.#####...######..##...##...####...######..######.
.##..##..##......###.###..##..##....##....##.....
.#####...####....##.#.##..##..##....##....####...
.##..##..##......##...##..##..##....##....##.....
.##..##..######..##...##...####.....##....######.
.................................................
.##..##..######..#####...######...####....####...#####...######.
.##..##....##....##..##..##......##..##..##..##..##..##..##.....
.##..##....##....#####...####....##......##..##..##..##..####...
..####.....##....##..##..##......##..##..##..##..##..##..##.....
...##....######..#####...######...####....####...#####...######.
................................................................
```

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8E?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

# Remote Vibecode [WIP]

A web dashboard for viewing and interacting with terminal sessions in real-time. Access your terminal from any browser, share sessions with teammates, and manage multiple terminal environments from a single interface.

Built on tmux and gotty, Remote Vibecode provides a clean, modern web interface for terminal sharing and remote access.

## What is Remote Vibecode?

Remote Vibecode transforms your terminal into a shareable, browser-accessible workspace. Whether you're pair programming, providing remote support, teaching workshops, or managing multiple server sessions, Remote Vibecode gives you a centralized dashboard for all your terminal needs.

### Use Cases

- **Pair Programming** - Share your terminal with a colleague in real-time
- **Remote Access** - Access your development environment from any device
- **Teaching & Workshops** - Demonstrate terminal workflows to students
- **Server Management** - Manage multiple server sessions from one browser tab
- **Code Reviews** - Walk through code together in a shared terminal

## How It Works

Remote Vibecode runs a web server that connects to your terminal sessions using tmux. The web interface uses xterm.js for full terminal emulation and communicates over WebSockets using the gotty protocol.

```
Browser (xterm.js) <---> WebSocket (gotty protocol) <---> tmux session
```

## Features

- **Real-time Terminal Access** - Full terminal emulation in your browser
- **Read-Only Sessions** - Sessions are read-only by default for safe viewing
- **Writable Sessions** - Use `-w` flag to allow web clients to type
- **Auto-discovery** - Sessions appear automatically as you create them
- **Multi-session Support** - Manage multiple sessions with a sidebar navigation
- **Responsive Design** - Works on desktop, tablet, and mobile
- **UTF-8 Support** - Full Unicode character support
- **Keepalive Connections** - Stable WebSocket connections with ping/pong

## Prerequisites

- macOS with Homebrew
- tmux (installed automatically by Homebrew)

## Installation

```bash
brew tap ibrahimsn98/homebrew-remote-vibecode
brew install rv
```

## Quick Start

### Step 1: Start the Web Server

```bash
rv serve
```

The server runs on http://localhost:7676 by default.

### Step 2: Start a Session

In a new terminal:

```bash
rv start my-project
```

This creates a **read-only** session - viewers can see but not type.

### Step 3: Open in Browser

Navigate to http://localhost:7676 and click on your session to connect.

That's it! You now have a browser-based terminal connected to your session.

## CLI Reference

### Start the Web Server

```bash
rv serve [--host 0.0.0.0] [--port 9000]
```

Starts the web server for remote terminal viewing.

**Options:**
- `--host` - Host to bind to (default: 127.0.0.1)
- `--port` - Port to listen on (default: 7676)

**Examples:**
```bash
rv serve                                    # Default: 127.0.0.1:7676
rv serve --host 0.0.0.0                     # Allow network access
rv serve --port 9000                        # Custom port
rv serve --host 192.168.1.100 --port 9000   # Both custom
```

### Start a New Session

```bash
rv start <session-name> [-w]
```

Creates and starts a new named tmux session.

**Options:**
- `-w, --writable` - Create a writable session (web clients can type)

**Examples:**
```bash
rv start frontend           # Read-only session (default)
rv start -w backend         # Writable session
rv start database -w        # Writable session
```

### List Sessions

```bash
rv list
```

Shows all active tmux sessions.

### Stop a Session

```bash
rv stop <session-name> [-f]
```

Stops and removes the specified session.

**Options:**
- `-f, --force` - Skip confirmation prompt

### Join a Session

```bash
rv join [session-name]
```

Attaches your terminal to an existing session (useful for direct terminal access).

## Usage Examples

### Multiple Sessions for Different Projects

```bash
# Terminal 1: Start the web server
rv serve

# Terminal 2: Start sessions
rv start frontend
rv start -w backend
rv start database

# Open http://localhost:7676
# Switch between sessions using the sidebar
```

### Read-Only Viewing Session

```bash
# Start a read-only session (default)
rv start demo-session

# Viewers can see your terminal but cannot type
```

### Writable Collaboration Session

```bash
# Start a writable session
rv start -w pair-session

# Both you and web viewers can type
# Great for pair programming
```

### Custom Server Configuration

```bash
# Allow access from other devices on your network
rv serve --host 0.0.0.0 --port 9000

# Access from another device
# http://YOUR_LOCAL_IP:9000
```

## Session Modes

### Read-Only (Default)

Sessions are read-only by default. Web viewers can:
- See everything happening in the terminal
- Scroll through history
- **Cannot** type or execute commands

Perfect for:
- Demonstrations
- Teaching
- Monitoring
- Safe code review

### Writable (with `-w` flag)

Sessions created with `-w` allow web clients to type.

Web viewers can:
- See everything
- Type commands
- Interact with the terminal

Perfect for:
- Pair programming
- Remote collaboration
- Interactive debugging

**Security Note:** Only use writable sessions with trusted collaborators.

## Configuration

The `rv serve` command uses CLI flags instead of environment variables:

### Port

```bash
rv serve --port 9000
```

### Host

```bash
# Local only (default)
rv serve --host 127.0.0.1

# Allow network access
rv serve --host 0.0.0.0

# Specific IP
rv serve --host 192.168.1.100
```

### Combined

```bash
rv serve --host 0.0.0.0 --port 9000
```

## Remote Connection

For secure remote access to your terminal sessions, use one of the following methods. **Never expose the service directly to the internet without authentication**.

### Option 1: Tailscale VPN

Tailscale creates a private network between your devices, allowing secure access from anywhere.

```bash
# Install Tailscale
brew install --cask tailscale

# Login and connect to your tailnet
tailscale up

# Find your Tailscale IP
tailscale ip -4
```

Access using your Tailscale IP: `http://<your-tailscale-ip>:7676`

### Option 2: Cloudflare Tunnels + Zero Trust

```bash
# Install cloudflared
brew install cloudflared

# Login to Cloudflare
cloudflared tunnel login

# Create a tunnel
cloudflared tunnel create rv

# Run the tunnel
cloudflared tunnel --url http://localhost:7676
```

Set up Zero Trust access in the Cloudflare dashboard for authentication.

## Troubleshooting

### Server Not Starting

```bash
# Check if port is already in use
lsof -i :7676

# Try a different port
rv serve --port 9000
```

### Session Not Appearing

```bash
# Verify the session is running
rv list

# Check the web console for WebSocket errors
# Ensure the server is running
```

### Browser Connection Issues

- Check that http://localhost:7676 is accessible
- Try a different browser
- Check your firewall settings
- Verify the server is running on the expected port

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/ibrahimsn98/remote-vibecode.git
cd remote-vibecode

# Build the binary
cd service && go build -o rv ./cmd/vibecode

# Run locally
./rv serve
```

### Project Structure

```
remote-vibecode/
├── service/
│   ├── cmd/vibecode/       # CLI entry point (rv command)
│   │   ├── main.go         # Main CLI with serve command
│   │   ├── commands/        # CLI subcommands (start, stop, list, join)
│   │   ├── web/            # Embedded web dashboard
│   │   └── internal/
│   │       └── banner/      # Startup banner
│   ├── internal/
│   │   ├── api/            # REST API handlers
│   │   ├── gotty/          # Gotty protocol implementation
│   │   ├── tmux/           # Session management
│   │   ├── session/        # Session tracking
│   │   └── ws/             # WebSocket handlers
│   └── web/                # Source web dashboard (embedded)
└── README.md
```

### Running Tests

```bash
cd service
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Security Disclaimer

**IMPORTANT**: The `rv serve` command runs on `127.0.0.1` (localhost) by default and is **not exposed to external networks**.

**Never use `--host 0.0.0.0`** without proper security measures. Doing so creates a **Remote Code Execution (RCE) vulnerability** - anyone on your network can access your terminal sessions without authentication.

For remote access, always use one of the following secure methods:
- A VPN service (Tailscale, WireGuard)
- SSH tunneling
- Cloudflare Tunnels with Zero Trust authentication

**Never expose Remote Vibecode directly to the internet without authentication.**

## FAQ

**Q: Can I access this from another computer?**

A: Yes! Use `rv serve --host 0.0.0.0` and ensure your firewall allows port 7676.

**Q: Is my terminal session secure?**

A: Remote Vibecode runs on localhost by default. When using `--host 0.0.0.0`, ensure you're on a trusted network or use a VPN.

**Q: Can multiple people view the same session?**

A: Yes! Multiple browsers can connect to the same session simultaneously.

**Q: What's the difference between read-only and writable sessions?**

A: Read-only sessions allow viewing only. Writable sessions (created with `-w`) allow web clients to type commands.

**Q: What happens to my session when I close the browser?**

A: Your session continues running in the background. You can reconnect anytime and pick up where you left off.

**Q: Does this work with SSH?**

A: Yes! You can use SSH within a Remote Vibecode session just like a normal terminal.

## License

MIT License - see LICENSE file for details.
