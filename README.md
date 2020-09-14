# gorelease
idiomatic build and release single-binary go applications

https://console.cloud.google.com/storage/browser/gorelease

```yaml
version: v0.1.0
env:
  CGO_ENABLED: 0
  GO111MODULE: on
flags:
  - -ldflags="-X 'main.Version=v0.1.0'"

targets:

  - name: "gorelease"
    file: main.go
    platforms:
      windows: ["386", "amd64", "arm"]
      linux: ["386", "amd64"]

````