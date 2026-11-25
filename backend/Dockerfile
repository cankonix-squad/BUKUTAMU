FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache tzdata ca-certificates && adduser -D -u 10001 app
ENV TZ=Asia/Jakarta STORAGE_DIR=/data/foto
COPY --from=build /out/server /usr/local/bin/server
RUN mkdir -p /data/foto && chown -R app /data
USER app
EXPOSE 8080
CMD ["server"]
