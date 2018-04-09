.PHONY: all

.DEFAULT_GOAL := all

build-ui: build-ui-en build-ui-bg

build-ui-en:
	cd rf-ng; ng build -op ui/en --base-href /en/ --prod

build-ui-bg:
	cd rf-ng; ng build --prod --i18nFile=./src/locale/messages.bg.xlf --locale=bg --i18nFormat=xlf -op ui/bg --base-href /bg/

build-ui-devel: build-ui-devel-en build-ui-devel-bg

build-ui-devel-en:
	cd rf-ng; ng build -op ui/en --base-href /en/

build-ui-devel-bg:
	cd rf-ng; ng build --i18nFile=./src/locale/messages.bg.xlf --locale=bg --i18nFormat=xlf -op ui/bg --base-href /bg/

build-ui-devel-watch:
	cd rf-ng; ng build -op ui/en --base-href /en/ --watch

xi18n-ui:
	cd rf-ng; ng xi18n --locale en --outputPath ./src/locale
	cd rf-ng; xliffmerge --profile xliffmerge.json en bg

generate: build-ui
	go generate .

build: generate
	go build -ldflags="-s -w" ./cmd/readeef

all: build
