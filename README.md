# gorelease
build and release single-binary go applications


```yaml
global:
  name: "gorelease%s-%s-%s"
  version: v0.1.0
  dir: bin
  flags:
    - -mod=vendor
    - -ldflags="-X 'main.Version=v0.1.0'"
  env:
    CGO_ENABLED: 0
    GO111MODULE: on


targets:
  - file: main.go
    name: "main%s-%s-%s"
    platforms:
      windows: ["amd64", "386"]
      linux: ["amd64"]
      darwin: ["amd64"]

  - file: main.go
    name: "main2%s-%s-%s"
    platforms:
      linux: ["amd64"]
````