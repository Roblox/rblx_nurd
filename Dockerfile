FROM golang:latest

RUN mkdir -p /go/src/nurd

WORKDIR /go/src/nurd

COPY . .

RUN go mod download

RUN make install

EXPOSE 8080

CMD ["nurd"]