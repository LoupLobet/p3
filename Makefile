all: p3

p3: p3.go
	go build p3.go

install: p3
	mkdir -p /usr/local/bin/
	cp p3 /usr/local/bin/p3

uninstall:
	rm -f /usr/local/bin/p3

.PHONY: p3 install uninstall