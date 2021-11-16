
# docker build -t waggle/sheets2json .
# docker run -ti -p 8000:8000 -v ${PWD}/config.yaml:/config.yaml waggle/sheets2json
FROM golang:1.17-alpine

WORKDIR /go/src/app

COPY vendor ./vendor
COPY go.mod go.sum main.go .

RUN go get -d -v ./...
RUN go install -v ./...


CMD ["/go/bin/sheets2json"]