FROM golang:1.13.3

WORKDIR /go/src/github.com/panghostlin/Proxy/

ADD go.mod .
ADD go.sum .
RUN go mod download

ADD . /go/src/github.com/panghostlin/Proxy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o panghostlin-proxy
RUN chmod +x wait-for-it.sh

ENTRYPOINT [ "/bin/bash", "-c" ]
CMD ["./wait-for-it.sh panghostlin-postgre:54320 --strict --timeout=2" , "--" , "./panghostlin-pictures"]
EXPOSE 80 443