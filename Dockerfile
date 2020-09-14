# linux:arm64

FROM golang@sha256:5219b39d2d6bf723fb0221633a0ff831b0f89a94beb5a8003c7ff18003f48ead as BUILDER
WORKDIR "/app/build/src"

ENV CGO_ENABLED=0
ENV GO111MODULE=on

COPY . .
RUN go build -mod vendor \
    -o /app/build/gorelease \
    -ldflags="-X 'main.Version=v0.1.0'" \
    gorelease/main.go

FROM alpine@sha256:a15790640a6690aa1730c38cf0a440e2aa44aaca9b0e8931a9f2b0d7cc90fd65
WORKDIR "/app"
COPY --from=BUILDER /app/build/gorelease /usr/local/bin/gorelease
RUN chmod +x /usr/local/bin/gorelease
ENTRYPOINT ["gorelease"]