# 🎮 Discord Quest Helper

> Ультралегкий, быстрый и безопасный эмулятор игр для выполнения квестов Discord (700 Orbs, аватары, эксклюзивные награды).  
> **Вес всего 5.6 МБ**, мгновенный запуск, автономное нативное окно в стиле Discord, автоматическое сворачивание игр.

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![OS](https://img.shields.io/badge/OS-Windows%20%7C%20Linux-5865F2?style=flat)
![Size](https://img.shields.io/badge/Binary%20Size-5.6%20MB-brightgreen?style=flat)
![License](https://img.shields.io/badge/License-MIT-green?style=flat)

---

## ⚡ Особенности

* **Полная автономность:** Один файл `.exe`, не требует установки Go, Python или браузерных расширений.
* **Нативный вид:** Открывается отдельным центрированным окном приложения (Chromium Standalone App Mode).
* **Фоновый запуск:** Окно эмулируемой игры сразу создается свёрнутым в панели задач и не перекрывает экран.
* **100% безопасность:** Не требует токенов аккаунта Discord, не внедряется в память, бан полностью исключен.
* **Кастомные игры:** Любые добавленные процессы сохраняются в памяти и доступны после перезапуска.

---

## 🎮 Поддерживаемые игры (процессы)

* **Roblox** $\rightarrow$ `RobloxPlayerBeta.exe` (*Scary Shawarma Kiosk*, *99 Nights in the Forest* и др.)
* **Call of Duty: Modern Warfare 4** $\rightarrow$ `cod26-cod.exe` (*700 Orbs, 2XP Token, Calling Card*)
* **Resident Evil Sale** $\rightarrow$ `re7.exe` (*Resident Evil 7*, *Village*, *RE4*)
* **Once Human** $\rightarrow$ `once_human.exe` (*Once Human Free Expansion*)
* **Where Winds Meet** $\rightarrow$ `wwm.exe` (*Version 2.3: Friends From Afar*)
* **Arknights: Endfield** $\rightarrow$ `endfield.exe` (*Endfield on Steam Quest*)
* **Firefox Browser** $\rightarrow$ `firefox.exe` (*Firefox 15-minute Quest*)
* **VALORANT** $\rightarrow$ `win64\VALORANT-Win64-Shipping.exe` (*Hush Quest Avatar Decoration*)
* **EA SPORTS FC 27** $\rightarrow$ `fc27.exe` (*Steam SKU 4080220 Integration*)
* **Пользовательский процесс** $\rightarrow$ поле ввода для любого `.exe`.

---

## 🚀 Скачать и запустить

### Для пользователей (без установки Go):
1. Перейдите во вкладку [Releases](https://github.com/veykax/discord-quest-helper/releases).
2. Скачайте файл **`discord-quest-helper.exe`**.
3. Запустите его двойным кликом.

### Инструкция:
1. В приложении Discord откройте вкладку **Quests** и нажмите **«Accept Quest»**.
2. В **Discord Quest Helper** выберите игру из списка.
3. Нажмите кнопку **«Запустить эмуляцию (15 мин)»**.
4. Окно игры автоматически появится свёрнутым в панели задач.
5. Через 15 минут прозвучит звуковой сигнал и появится уведомление — заберите награду в Discord!

---

## 🛠️ Сборка из исходников

Если вы хотите собрать бинарники самостоятельно:

```bash
# Windows
build.bat

# Linux
./build.sh
```

---

## 📄 Лицензия
MIT License.
