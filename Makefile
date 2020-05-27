PREFIX = /usr/local/bin

all: gonag

gonag:
	@go build -o gonag main.go formatter.go objtype.go attributes.go collection.go colors.go errors.go parser.go
	@echo "Successfully built gonag"


run:
	go run gonag

install: gonag
	@mv gonag $(PREFIX)/gonag
	@echo "Installed gonag to $(PREFIX)/gonag"

clean:
	@rm -f gonag
