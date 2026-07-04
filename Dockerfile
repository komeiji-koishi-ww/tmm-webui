FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/tmmweb ./cmd/tmmweb

FROM alpine:3.20
RUN apk add --no-cache ca-certificates ffmpeg && adduser -D -H -u 1000 app
WORKDIR /app
COPY --from=build /out/tmmweb /usr/local/bin/tmmweb
ENV TMMWEB_ADDR=:8080
ENV TMMWEB_DATA=/config
ENV TMMWEB_SCAN_MEDIAINFO=1
ENV TMMWEB_MEDIAINFO_WORKERS=1
VOLUME ["/config", "/media"]
EXPOSE 8080
USER app
ENTRYPOINT ["tmmweb"]
