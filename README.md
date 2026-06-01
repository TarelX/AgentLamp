# AgentLamp

> 程序员的过街信号 / The Pedestrian Signal for Coders

跨平台 AI Agent 状态灯桌面端 — 用一盏物理交通信号灯告诉你 Claude Code、Codex、Cursor 现在到底在干什么。纯软件、零硬件、跨平台、永久免费、MIT 开源。

A cross-platform desktop status light for AI coding agents. Watch one tiny traffic light and instantly know whether your agent is thinking, waiting on you, or has hit an error. Pure software, zero hardware, MIT licensed, free forever.

---

## Status

`v0.1` early development. Not yet ready for use.

| Platform | Status |
|---|---|
| Windows 10/11 | dev (primary) |
| macOS 12+     | planned |
| Linux         | planned |

| Agent       | v0.1 | v0.2 | v0.3+ |
|---|:---:|:---:|:---:|
| Claude Code | yes  |      |       |
| Cursor      | yes  |      |       |
| Codex       |      | yes  |       |
| Cline       |      |      | yes   |
| Aider       |      |      | yes   |
| Continue    |      |      | yes   |
| Windsurf    |      |      | yes   |

---

## Status semantics

| Light            | Meaning                                                       | Sound default |
|---|---|---|
| green steady     | idle / last task finished cleanly                              | mute |
| yellow slow blink| running (thinking / tool call / file edit)                     | mute |
| yellow fast blink| **waiting on user** (permission / input)                       | W1 (3 x 800Hz square) |
| red steady       | task failed                                                    | E1 (1 x 400Hz square long) |
| red fast blink   | API failure (rate limit / token / overload)                    | E1 |
| gray             | agent disabled or no signal                                    | mute |

Aggregation priority: `red > yellow(waiting) > yellow(running) > green > gray`.

---

## Tech stack

- **Framework**: [Wails v3](https://v3.wails.io/) (Go + Web)
- **Frontend**: React 18 + TypeScript + Vite
- **Backend**: Go 1.25 (`fsnotify`, `mattn/go-sqlite3`)
- **Audio**: Web Audio API live synthesis (zero audio files)
- **Build**: GitHub Actions (macOS / Windows / Linux)

Bundle target: 10 - 25 MB. Idle memory target: 50 - 150 MB.

---

## Project layout

```
AgentLamp/
├── main.go                        # Wails app entry
├── greetservice.go                # placeholder (replaced in Day 3-4)
├── go.mod / go.sum
├── Taskfile.yml                   # `wails3 task ...` orchestration
├── build/                         # platform packaging assets
│   ├── config.yml
│   ├── windows/  darwin/  linux/
├── frontend/
│   ├── index.html
│   ├── package.json / vite.config.ts / tsconfig.json
│   └── src/                       # App.tsx, components/, audio/, stores/, ...
├── backend/                       # (created during Day 3-4)
│   ├── adapters/     claude.go, cursor.go, ...
│   ├── aggregator/
│   ├── notify/
│   ├── store/
│   └── installer/                 # one-click hook installer
├── ui/                            # 10 design prototypes (HTML)
└── AI-Agent-StatusLight-prd.md    # full product / tech / roadmap doc
```

---

## Development

```bash
# Install Wails v3 CLI
go install github.com/wailsapp/wails/v3/cmd/wails3@latest

# Verify environment
wails3 doctor

# Run dev (hot reload)
wails3 dev

# Build
wails3 build
```

---

## Roadmap

- **v0.1** Day 1-7 : floating window + menubar + Claude Code & Cursor adapters + W1/E1 sounds
- **v0.2** Day 8-12: animations + themes + one-click hook installer + Codex
- **v0.3** Day 13-17: Cline / Windsurf / Aider + work-time stats + cross-platform packaging
- **v0.4** Day 18-19: README polish + 2-min demo video
- **v1.0** Day 20-21: Show HN + r/ClaudeAI + r/cursor + V2EX launch

Full PRD: [`AI-Agent-StatusLight-prd.md`](./AI-Agent-StatusLight-prd.md)

---

## License

MIT - see [`LICENSE`](./LICENSE). Free forever, no paid tier.
