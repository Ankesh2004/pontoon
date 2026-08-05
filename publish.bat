@echo off
setlocal

set /p DOCKER_USER="Enter your Docker Hub username: "
if "%DOCKER_USER%"=="" (
    echo Error: Docker Hub username cannot be empty.
    exit /b 1
)

echo.
echo Building API image...
docker build --target api -t %DOCKER_USER%/pontoon-api:latest .
if %ERRORLEVEL% neq 0 (
    echo Error building API image.
    exit /b 1
)

echo.
echo Building Worker image...
docker build --target worker -t %DOCKER_USER%/pontoon-worker:latest .
if %ERRORLEVEL% neq 0 (
    echo Error building Worker image.
    exit /b 1
)

echo.
echo Pushing API image...
docker push %DOCKER_USER%/pontoon-api:latest

echo.
echo Pushing Worker image...
docker push %DOCKER_USER%/pontoon-worker:latest

echo.
echo ==============================================
echo Success! Both images pushed to Docker Hub!
echo ==============================================
echo On your GCP server, change your docker-compose.yml to:
echo api:
echo   image: %DOCKER_USER%/pontoon-api:latest
echo   ...
echo worker:
echo   image: %DOCKER_USER%/pontoon-worker:latest
echo   ...
echo ==============================================
pause
