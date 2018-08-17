FROM cjsaylor/go-alpine-sdk:1.10 as builder
COPY . /go/src/github.com/cjsaylor/chessbot
WORKDIR /go/src/github.com/cjsaylor/chessbot
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "-s" -v -o web ./cmd/web/web.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN apk --no-cache add msttcorefonts-installer fontconfig && \
	update-ms-fonts && \
	fc-cache -f
WORKDIR /app
COPY --from=builder /go/src/github.com/cjsaylor/chessbot/assets assets
COPY --from=builder /go/src/github.com/cjsaylor/chessbot/web web
EXPOSE 8080
VOLUME [ "/app/db/" ]

CMD ["./web"]