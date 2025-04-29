FROM golang:alpine AS builder

# set go specific env vars
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64


ARG PATH_CONFIG="/go-pandoc/config/app.conf"

RUN mkdir /config
ADD app.conf /config/app.conf
ADD /data /data
ADD /docs /docs
ADD /templates /templates



RUN mkdir /build
ADD . /build/
WORKDIR /build

# download dependencies
RUN go mod download


# run tests
RUN go test ./...

# build single linked binary
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o go-pandoc 

# start over using scratch image. no need for anything else anymore
FROM alpine:latest


RUN apk --no-cache add --no-check-certificate ca-certificates \
  && update-ca-certificates

COPY --from=builder /build/go-pandoc /go-pandoc/
COPY app.conf /go-pandoc/config/app.conf
COPY --from=builder /data /go-pandoc/data
COPY --from=builder /docs /go-pandoc/docs
COPY --from=builder /templates /go-pandoc/templates

WORKDIR /go-pandoc

EXPOSE 8080

CMD ["./go-pandoc", "run", "--config", "$PATH_CONFIG"]