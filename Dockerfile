FROM golang:1-alpine3.14 as builder
WORKDIR /src
COPY ./entrypoint.go .
COPY ./go.mod .
RUN go build -o entrypoint

FROM traefik:alpine
COPY --from=builder /src/entrypoint /
COPY ./traefik.toml /etc/traefik/traefik.toml
RUN mkdir /cert
ENTRYPOINT [ "/entrypoint" ]
CMD ["/usr/local/bin/traefik"]
