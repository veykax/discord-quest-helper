@echo off
chcp 65001 >nul
cd /d "%~dp0"
if not exist "build\discord-quest-helper.exe" (
    echo [INFO] Исполняемый файл не найден. Запуск сборки...
    call build.bat
)
if exist "build\discord-quest-helper.exe" (
    start "" "build\discord-quest-helper.exe"
)
