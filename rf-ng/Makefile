build: build-en build-bg

build-en:
	ng build -op ui/en --base-href /en/ --prod

build-bg:
	ng build --prod --i18nFile=./src/locale/messages.bg.xlf --locale=bg --i18nFormat=xlf -op ui/bg --base-href /bg/

build-devel: build-devel-en build-devel-bg

build-devel-en:
	ng build -op ui/en --base-href /en/

build-devel-bg:
	ng build --i18nFile=./src/locale/messages.bg.xlf --locale=bg --i18nFormat=xlf -op ui/bg --base-href /bg/

build-devel-watch:
	ng build -op ui/en --base-href /en/ --watch

xi18n:
	ng xi18n --locale en --outputPath ./src/locale
	xliffmerge --profile xliffmerge.json en bg
