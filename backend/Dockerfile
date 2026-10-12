FROM golang:1.24.5-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/backend ./cmd/server

FROM alpine:3.22

COPY --from=build /app/backend /usr/local/bin/backend

EXPOSE 8080

CMD ["backend"]
