FROM cjsaylor/go-alpine-sdk:1.10 as builder
COPY . /go/src/github.com/cjsaylor/chessbot
WORKDIR /go/src/github.com/cjsaylor/chessbot
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "-s" -v -o web ./cmd/web/web.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN apk --no-cache add msttcorefonts-installer fontconfig && \
	update-ms-fonts && \
	fc-cache -f
RUN adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=builder /go/src/github.com/cjsaylor/chessbot/assets assets
COPY --from=builder /go/src/github.com/cjsaylor/chessbot/web web
RUN chmod -R 444 assets/*
RUN mkdir data && chown -R appuser data
USER appuser
EXPOSE 8080
VOLUME [ "/app/data" ]

CMD ["./web"]