#!/bin/sh
# Лёгкие установщики: kok-edit-setup.exe (клиенты) и kok-edit-setup-test.exe (тестовая среда)
set -e
cd "$(dirname "$0")"
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.channel=main" -o kok-edit-setup.exe .
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.channel=test" -o kok-edit-setup-test.exe .
ls -la *.exe
