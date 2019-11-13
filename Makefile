SRC    = main.go
SRCDIR = ./cmd/server/
EXEC   = server
BINDIR = ./bin/
CC     = go
BLD    = build

DBPW   = password
DB     = innovision
DBUSR  = ip
DBPORT = 3306
SCHEMA = ./init/schema.sql
DEMODATA = ./init/demodata.sql

all: build run

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
