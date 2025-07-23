BIN_DIR := bin
APP_NAME := docker-publish

.PHONY: all build clean push

all: build push

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) .

clean:
	rm -rf $(BIN_DIR)

push:
	git add $(BIN_DIR)/$(APP_NAME)
	git commit -m "Update binary $(APP_NAME)"
	git push origin main
