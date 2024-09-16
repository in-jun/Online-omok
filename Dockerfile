FROM golang:1.20-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o online-omok .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/online-omok .
COPY HTML/ CONFIGS/ IMAGE/ SOUND/ ./
EXPOSE 8080
CMD ["./online-omok"]
