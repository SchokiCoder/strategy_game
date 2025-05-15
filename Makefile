# SPDX-License-Identifier: GPL-2.0-only
# Copyright (C) 2025  Andy Frank Schoknecht

APP_NAME :=strategy_game

.PHONY: all build run vet

all: vet build

build: $(APP_NAME)

clean:
	rm -f $(APP_NAME)

run: clean all
	./$(APP_NAME)

vet:
	go vet

$(APP_NAME): main.go fieldbiome_string.go
	go build -ldflags "-X 'main.AppName=$(APP_NAME)'"

fieldbiome_string.go: main.go
	go generate

