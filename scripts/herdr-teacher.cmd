@echo off
setlocal
rem herdr-teacher - runs the herdr-soho engine with this skill's roles and config.

node -e "process.exit(Number(process.versions.node.split('.')[0]) >= 20 ? 0 : 1)" >nul 2>nul
if errorlevel 1 goto try_bun
node "%~dp0herdr-teacher.mjs" %*
exit /b %errorlevel%

:try_bun
where bun >nul 2>nul
if errorlevel 1 goto no_runtime
bun "%~dp0herdr-teacher.mjs" %*
exit /b %errorlevel%

:no_runtime
echo herdr-teacher: Node.js 20+ or Bun is required 1>&2
exit /b 2
