SRC    = main.go
SRCDIR = ./cmd/server/
EXEC   = server
BINDIR = ./bin/
CC     = go
BLD    = build
APP_PORT   = 5000
APP_HOST   = 127.0.0.1
APP_PROTOC = http

DBPW     = password
DB       = innovision
DBUSR    = ip
DBPORT   = 3306
SCHEMA   = ./init/schema.sql
DEMODATA = ./init/demodata.sql

SWAGGERDEF=SWAGGER_JSON

SWAGGERDIR   = swagger
API_HOST     = $(APP_PROTOC)://$(APP_HOST):$(APP_PORT)

ifeq ($(SWAGGERDEF),SWAGGER_JSON)
	SWAGGERFILE=swagger.json
else
	SWAGGERFILE=swagger.yml
endif


build:
	@mkdir -p $(BINDIR)
	$(CC) $(BLD) -o $(BINDIR)$(EXEC) $(SRCDIR)$(SRC)

run:
	@$(BINDIR)$(EXEC)

run-docker:
	docker-compose up

init-database:
	mysql -h 127.0.0.1 -P $(DBPORT) --protocol=tcp -u $(DBUSR) --password=$(DBPW) -D $(DB) < $(SCHEMA)

demodata:
	mysql -h 127.0.0.1 -P $(DBPORT) --protocol=tcp -u $(DBUSR) --password=$(DBPW) -D $(DB) < $(DEMODATA)

connect: 
	mysql -h 127.0.0.1 -P $(DBPORT) --protocol=tcp -u $(DBUSR) --password=$(DBPW) -D $(DB)

swag-doc: init
	swagger generate spec -o $(SWAGGERDIR)/$(SWAGGERFILE)
	swagger validate $(SWAGGERDIR)/$(SWAGGERFILE)
	swagger serve $(SWAGGERDIR)/$(SWAGGERFILE)

swag-gen-doc: init
	swagger generate spec -o $(SWAGGERDIR)/$(SWAGGERFILE)
	swagger validate $(SWAGGERDIR)/$(SWAGGERFILE)

swagger-ui: init
	docker run -p 9000:8080 -e $(SWAGGERDEF)=/mnt/$(SWAGGERFILE) -e API_URL=$(API_HOST):$(API_PORT) -v ${PWD}:/mnt swaggerapi/swagger-ui
