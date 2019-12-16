include .env

.DEFAULT_GOAL := all

## build-ui		:	[DESCRIPTION]
build-ui: build-ui-en build-ui-bg

## build-ui-deps		:	[DESCRIPTION]
build-ui-deps:
	go get -u github.com/urandom/embed/cmd/embed
	npm install --unsafe-perm -g node-gyp webpack-dev-server rimraf webpack typescript @angular/cli @angular/compiler-cli @angular/compiler @angular/core rxjs
	cd rf-ng; npm install

## build-ui-en		:	[DESCRIPTION]
build-ui-en:
	cd rf-ng; ng build --output-path ui/en --base-href /en/ --prod

## build-ui-bg		:	[DESCRIPTION]
build-ui-bg:
	cd rf-ng; ng build --prod --i18n-file=./src/locale/messages.bg.xlf --i18n-locale=bg --i18n-format=xlf --output-path ui/bg --base-href /bg/

## build-ui-devel		:	[DESCRIPTION]
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

## xi18n-ui		:	[DESCRIPTION]
xi18n-ui:
	cd rf-ng; ng xi18n --i18n-locale en --output-path ./src/locale
	cd rf-ng; xliffmerge --profile xliffmerge.json en bg

## generate		:	Generate UI files (local).
generate: build-ui
	go generate .

## build			:	Build ./readeef executable (local).
build: generate
	go build -ldflags="-s -w" ./cmd/readeef

## all			:	Build all files (local).
.PHONY: all
all: build

## docker-build		:	Build the production container.
.PHONY: docker-build
docker-build:
	@docker build -t urandom/readeef:alpine3.10-go1.13 .

## docker-build-dev	:	Build the dev container.
.PHONY: docker-build-dev
docker-build-dev:
	@docker build -t urandom/readeef:alpine3.10-go1.13 -f Dockerfile.dev .

## docker-run		:	Run the production container.
.PHONY: docker-run
docker-run:
	@docker run -ti -p 8080:8080 urandom/readeef:alpine3.10-go1.13

## docker-run-dev		:	Run the dev container.
.PHONY: docker-run-dev
docker-run-dev:
	@docker run -ti -p 8080:8080 urandom/readeef:alpine3.10-go1.13

## docker-ps		:	List running containers.
.PHONY: docker-ps
docker-ps:
	@docker ps --filter name='$(PROJECT_NAME)*'

## docker-logs		:	View containers logs.
##				You can optinally pass an argument with the service name to limit logs
##				logs readeef	: View `readeef` container logs.
##				logs readeef postgres	: View `readeef` and `postgres` containers logs.
.PHONY: docker-logs
docker-logs:
	@docker-compose logs -f $(filter-out $@,$(MAKECMDGOALS))

## compose-up		:	Start up containers.
.PHONY: compose-up
compose-up:
	@echo "Starting up containers for $(PROJECT_NAME)..."
	docker-compose pull
	docker-compose up -d --remove-orphans

## compose-down		:	Stop containers.
.PHONY: compose-down
compose-down: stop

## compose-start		:	Start containers without updating.
.PHONY: compose-start
compose-start:
	@echo "Starting containers for $(PROJECT_NAME) from where you left off..."
	@docker-compose start

## compose-stop		:	Stop containers.
.PHONY: compose-stop
compose-stop:
	@echo "Stopping containers for $(PROJECT_NAME)..."
	@docker-compose stop

## compose-prune		:	Remove containers and their volumes.
##				You can optionally pass an argument with the service name to prune single container
##				prune postgres	: Prune `postgres` container and remove its volumes.
##				prune postgres elasticv7	: Prune `postgres` and `elasticv7` containers and remove their volumes.
.PHONY: compose-prune
compose-prune:
	@echo "Removing containers for $(PROJECT_NAME)..."
	@docker-compose down -v $(filter-out $@,$(MAKECMDGOALS))

## help			:	Print commands help.
.PHONY: help
help : Makefile
	@sed -n 's/^##//p' $<

# https://stackoverflow.com/a/6273809/1826109
%:
	@:


