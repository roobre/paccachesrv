FROM golang:1.18-alpine3.16 as builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o /paccachesrv ./cmd

FROM alpine:3.16

COPY --from=builder /paccachesrv /bin/

ENTRYPOINT ["/bin/paccachesrv"]
CMD []
