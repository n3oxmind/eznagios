PREFIX = /usr/local/bin

all: eznagios

eznagios:
	@go build -o eznagios main.go formatter.go objtype.go attributes.go collection.go colors.go errors.go parser.go
	@echo "Successfully built eznagios"


run:
	go run eznagios

install: eznagios
	@mv eznagios $(PREFIX)/eznagios
	@echo "Installed eznagios to $(PREFIX)/eznagios"

clean:
	@rm -f eznagios
