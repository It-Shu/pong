# PONG

```text
██████╗  ██████╗ ███╗   ██╗ ██████╗
██╔══██╗██╔═══██╗████╗  ██║██╔════╝
██████╔╝██║   ██║██╔██╗ ██║██║  ███╗
██╔═══╝ ██║   ██║██║╚██╗██║██║   ██║
██║     ╚██████╔╝██║ ╚████║╚██████╔╝
╚═╝      ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝
```

Terminal Pong built with Go, Bubble Tea, Lip Gloss, and Bubbles.

Fast rounds, AI difficulty levels, animated win and lose states, keyboard-first gameplay, and a clean terminal UI.

Creator: [`github.com/It-Shu`](https://github.com/It-Shu)

## Install In One Command

Any system with Node.js and npm:

```bash
npm install -g @shsergei/pong-terminal
```

After install, run:

```bash
pong-terminal
```

## What Is Inside

- Single-player Pong against AI
- `Easy`, `Medium`, and `Hard` difficulty modes
- Pause, restart, and return-to-menu flow
- Terminal-native UI effects for goals, victory, and defeat
- Cross-platform Go project for Windows, macOS, and Linux

## Controls

- `up` move up
- `down` move down
- `space` start or resume
- `p` pause
- `r` restart
- `m` back to menu
- `q` quit

## Run It Locally

If you have Go installed:

```bash
go run .
```

## Build It

Linux / macOS:

```bash
go build -o pong-terminal .
./pong-terminal
```

Windows:

```powershell
go build -o pong-terminal.exe .
.\pong-terminal.exe
```

## Tech

- Go
- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- [Bubbles](https://github.com/charmbracelet/bubbles)

## Notes

- The game is designed for a modern terminal with Unicode support.
- Best experience is in Windows Terminal, iTerm2, or a modern Linux terminal.
- If the window is too small, enlarge the terminal a bit for a cleaner layout.

## License

This project is proprietary.

You may install, run, and play the game for personal, non-commercial use.
You may not modify, redistribute, or use the source code or packaged files
without prior written permission.

See [LICENSE](LICENSE) for the full terms.
