#get a base image
FROM golang:1.23.0-alpine3.20

WORKDIR /go/src/app
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./*.go .

RUN go get -d -v
RUN go build -v

EXPOSE 3000

CMD ["./cat_facts"]
