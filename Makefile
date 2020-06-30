GOLANG ?= /usr/local/go/bin/go
BINARY ?= nurd.out

nurd:
	go build -o $(BINARY) cluster.go config.go db.go main.go

clean:
	rm *.out