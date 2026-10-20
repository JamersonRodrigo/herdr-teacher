@echo off
setlocal

if not defined HERDR_TEACHER_SKILL_DIR (
  for %%I in ("%~dp0..") do set "HERDR_TEACHER_SKILL_DIR=%%~fI"
)

if defined HERDR_TEACHER_BIN goto have_custom_bin
goto check_local_bin

:have_custom_bin
if not exist "%HERDR_TEACHER_BIN%" goto bad_bin
"%HERDR_TEACHER_BIN%" %*
exit /b %errorlevel%

:check_local_bin
set "HERDR_TEACHER_LOCAL_BIN=%HERDR_TEACHER_SKILL_DIR%\bin\herdr-teacher.exe"

where go >nul 2>nul
if %errorlevel% equ 0 (
  go -C "%HERDR_TEACHER_SKILL_DIR%" build -o bin\herdr-teacher.exe ./cmd/herdr-teacher
  if errorlevel 1 exit /b 2
) else (
  if not exist "%HERDR_TEACHER_LOCAL_BIN%" goto no_go
)

"%HERDR_TEACHER_LOCAL_BIN%" %*
exit /b %errorlevel%

:bad_bin
1>&2 echo herdr-teacher: HERDR_TEACHER_BIN is not executable: %HERDR_TEACHER_BIN%
exit /b 2

:no_go
1>&2 echo herdr-teacher: Go 1.25+ is required to build bin/herdr-teacher (or set HERDR_TEACHER_BIN)
exit /b 2
