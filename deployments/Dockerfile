FROM golang:1.22 AS builder

WORKDIR /go/src/app
COPY ../ ./

RUN go mod download

RUN CGO_ENABLED=0 go build -v -o /go/bin/app cmd/api/main.go

FROM gcr.io/distroless/static-debian12

COPY --from=builder /go/bin/app /

CMD ["/app"]
