FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/tmmweb ./cmd/tmmweb

FROM alpine:3.20
RUN adduser -D -H -u 1000 app
WORKDIR /app
COPY --from=build /out/tmmweb /usr/local/bin/tmmweb
ENV TMMWEB_ADDR=:8080
ENV TMMWEB_DATA=/config
VOLUME ["/config", "/media"]
EXPOSE 8080
USER app
ENTRYPOINT ["tmmweb"]
