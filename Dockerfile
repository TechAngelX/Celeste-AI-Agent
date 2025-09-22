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
COPY --from=builder /app/web/home.html .
COPY --from=builder /app/data/itemCatalogue.json ./data/
COPY --from=builder /app/api-comparison.html .


RUN mkdir -p data

EXPOSE 8080
CMD ["./celeste-agent"]
