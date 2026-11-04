@echo off
chcp 65001 >nul
cd /d "%~dp0"
if not exist "build" mkdir "build"
copy /Y games.json internal\config\default_games.json >nul
echo Сборка Discord Quest Helper для Windows...
set CGO_ENABLED=0
go build -ldflags="-s -w -H windowsgui" -o build\discord-quest-helper.exe
if %errorlevel% equ 0 (
    echo [УСПЕХ] Файл build\discord-quest-helper.exe успешно собран!
) else (
    echo [ОШИБКА] Сборка не удалась.
    pause
)
