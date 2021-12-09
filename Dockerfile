FROM golang:1-alpine3.7 as builder
WORKDIR /go/src/entrypoint
COPY ./entrypoint.go /go/src/entrypoint/
RUN go build

FROM traefik:alpine
COPY --from=builder /go/src/entrypoint/entrypoint /
COPY ./traefik.toml /etc/traefik/traefik.toml
RUN mkdir /cert
ENTRYPOINT [ "/entrypoint" ]
CMD ["/usr/local/bin/traefik"]