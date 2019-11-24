SRC    = main.go
SRCDIR = ./cmd/server/
EXEC   = innoserver
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
SWAGGERUIPORT=9000
SWAGGERUIPROT=http
SWAGGERBIN=/usr/bin/swagger

SWAGGERDIR   = swagger
API_HOST     = $(SWAGGERUIPROT)://127.0.0.1:$(SWAGGERUIPORT)

ifeq ($(SWAGGERDEF),SWAGGER_JSON)
	SWAGGERFILE=swagger.json
else
	SWAGGERFILE=swagger.yml
endif

all: swag-gen-doc build run-docker run

build:
	@mkdir -p $(BINDIR)
	$(CC) $(BLD) -o $(BINDIR)$(EXEC) $(SRCDIR)$(SRC)

run:
	sleep 4
	@$(BINDIR)$(EXEC)

run-docker:
	docker-compose up&

init-database:
	mysql -h 127.0.0.1 -P $(DBPORT) --protocol=tcp -u $(DBUSR) --password=$(DBPW) -D $(DB) < $(SCHEMA)

demodata:
	mysql -h 127.0.0.1 -P $(DBPORT) --protocol=tcp -u $(DBUSR) --password=$(DBPW) -D $(DB) < $(DEMODATA)

connect:
	mysql -h 127.0.0.1 -P $(DBPORT) --protocol=tcp -u $(DBUSR) --password=$(DBPW) -D $(DB)

swag-doc:
	$(SWAGGERBIN) generate spec -o $(SWAGGERDIR)/$(SWAGGERFILE)
	$(SWAGGERBIN) validate $(SWAGGERDIR)/$(SWAGGERFILE)
	$(SWAGGERBIN) serve $(SWAGGERDIR)/$(SWAGGERFILE)

swag-gen-doc:
	$(SWAGGERBIN) generate spec -o $(SWAGGERDIR)/$(SWAGGERFILE)
	$(SWAGGERBIN) validate $(SWAGGERDIR)/$(SWAGGERFILE)

swagger-ui:
	docker run -p $(SWAGGERUIPORT):8080 --rm --name swagger_ui -e $(SWAGGERDEF)=/mnt/$(SWAGGERFILE) \
		-e API_URL=$(API_HOST)/$(SWAGGERFILE) \
		-v ${PWD}:/mnt \
		-v ${PWD}/$(SWAGGERDIR)/$(SWAGGERFILE):/usr/share/nginx/html/$(SWAGGERFILE) \
		swaggerapi/swagger-ui

shutdown:
	-killall innoserver
	-killall swagger
	-docker-compose down
