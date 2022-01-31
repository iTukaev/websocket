FROM golang:alpine3.15 as Builder
COPY . /go/src/
WORKDIR /go/src/
RUN go build -o /go/bin/websocked /go/src/cmd/*.go

FROM alpine:3.15
COPY --from=Builder /go/bin/websocked /websocked
EXPOSE 8081
LABEL developer="iTukaev" maintainer="RarogCmex" description="Websocket daemon for betspoiler"
# See .env for description of environment variables
COPY .env /.env
CMD ["/websocked"]
