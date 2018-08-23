@echo off
rem Check go version
set MIN_VERSION=1.9.0
set VERSION_EXT=.0
for /f "tokens=* delims=" %%i in ('go version') do set GO_VERSION=%%i
set "GO_VERSION=%GO_VERSION:~13,-14%"

set num=0
set str=%GO_VERSION%
set ch1=.
if not "%str%"=="" (
    set "ch1=%str:~0,1%"
    if "%ch1%"=="." ( set /a num+=1 )
    set "str=%str:~1%"
)

if %num% equ 1 ( set "GO_VERSION=%GO_VERSION%%VERSION_EXT%" ) 

set "version1=%GO_VERSION:.=0%"
set "version2=%MIN_VERSION:.=0%"
if %version1% lss %version2% (
    echo The current version is %GO_VERSION%
    echo The smallest version is %MIN_VERSION%, Please upgrade the go version
    goto FAILED
)

rem check CGO_ENABLED
go env | findstr "CGO_ENABLED" > tmp.txt
for /f "tokens=* delims=" %%i in (tmp.txt) do set GO_ENABLED=%%i
del tmp.txt
set "GO_ENABLED=%GO_ENABLED:~-1%"
if %GO_ENABLED% neq 1 (
    echo This project used CGO, so set the CGO_ENABLED=1
    goto FAILED
)

rem create directory .\build\
set BUILD_PATH=.\build\
if not exist %BUILD_PATH% (
    md %BUILD_PATH%
)

rem Determine if ecoclient.exe exists
if not exist %BUILD_PATH%ecoclient.exe (
    goto ECOCLIENT
)

choice /c yn /m "Whether to overwrite the existing ecoclient.exe file"
if %errorlevel%==2 goto CHECK_ECOWALLET 
if %errorlevel%==1 goto ECOCLIENT 

rem build ecoclient.exe
:ECOCLIENT
go build -v -o ecoclient.exe client/client.go 
if %ERRORLEVEL% neq 0 (
    echo build ecoclient.exe failed!!!!
    goto CHECK_ECOWALLET 
)
move ecoclient.exe %BUILD_PATH% > nul
echo Build ecoclient.exe success!!!!

rem Determine if ecowallet.exe exists
:CHECK_ECOWALLET
if not exist %BUILD_PATH%ecowallet.exe (
    goto ECOWALLET
)

choice /c yn /m "Whether to overwrite the existing ecowallet.exe file"
if %errorlevel%==2 goto CHECK_ECOBALL 
if %errorlevel%==1 goto ECOWALLET 

rem build ecowallet.exe
:ECOWALLET
go build -v -o ecowallet.exe walletserver/main.go 
if %ERRORLEVEL% neq 0 (
    echo build ecowallet.exe failed!!!!
    goto CHECK_ECOBALL 
)
move ecowallet.exe %BUILD_PATH% > nul
echo Build ecowallet.exe success!!!!

rem Determine if build ecoball.exe exists
:CHECK_ECOBALL
if not exist %BUILD_PATH%ecoball.exe (
    goto ECOBALL
) 

choice /c yn  /m "Whether to overwrite the existing ecoball.exe file"
if %errorlevel%==2 goto FAILED 
if %errorlevel%==1 goto ECOBALL

rem build ecoball.exe
:ECOBALL
cd node 
go build -v
if %ERRORLEVEL% neq 0 (
    echo build ecoball.exe failed!!!!
    cd ../
    goto FAILED
) 
move node.exe ..\build\ecoball.exe > nul
cd ../
echo Build ecoball.exe success!!!!

:FAILED
pause>nul