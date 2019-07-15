CC = go

CFLAGS = -v -x

TARGET = ${GOPATH}/bin/gobostick

SRC = $(wildcard *go)

RM = /bin/rm -f

.PHONY: run
run:
	$(CC) run $(SRC)

.PHONY: build
build:
	$(CC) build -o $(TARGET) $(CFLAGS) $(SRC)

.PHONY: clean
clean:
	$(RM) $(TARGET)