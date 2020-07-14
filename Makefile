ifndef $(GOLANG)
    GOLANG=$(shell which go)
    export GOLANG
endif
BINARY ?= nurd
BINDIR ?= $(DESTDIR)/usr/local

nurd:
	$(GOLANG) build -o $(BINARY) cluster.go config.go db.go main.go

install:
	install -m 755 $(BINARY) $(BINDIR)

clean:
	rm -f $(BINARY)