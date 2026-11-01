# 🎮 Discord Quest Helper

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version" />
  <img src="https://img.shields.io/badge/Platform-Windows%20%7C%20Linux-5865F2?style=for-the-badge&logo=windows&logoColor=white" alt="Platform" />
  <img src="https://img.shields.io/badge/Binary%20Size-5.6%20MB-23A55A?style=for-the-badge" alt="Binary Size" />
  <img src="https://img.shields.io/badge/License-MIT-FEE75C?style=for-the-badge&labelColor=313338" alt="License" />
</p>

<p align="center">
  A lightweight, safe, and zero-dependency utility to effortlessly complete <b>Discord Quests</b> (700 Orbs, avatars, badges, and rewards) without running or installing heavy games.
</p>

---

## ✨ Features

* 🚀 **Zero Dependencies:** Single standalone executable (~5.6 MB). No Python, Go, Node.js, or Electron required.
* 🖥️ **Native Standalone Window:** Launches in an isolated Chromium App window, automatically centered on your screen with a protected layout.
* 🔕 **Auto-Minimized Game Stubs:** Emulated game processes immediately start minimized in the taskbar so they never interrupt your screen or gameplay.
* ⏱️ **Auto-Sync & Lifecycle Tracking:** Closing the game simulation window automatically stops the timer.
* 🛡️ **100% Safe:** No memory injection, no user tokens, no private API requests. Works purely through legitimate Windows window enumeration and Steam manifest integration.
* ✍️ **Custom Process Support:** Add any custom `.exe` for future quests; saved locally in memory across restarts.
* 🔔 **Sound Alerts & Toast Notifications:** Built-in Web Audio chime and dark-mode toast alerts when the 15-minute quest completes.

---

## 🎮 Verified Games & Quests

| Game | Simulated Process | Quest Reward | Detection |
| :--- | :--- | :--- | :--- |
| **Roblox** | `RobloxPlayerBeta.exe` | 700 Orbs | Native Executable & Window |
| **Call of Duty: MW4** | `cod26-cod.exe` | 700 Orbs, 2XP Token, Calling Card | Native Executable |
| **Resident Evil Sale** | `re7.exe` | 700 Orbs | Steam Common & Executable |
| **Once Human** | `once_human.exe` | 700 Orbs | Directory Match (`Once Human`) |
| **Where Winds Meet** | `wwm.exe` | 700 Orbs | Native Executable |
| **Arknights: Endfield** | `endfield.exe` | 700 Orbs | Steam AppID & Manifest |
| **Firefox Browser** | `firefox.exe` | 700 Orbs | Native Window & Process |
| **VALORANT** | `win64\VALORANT-Win64-Shipping.exe` | Avatar Decoration | Subfolder Match (`win64/`) |
| **EA SPORTS FC 27** | `fc27.exe` | 700 Orbs | Steam SKU 4080220 Manifest |
| **Custom Game** | *Custom (.exe)* | Any Quest | User-defined |

---

## 📥 Download & Usage

1. Go to the [**Releases**](https://github.com/veykax/discord-quest-helper/releases) page and download `discord-quest-helper.exe`.
2. Open Discord, navigate to **User Settings $\rightarrow$ Quests**, and click **Accept Quest**.
3. Launch `discord-quest-helper.exe` (run directly, no installation needed).
4. Select the corresponding game and click **Запустить эмуляцию (15 мин)**.
5. The game will run minimized in the taskbar. After 15 minutes, claim your reward in Discord!

> 💡 **Tip for EA SPORTS FC 27:**  
> If Steam is running in the background, fully exit Steam from the tray (*Right Click $\rightarrow$ Exit*) and relaunch it once so Steam detects the generated game manifest.

---

## 🛠️ Building from Source

```bash
# Windows (no console window)
build.bat

# Linux
./build.sh
```

Or manually:
```bash
go build -ldflags="-s -w -H windowsgui" -o build/discord-quest-helper.exe
```

---

## 📄 License

Distributed under the [MIT License](LICENSE).

<p align="center">
  <sub>Built with AI assistance.</sub>
</p>
