# build stage
FROM golang:1.12.3  AS build-env

# Set our workdir to our current service in the gopath
WORKDIR /go/src/uploadsrv/
# Copy the current code into our workdir
COPY . .
ENV GOPATH /go/
ENV GO111MODULE on
RUN go mod init /go/src/uploadsrv/
RUN go build -o uploadsrv main.go

# final stage
FROM ubuntu:bionic

RUN apt-get update && apt-get install -y openssl

WORKDIR /app

COPY --from=build-env /go/src/uploadsrv/public/ /app/public/
RUN mkdir /app/data
RUN mkdir /app/db


COPY --from=build-env /go/src/uploadsrv/uploadsrv /app/

EXPOSE 1323

CMD ./uploadsrv
