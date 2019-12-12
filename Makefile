include .docker/docker.mk

.PHONY: all

.DEFAULT_GOAL := all

## build-ui	:	[DESCRIPTION]
build-ui: build-ui-en build-ui-bg

## build-ui-en	:	[DESCRIPTION]
build-ui-en:
	cd rf-ng; ng build --output-path ui/en --base-href /en/ --prod

## build-ui-bg	:	[DESCRIPTION]
build-ui-bg:
	cd rf-ng; ng build --prod --i18n-file=./src/locale/messages.bg.xlf --i18n-locale=bg --i18n-format=xlf --output-path ui/bg --base-href /bg/

## build-ui-devel	:	[DESCRIPTION]
build-ui-devel: build-ui-devel-en build-ui-devel-bg

## build-ui-devel-en	:	[DESCRIPTION]
build-ui-devel-en:
	cd rf-ng; ng build --output-path ui/en --base-href /en/

## build-ui-devel-bg	:	[DESCRIPTION]
build-ui-devel-bg:
	cd rf-ng; ng build --i18n-file=./src/locale/messages.bg.xlf --i18n-locale=bg --i18n-format=xlf --output-path ui/bg --base-href /bg/

## build-ui-devel-watch	:	[DESCRIPTION]
build-ui-devel-watch:
	cd rf-ng; ng build --output-path ui/en --base-href /en/ --watch

## xi18n-ui	:	[DESCRIPTION]
xi18n-ui:
	cd rf-ng; ng xi18n --i18n-locale en --output-path ./src/locale
	cd rf-ng; xliffmerge --profile xliffmerge.json en bg

## generate	:	Generate UI files (local).
generate: build-ui
	go generate .

## build	:	Build ./readeef executable (local).
build: generate
	go build -ldflags="-s -w" ./cmd/readeef

## all	:	Build all files (local).
all: build

