FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o b1cron ./cmd/app

FROM alpine:latest
RUN apk add --no-cache ca-certificates sqlite python3 py3-pip bash tzdata && \
    ln -sf python3 /usr/bin/python && \
    which python3 && python3 --version

# 设置时区为中国标准时间
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
WORKDIR /app
COPY --from=builder /app/b1cron .
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/config.yaml ./config.yaml
EXPOSE 8080
CMD ["./b1cron"]