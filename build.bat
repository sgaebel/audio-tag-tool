@echo off
set "CGO_ENABLED=0"
go build -ldflags="-X 'go.senan.xyz/taglib.binaryPath=C:\Files\Programs-and-Code\go-taglib\taglib.wasm'" -o tagTool.exe main.go
if errorlevel 1 exit /b %errorlevel%
