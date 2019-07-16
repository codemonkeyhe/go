set name=mod_demo.exe


set GOPRO=E:\gitPro\go

title ===== go Windows =====
::::::::::::::::::::::::::::::::::::

set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

set GOROOT=e:\Go
set GOBIN=%GOPRO%\bin
::set GOPATH=%GOPRO%

echo off
set PATH=%GOROOT%\bin;%PATH%
echo %PATH%
echo on

::set GOPROXY=file://E:/GoOpen/pkg/mod/cache/download
set GOPROXY=file://E:/gitPro/go/go-proxy


echo GOPROXY=%GOPROXY%


del /F/S/Q .\%name% 
go build -o %name%

::del /F/S/Q %GOBIN%\%name% 
::copy /y  %name%  %GOBIN%\%name%  

cmd





