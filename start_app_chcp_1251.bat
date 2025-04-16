@echo off
chcp 1251 >nul
echo Запуск приложения...

:: Проверка статуса службы PostgreSQL
echo Проверка статуса службы postgresql-x64-17...
for /f "tokens=4" %%i in ('sc query postgresql-x64-17 ^| findstr "STATE"') do set "service_state=%%i"
if /i "%service_state%"=="RUNNING" (
    echo Служба postgresql-x64-17 уже запущена.
) else (
    echo Служба postgresql-x64-17 не запущена. Попытка запуска...
    net start postgresql-x64-17
    if %ERRORLEVEL% neq 0 (
        echo Ошибка: Не удалось запустить службу PostgreSQL. Проверьте настройки службы.
        pause
        exit /b 1
    )
    echo Служба запущена, ждем 5 секунд для полной готовности...
    timeout /t 5 /nobreak >nul
)

:: Установка TELEGRAM_BOT_TOKEN для текущей сессии
echo Установка TELEGRAM_BOT_TOKEN...
set "TELEGRAM_BOT_TOKEN=7964767527:AAHWj-x5ItudN-IyFEFwPaErz8yCa9y4Elw"

:: Запуск бэкенда (Go) в отдельном окне
echo Запуск бэкенда...
start "Backend" cmd /k "cd /d C:\00_my_Calendar_2025_prod_v2\backend && echo Запуск бэкенда... && go run main.go datasave.go"
timeout /t 3 /nobreak >nul

:: Установка политики выполнения для PowerShell (для npm start)
echo Установка политики PowerShell...
powershell -Command "Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass -Force" >nul
timeout /t 1 /nobreak >nul

:: Запуск фронтенда (React) в отдельном окне
echo Запуск фронтенда...
start "Frontend" cmd /k "cd /d C:\00_my_Calendar_2025_prod_v2\frontend && echo Запуск фронтенда... && npm start"
timeout /t 5 /nobreak >nul

:: Открытие браузера
echo Открытие браузера...
start http://localhost:3000
timeout /t 2 /nobreak >nul

echo Приложение запущено! Оставьте это окно открытым для управления процессами.
echo Проверьте консоли бота, бэкенда и фронтенда для ошибок. Нажмите любую клавишу для остановки...
pause

:: Остановка процессов (без остановки PostgreSQL)
echo Остановка приложения...
taskkill /FI "WINDOWTITLE eq Telegram Bot" /T /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq Backend" /T /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq Frontend" /T /F >nul 2>&1
echo Приложение остановлено. PostgreSQL продолжает работать через службу.
echo Используйте 'net stop postgresql-x64-17' для остановки сервера, если нужно.
pause
exit