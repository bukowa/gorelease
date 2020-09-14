# linux:arm64
# golang:1.15.2-alpine3.12

FROM golang@sha256:fc801399d044a8e01f125eeb5aa3f160a0d12d6e03ba17a1d0b22ce50dfede81 as BUILDER
WORKDIR "/app/build/src"

ENV CGO_ENABLED=0
ENV GO111MODULE=on

COPY . .
RUN go build -mod vendor \
    -o /app/build/gorelease \
    -ldflags="-X 'main.Version=v0.1.0'" \
    gorelease/main.go


FROM golang@sha256:fc801399d044a8e01f125eeb5aa3f160a0d12d6e03ba17a1d0b22ce50dfede81
WORKDIR "/app"
COPY --from=BUILDER /app/build/gorelease /usr/local/bin/gorelease
RUN chmod +x /usr/local/bin/gorelease
ENTRYPOINT ["gorelease"]