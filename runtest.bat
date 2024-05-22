@echo off

SETLOCAL EnableExtensions DisableDelayedExpansion
for /F %%a in ('echo prompt $E ^| cmd') do (
  set "ESC=%%a"
)

where gotestsum >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo.
    echo    %ESC%[1m%ESC%[31mMissing Go Module: %ESC%[0mhttps://github.com/gotestyourself/gotestsum
) else (
    gotestsum -f testdox --watch ./...
)