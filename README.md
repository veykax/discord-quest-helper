# 🎮 Discord Quest Helper

An ultra-lightweight, high-performance, and safe game process emulator designed to complete **Discord Quests** (700 Orbs, avatar decorations, calling cards, 2XP tokens) without downloading or running heavy games.

Built as a standalone, zero-dependency executable (~5.6 MB) with a sleek Discord-themed native dark interface.

---

## 🤖 Built with AI Assistance

This project was architected, written, and optimized in active pair-programming collaboration with **Antigravity (Google DeepMind Advanced Agentic Coding AI)**. 

From exploring Discord's internal detectable game database and Steam SKU reverse-engineering to crafting the minimal-memory Go runtime and responsive UI, every component was developed through human-AI agentic collaboration.

---

## ⚡ Highlights & Key Features

* **Zero Dependencies:** A single standalone `.exe` (5.6 MB). Does **not** require Go, Python, Node.js, Electron, or any browser extensions.
* **Native Standalone Window:** Launches as an isolated application window (Chromium App Mode) centered on your screen, with no address bar, no tabs, and a protected minimum layout size.
* **Minimized Game Simulation:** Emulated game stubs (`Roblox`, `Call of Duty`, `Resident Evil`, etc.) launch directly **minimized into the taskbar** with zero screen interruption.
* **Process Lifecycle Tracking:** If you close the game window or exit the application, the emulation stops immediately and cleans up all temporary stubs.
* **100% Account Safety:** Does **not** inspect or touch your Discord memory, passwords, or User Tokens. Discord detects the simulated processes purely through native Windows API window enumeration (`EnumWindows`) and Steam SKU integration.
* **Custom Process Persistence:** Add any `.exe` for future or unlisted Discord quests; added games are saved locally across restarts.
* **Smart Audio & Toast Alerts:** Gentle Web Audio chime and dark-mode toast notification when your 15-minute quest is ready to claim.

---

## 🎮 Verified Supported Games

| Game | Simulated Process | Supported Quests | Detection Method |
| :--- | :--- | :--- | :--- |
| **Roblox** | `RobloxPlayerBeta.exe` | *Scary Shawarma Kiosk*, *99 Nights in the Forest*, etc. | Native Window & Executable |
| **Call of Duty: MW4** | `cod26-cod.exe` | *700 Orbs, 2XP Token, Calling Card* | Native Executable & Subsystem |
| **Resident Evil Sale** | `re7.exe` | *Resident Evil 7: Biohazard*, *Village*, *RE4* | Steam Common & Executable |
| **Once Human** | `once_human.exe` | *Once Human Free Expansion Quest* | Directory Match (`Once Human`) |
| **Where Winds Meet** | `wwm.exe` | *Version 2.3: Friends From Afar* | Native Window & Executable |
| **Arknights: Endfield** | `endfield.exe` | *Endfield is now on Steam Quest* | Steam AppID & Manifest |
| **Firefox Browser** | `firefox.exe` | *Firefox 15-Minute Browser Quest* | Native Window & Process |
| **VALORANT** | `win64\VALORANT-Win64-Shipping.exe` | *VALORANT Hush Quest (Avatar Decoration)* | Subfolder Match (`win64/`) |
| **EA SPORTS FC 27** | `fc27.exe` | *EA SPORTS FC 27 Quest (700 Orbs)* | Steam SKU 4080220 Manifest |
| **Custom Process** | *Your choice (.exe)* | *Any future or community game* | User-defined entry |

---

## 🚀 Quick Start (No Installation)

1. Download **`discord-quest-helper.exe`** from [GitHub Releases](https://github.com/veykax/discord-quest-helper/releases).
2. Open your Discord desktop app, go to **User Settings $\rightarrow$ Quests**, and click **«Accept Quest»**.
3. Run `discord-quest-helper.exe`.
4. Select the corresponding game from the left panel and click **«Запустить эмуляцию (15 мин)»**.
5. The game will automatically run minimized in the taskbar.
6. Once the 15-minute countdown finishes, claim your rewards directly in Discord!

> **Note for EA SPORTS FC 27:**  
> Discord detects FC 27 exclusively via Steam library integration. If Steam is currently running, exit Steam completely from your taskbar tray and restart it once so Steam reads the newly created game manifest.

---

## 🛠️ Building from Source

### Prerequisites
* [Go 1.22+](https://go.dev/dl/)

### Build Commands

```bash
# Clone the repository
git clone https://github.com/veykax/discord-quest-helper.git
cd discord-quest-helper

# Build Windows executable (GUI mode, no command prompt)
go build -ldflags="-s -w -H windowsgui" -o build/discord-quest-helper.exe

# Build Linux binary
go build -ldflags="-s -w" -o build/discord-quest-helper
```

Or run the included build scripts:
* Windows: `build.bat`
* Linux: `./build.sh`

---

## 🛡️ Privacy & Safety Disclaimer

* **No Self-botting:** This tool never communicates with Discord's private REST endpoints and never requests your user authorization tokens.
* **No Memory Injection:** Does not hook or modify Discord or third-party game binaries.
* **Compliant:** Operates solely as an OS-level top-level window stub that registers with Discord's native game activity scanner.

---

## 📄 License

Distributed under the [MIT License](LICENSE).
