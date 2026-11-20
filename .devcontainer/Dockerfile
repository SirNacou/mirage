FROM mcr.microsoft.com/devcontainers/go:1.27-trixie

RUN go install github.com/stephenafamo/bob/gen/bobgen-psql@latest \
  && chown -R vscode:vscode /go
