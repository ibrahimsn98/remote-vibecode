# Remote Vibecode

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
- **Auto-discovery** - Sessions appear automatically as you create them
- **Multi-session Support** - Manage multiple sessions with a sidebar navigation
- **Responsive Design** - Works on desktop, tablet, and mobile
- **UTF-8 Support** - Full Unicode character support
- **Keepalive Connections** - Stable WebSocket connections with ping/pong

## Prerequisites

- macOS with Homebrew

## Installation

### Add the Tap

```bash
brew tap yourusername/tap
```

### Install

```bash
brew install remote-vibecode
```

### Start the Service

```bash
brew services start remote-vibecode
```

### Verify Installation

```bash
brew services list | grep remote-vibecode
```

You should see `remote-vibecode` listed as `started`.

The service runs on http://localhost:8080 by default.

### Uninstall

```bash
brew services stop remote-vibecode
brew uninstall remote-vibecode
```

## Quick Start

### Step 1: Start a Session

```bash
vibecode start my-project
```

### Step 2: Open in Browser

Navigate to http://localhost:8080 and click on your session to connect.

That's it! You now have a browser-based terminal connected to your session.

## CLI Reference

### Start a New Session

```bash
vibecode start <session-name>
```

Creates and starts a new named session.

```bash
vibecode start frontend
vibecode start backend
vibecode start database
```

### List Sessions

```bash
vibecode list
```

Shows all active sessions.

### Stop a Session

```bash
vibecode stop <session-name>
```

Stops and removes the specified session.

### Attach to a Session

```bash
vibecode attach <session-name>
```

Attaches your terminal to an existing session (useful for direct terminal access).

## Usage Examples

### Multiple Sessions for Different Projects

```bash
# Start sessions for different parts of your stack
vibecode start frontend
vibecode start backend
vibecode start database

# Open http://localhost:8080
# Switch between sessions using the sidebar
```

### Remote Pair Programming

```bash
# Start your session
vibecode start pair-session

# Share your screen with your teammate
# They can watch you work or you can grant them control
```

### Workshop Presentation

```bash
# Create a session for your workshop
vibecode start workshop-demo

# Participants can follow along on the big screen
# Or they can connect to their own sessions
```

## Configuration

Remote Vibecode can be configured via environment variables.

### Port

Change the default port (8080):

```bash
export PORT=9000
brew services restart remote-vibecode
```

### Host

Change the bind address:

```bash
export HOST=192.168.1.100
brew services restart remote-vibecode
```

### Service Configuration

For persistent configuration, create or edit `~/Library/LaunchAgents/homebrew.mxcl.remote-vibecode.plist`:

```xml
<key>EnvironmentVariables</key>
<dict>
    <key>PORT</key>
    <string>9000</string>
    <key>HOST</key>
    <string>0.0.0.0</string>
</dict>
```

## Troubleshooting

### Service Not Starting

If the service fails to start:

```bash
# Check service status
brew services list

# View logs
log show --predicate 'process == "remote-vibecode"' --last 1h

# Try starting manually for more error info
remote-vibecode-server
```

### Session Not Appearing

If your session doesn't show up in the browser:

- Verify the session is running: `vibecode list`
- Check the web console for WebSocket errors
- Ensure the service is running: `brew services list | grep remote-vibecode`

### Browser Connection Issues

If you can't connect from your browser:

- Check that http://localhost:8080 is accessible
- Try a different browser
- Check your firewall settings
- Verify the service is running on the expected port

### Permission Issues

If you see permission errors:

```bash
# Ensure homebrew services have correct permissions
sudo brew services restart remote-vibecode
```

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/remote-vibecode.git
cd remote-vibecode

# Build the service
cd service && go build -o remote-vibecode-server ./cmd/server

# Run locally
./remote-vibecode-server
```

### Project Structure

```
remote-vibecode/
├── service/              # Main Go service
│   ├── cmd/server/       # Service entry point
│   ├── internal/         # Internal packages
│   │   ├── api/          # REST API handlers
│   │   ├── gotty/        # Gotty protocol implementation
│   │   ├── tmux/         # Session management
│   │   └── ws/           # WebSocket handlers
│   └── web/              # Static web dashboard
├── scripts/              # Utility scripts
└── README.md             # This file
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

## Screenshots

<!-- TODO: Add demo GIF or screenshots here -->

## FAQ

**Q: Can I access this from another computer?**

A: Yes! Set the `HOST` environment variable to your local IP address and ensure port 8080 is open on your firewall.

**Q: Is my terminal session secure?**

A: Remote Vibecode runs on localhost by default. When exposing it to a network, consider using a reverse proxy with SSL/TLS termination.

**Q: Can multiple people view the same session?**

A: Yes! Multiple browsers can connect to the same session simultaneously.

**Q: What happens to my session when I close the browser?**

A: Your session continues running in the background. You can reconnect anytime and pick up where you left off.

**Q: Does this work with SSH?**

A: Yes! You can use SSH within a Remote Vibecode session just like a normal terminal.

## License

MIT License - see LICENSE file for details.
