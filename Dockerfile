FROM golang:1.13.3

WORKDIR /go/src/github.com/panghostlin/Proxy/

ADD go.mod .
ADD go.sum .
RUN go mod download

ADD . /go/src/github.com/panghostlin/Proxy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o panghostlin-proxy

ENTRYPOINT ["./panghostlin-proxy"]
EXPOSE 80 443