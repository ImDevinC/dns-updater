ARG GO_VERSION=1.17

FROM golang:${GO_VERSION} as builder

WORKDIR /src

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o /updater ./cmd/updater/main.go

FROM gcr.io/distroless/static

USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /updater /updater

ENTRYPOINT [ "/updater" ]