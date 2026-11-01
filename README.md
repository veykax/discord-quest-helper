<p align="center">
  <img src="https://assets-global.website-files.com/6257adef93867e50d84d30e2/636e0a6a49cf127bf92de1e2_icon_clyde_blurple_RGB.png" alt="Discord Quest Helper Logo" width="68px"/>
</p>

<h1 align="center"><b>Discord Quest Helper</b></h1>

<p align="center">
  A lightweight, safe, and zero-dependency desktop utility to effortlessly complete <b>Discord Quests</b> (700 Orbs, avatar decorations, and rewards) without running heavy games.
</p>

<p align="center">
  <a href="https://github.com/veykax/discord-quest-helper/releases">Latest Release</a> •
  <a href="#how-to-use">How to Use</a> •
  <a href="#supported-games">Supported Games</a> •
  <a href="#tech-stack">Tech Stack</a>
</p>

<div align="center">
  <a href="https://github.com/veykax/discord-quest-helper/releases">
    <img src="https://img.shields.io/github/v/release/veykax/discord-quest-helper?style=flat-square&color=5865F2&label=Release" alt="Latest Release" />
  </a>
  <img src="https://img.shields.io/badge/Go-1.23-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go Version" />
  <img src="https://img.shields.io/badge/Platform-Windows%20%7C%20Linux-5865F2?style=flat-square&logo=windows&logoColor=white" alt="Platform" />
  <img src="https://img.shields.io/badge/Size-5.6%20MB-23A55A?style=flat-square" alt="Binary Size" />
  <img src="https://img.shields.io/badge/License-MIT-FEE75C?style=flat-square" alt="License" />
</div>

<h5 align="center">
  Leave a star if you find this project helpful! ⭐️
</h5>

---

## About

**Discord Quest Helper** is a standalone, cross-platform utility that emulates official game process signatures and top-level window titles required by Discord Quests. It allows you to claim quest rewards without having to download tens of gigabytes or run resource-heavy games.

> [!IMPORTANT]
> **Safety First:** This tool does **not** read, hook, or modify your Discord client, does not touch Discord memory, and never requests your user tokens or credentials. Emulation operates strictly via standard OS top-level window enumeration and Steam library manifests.

---

## Features

- **Zero Dependencies:** Compiles into a single standalone executable (~5.6 MB). Does not require Go, Python, Node.js, Electron, or external runtimes.
- **Native Standalone Window:** Launches in an isolated, border-clean application window (Chromium App Mode) automatically centered on your display.
- **Auto-Minimized Game Stubs:** Game simulation windows automatically launch minimized into the taskbar so they never interrupt your desktop work or gaming.
- **Process Lifecycle Tracking:** Closing the game simulation window immediately halts the quest timer and resets the state.
- **Steam & Subfolder Verification:** Supports advanced Discord detection rules, including subpath verification (`win64/`) and Steam library manifests (`appmanifest_*.acf` + `steam_appid.txt` for EA SPORTS FC 27).
- **Custom Game Persistence:** Add any `.exe` for future quests; entries persist locally across application restarts.
- **Audio & Toast Notifications:** Built-in chime sound and non-intrusive dark-mode toast alerts when the 15-minute quest completes.

---

## Supported Games

| Game | Simulated Process | Reward | Verification Mode |
| :--- | :--- | :--- | :--- |
| **Roblox** | `RobloxPlayerBeta.exe` | 700 Orbs | Native Executable & Window |
| **Call of Duty: MW4** | `cod26-cod.exe` | 700 Orbs, 2XP Token, Calling Card | Native Executable Signature |
| **Resident Evil Sale** | `re7.exe` | 700 Orbs | Steam Common Directory |
| **Once Human** | `once_human.exe` | 700 Orbs | Directory Hierarchy Match |
| **Where Winds Meet** | `wwm.exe` | 700 Orbs | Top-level Window Match |
| **Arknights: Endfield** | `endfield.exe` | 700 Orbs | Steam AppID & Manifest |
| **Firefox Browser** | `firefox.exe` | 700 Orbs | Native Browser Process |
| **VALORANT** | `win64\VALORANT-Win64-Shipping.exe` | Avatar Decoration | Subfolder Relative Path |
| **EA SPORTS FC 27** | `fc27.exe` | 700 Orbs | Steam SKU `4080220` Integration |
| **Custom Process** | *User-defined (.exe)* | Any Future Quest | Local persistence |

---

## How to Use

1. Open Discord, navigate to **User Settings $\rightarrow$ Quests**, and click **«Accept Quest»**.
2. Download and run **`discord-quest-helper.exe`** from [**Releases**](https://github.com/veykax/discord-quest-helper/releases) (no installation required).
3. Select your target game from the list.
4. Click **«Запустить эмуляцию (15 мин)»**.
5. The game stub will run minimized in your taskbar.
6. Once the 15-minute timer completes, claim your reward directly in Discord!

> [!TIP]
> **For EA SPORTS FC 27:**  
> Discord detects FC 27 through Steam library integration. If Steam is already running, exit Steam completely from your system tray (*Right-click Steam tray icon $\rightarrow$ Exit*) and reopen it once so it registers the newly created game manifest.

---

## Tech Stack

* **Core Runtime:** [Go](https://go.dev/) (Go 1.23)
* **Frontend:** HTML5, CSS3, Vanilla JavaScript (Embedded statically via `embed.FS`)
* **Window Engine:** Native Chromium Standalone App Mode (`--app`, `--window-size`, `--window-position`)
* **Process Emulation:** C# WinForms stub compiler (`csc.exe`) with taskbar minimization & Steam ACF manifest generator

---

## Development & Build

### Prerequisites
* [Go 1.22+](https://go.dev/dl/)

### Build Commands

```bash
# Windows build (GUI subsystem, stripped binary)
go build -ldflags="-s -w -H windowsgui" -o build/discord-quest-helper.exe

# Linux build (stripped binary)
go build -ldflags="-s -w" -o build/discord-quest-helper
```

Or run the included build scripts:
* Windows: `build.bat`
* Linux: `./build.sh`

---

## License

Distributed under the [MIT License](LICENSE).

<p align="center">
  <sub>Developed with human-AI pair programming assistance.</sub>
</p>
