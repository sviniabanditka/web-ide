FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /ide-server ./cmd/ide-server/

FROM alpine:3.21

RUN apk add --no-cache ca-certificates git bash

WORKDIR /app
COPY --from=builder /ide-server .

RUN mkdir -p /app/data /app/projects

ENV IDE_DATA_DIR=/app/data
ENV IDE_PROJECTS_DIR=/app/projects
ENV IDE_HTTP_ADDR=:8080

EXPOSE 8080

CMD ["./ide-server"]
