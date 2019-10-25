SRC    = main.go
SRCDIR = ./cmd/server/
EXEC   = server
BINDIR = ./bin/
CC     = go
BLD    = build

all: build run

build:
		@mkdir $(BINDIR)
		$(CC) $(BLD) -o $(BINDIR)$(EXEC) $(SRCDIR)$(SRC)

run:
		@$(BINDIR)$(EXEC)
