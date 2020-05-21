PREFIX = /usr/local/bin
obj = eznag


all: eznag

eznag: $(obj)
	go build -o eznag main.go formater.go objstype.go attributes.go collection.go colors.go errors.go parser.go

run:
	go run eznag

clean: rm -f eznag
