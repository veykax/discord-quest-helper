# 🎮 Discord Quest Helper (Go Edition)

> Ультралегкий, быстрый и безопасный эмулятор игр для выполнения квестов Discord (700 Orbs, аватары, эксклюзивные награды).  
> **Вес всего ~5.6 МБ**, мгновенный запуск, 0% лагов, нативный тёмный интерфейс в стиле Discord.

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![OS](https://img.shields.io/badge/OS-Windows%20%7C%20Linux-5865F2?style=flat)
![Size](https://img.shields.io/badge/Binary%20Size-5.6%20MB-brightgreen?style=flat)
![License](https://img.shields.io/badge/License-MIT-green?style=flat)

---

## ⚡ Почему Go-версия лучше

| Параметр | Старая Python-версия | Новая Go-версия |
| :--- | :--- | :--- |
| **Размер файла** | ~150 МБ (с PyInstaller) | **5.6 МБ** (один автономный `.exe`) |
| **Время запуска** | 2–4 секунды | **0.05 секунды** (мгновенно) |
| **Потребление RAM** | ~120 МБ | **~25 МБ** |
| **Плавность GUI** | Лаги, спам painter | **Идеальные 144 FPS** (CSS/SVG аппаратное ускорение) |
| **Поддержка Linux** | Нет (только Windows) | **Полная кроссплатформенность (Windows + Linux)** |
| **Установка зависимостей** | Нужен Python / pip | **НЕ требует установки чего-либо** |

---

## 🎮 Поддерживаемые игры (процессы)

Вместо привязки к постоянно меняющимся названиям квестов приложение сфокусировано на постоянных процессах игр:

* 🧱 **Roblox** $\rightarrow$ `RobloxPlayerBeta.exe` (*Scary Shawarma Kiosk*, *99 Nights in the Forest* и др.)
* 🎯 **Call of Duty: Modern Warfare 4** $\rightarrow$ `cod26-cod.exe` (*700 Orbs, 2XP Token, Calling Card*)
* 🧟 **Resident Evil Sale** $\rightarrow$ `re7.exe`, `re4.exe` (*Resident Evil 7*, *Village*, *RE4*)
* 🧟‍♂️ **Once Human** $\rightarrow$ `once_human.exe` (*Once Human Free Expansion*)
* ⚔️ **Where Winds Meet** $\rightarrow$ `wwm.exe` (*Version 2.3: Friends From Afar*)
* 🛡️ **Arknights: Endfield** $\rightarrow$ `endfield.exe` (*Endfield on Steam Quest*)
* 🦊 **Firefox Browser** $\rightarrow$ `firefox.exe` (*Firefox 15-minute Quest*)
* 🔫 **VALORANT** $\rightarrow$ `VALORANT-Win64-Shipping.exe` (*Avatar Decoration Quest*)
* ✍️ **Свой процесс** $\rightarrow$ поле ввода любого кастомного `.exe` для любых новых игр.

---

## 🚀 Как пользоваться

1. В Discord откройте вкладку **Quests** и нажмите **«Accept Quest»**.
2. Запустите **`discord-quest-helper.exe`** (или `run.bat`).
3. Выберите нужную игру в списке слева (или введите свой `.exe`).
4. Нажмите **«🚀 Запустить (15 мин)»**.
5. Появится небольшое окно игры. Сверните его (не закрывая).
6. Через 15 минут прозвучит звуковой сигнал, и квест в Discord будет готов к получению наград!

---

## 🛠️ Сборка из исходников

Требуется установленный Go:

```bash
# Для Windows (без консольного окна)
go build -ldflags="-s -w -H windowsgui" -o discord-quest-helper.exe

# Для Linux
go build -ldflags="-s -w" -o discord-quest-helper
```

---

## 📄 Лицензия
MIT License.
