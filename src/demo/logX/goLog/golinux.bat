
title ===== go linux =====

::::::::::::::::::::::::::::::::::::
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
::set GOBIN=F:\svn_service\gua_service\go\bin
::set GOPATH=F:\svn_service\gua_service\go
set GOBIN=%GOGUA%\bin
set GOPATH=%GOGUA%
go build -o guaLog  ./
echo.
set UPLOADDIR=D:\SzRz\Upload
del /F/S/Q %UPLOADDIR%\guaLog 
copy /y  guaLog  %UPLOADDIR%\guaLog 
echo.

cmd
