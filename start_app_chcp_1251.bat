@echo off
chcp 1251 >nul
echo ������ ����������...

:: �������� ������� ������ PostgreSQL
echo �������� ������� ������ postgresql-x64-17...
for /f "tokens=4" %%i in ('sc query postgresql-x64-17 ^| findstr "STATE"') do set "service_state=%%i"
if /i "%service_state%"=="RUNNING" (
    echo ������ postgresql-x64-17 ��� ��������.
) else (
    echo ������ postgresql-x64-17 �� ��������. ������� �������...
    net start postgresql-x64-17
    if %ERRORLEVEL% neq 0 (
        echo ������: �� ������� ��������� ������ PostgreSQL. ��������� ��������� ������.
        pause
        exit /b 1
    )
    echo ������ ��������, ���� 5 ������ ��� ������ ����������...
    timeout /t 5 /nobreak >nul
)

:: ��������� TELEGRAM_BOT_TOKEN ��� ������� ������
echo ��������� TELEGRAM_BOT_TOKEN...
set "TELEGRAM_BOT_TOKEN=7964767527:AAHWj-x5ItudN-IyFEFwPaErz8yCa9y4Elw"

:: ������ ������� (Go) � ��������� ����
echo ������ �������...
start "Backend" cmd /k "cd /d C:\00_my_Calendar_2025_prod_v2\backend && echo ������ �������... && go run main.go datasave.go"
timeout /t 3 /nobreak >nul

:: ��������� �������� ���������� ��� PowerShell (��� npm start)
echo ��������� �������� PowerShell...
powershell -Command "Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass -Force" >nul
timeout /t 1 /nobreak >nul

:: ������ ��������� (React) � ��������� ����
echo ������ ���������...
start "Frontend" cmd /k "cd /d C:\00_my_Calendar_2025_prod_v2\frontend && echo ������ ���������... && npm start"
timeout /t 5 /nobreak >nul

:: �������� ��������
echo �������� ��������...
start http://localhost:3000
timeout /t 2 /nobreak >nul

echo ���������� ��������! �������� ��� ���� �������� ��� ���������� ����������.
echo ��������� ������� ����, ������� � ��������� ��� ������. ������� ����� ������� ��� ���������...
pause

:: ��������� ��������� (��� ��������� PostgreSQL)
echo ��������� ����������...
taskkill /FI "WINDOWTITLE eq Telegram Bot" /T /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq Backend" /T /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq Frontend" /T /F >nul 2>&1
echo ���������� �����������. PostgreSQL ���������� �������� ����� ������.
echo ����������� 'net stop postgresql-x64-17' ��� ��������� �������, ���� �����.
pause
exit