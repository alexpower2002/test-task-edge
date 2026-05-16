FROM golang:latest AS builder
WORKDIR /go/src/app
ENV GO111MODULE=on

ADD go.mod .
ADD go.sum .

ADD . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/lendingparser ./cmd/lendingparser

FROM alpine:latest AS app
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /go/src/app/bin/lendingparser .
COPY --from=builder /go/src/app/migrations/sql migrations/sql
COPY --from=builder /go/src/app/jobs.json .
COPY --from=builder /go/src/app/.env .
ENTRYPOINT ["./lendingparser"]
