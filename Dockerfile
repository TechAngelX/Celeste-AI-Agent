FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o celeste-agent main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/celeste-agent .
COPY --from=builder /app/home.html .


EXPOSE 8080
CMD ["./celeste-agent"]
