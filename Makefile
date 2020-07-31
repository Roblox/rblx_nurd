ifndef $(GOLANG)
    GOLANG=$(shell which go)
    export GOLANG
endif
BINARY ?= nurd
BINDIR ?= $(DESTDIR)/usr/local/bin
SYSCONFDIR ?= $(DESTDIR)/etc/nurd

build:
	$(GOLANG) build -o $(BINARY) cluster.go config.go db.go main.go

install:
	mkdir -p $(SYSCONFDIR)

	if [ ! -f "$(SYSCONFDIR)/config.json" ]; then \
		install -m 644 etc/nurd/config.json $(SYSCONFDIR)/config.json; \
	fi


	$(GOLANG) build -o $(BINARY) cluster.go config.go db.go main.go
	install -m 755 $(BINARY) $(BINDIR)

test:
	$(GOLANG) test -count=1 -v ./...

clean:
	rm -f $(BINARY)