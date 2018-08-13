@echo off

set BUILD_PATH=.\build\

rem create directory .\build\
if not exist %BUILD_PATH% (
    md %BUILD_PATH%
)

rem Determine if ecoclient.exe exists
if not exist %BUILD_PATH%ecoclient.exe (
    goto ECOCLIENT
)

choice /c yn /m "Whether to overwrite the existing ecoclient.exe file"
if %errorlevel%==2 goto CHECK_ECOBALL 
if %errorlevel%==1 goto ECOCLIENT 

rem build ecoclient.exe
:ECOCLIENT
go build -v -o ecoclient.exe client/client.go 
move ecoclient.exe %BUILD_PATH%

rem Determine if build ecoball.exe exists
:CHECK_ECOBALL
if not exist %BUILD_PATH%ecoball.exe (
    goto ECOBALL
) 

choice /c yn  /m "Whether to overwrite the existing ecoball.exe file"
if %errorlevel%==2 goto END 
if %errorlevel%==1 goto ECOBALL

rem build ecoball.exe
:ECOBALL
cd node 
go build -v 
move node.exe ..\build\ecoball.exe  
cd ../

rem prompt message
:END
echo .....................................
echo .....................................
echo build ok!!!(To exit, press any key or close the form directly)
pause>nul