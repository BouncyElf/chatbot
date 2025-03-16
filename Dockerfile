# 构建阶段
FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN go mod download 
RUN CGO_ENABLED=0 GOOS=linux go build -o /chat-server .

# 运行阶段
FROM alpine:3.18
COPY --from=builder /chat-server /chat-server
COPY --from=builder /app/bot-config.yaml /bot-config.yaml
EXPOSE 8080
CMD ["/chat-server"]
