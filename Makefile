ifndef $(GOLANG)
    GOLANG=$(shell which go)
    export GOLANG
endif
BINARY ?= nurd

nurd:
	$(GOLANG) build -o $(BINARY) cluster.go config.go db.go main.go

clean:
	rm -f $(BINARY)