set name=gua_test_d

title ===== go linux =====
::::::::::::::::::::::::::::::::::::
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
::set GOBIN=F:\svn_service\gua_service\go\bin
::set GOPATH=F:\svn_service\gua_service\go
set GOBIN=%GOGUA%\bin
set GOPATH=%GOGUA%

go build -o %name%

del /F/S/Q %GOBIN%\%name% 
copy /y  %name%  %GOBIN%\%name%  

cmd