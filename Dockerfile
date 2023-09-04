# Build the approver-policy binary
FROM docker.io/library/golang:1.21 as builder
ARG GOPROXY

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source files
COPY cmd/ cmd/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 go build -o cert-manager-cel-approver-policy-plugin cmd/main.go

FROM gcr.io/distroless/static@sha256:e7e79fb2947f38ce0fab6061733f7e1959c12b843079042fe13f56ca7b9d178c
LABEL description="Experimental CEL cert-manager approver-policy plugin"

WORKDIR /
USER 1001
COPY --from=builder /workspace/cert-manager-cel-approver-policy-plugin /usr/bin/cert-manager-approver-policy

ENTRYPOINT ["/usr/bin/cert-manager-approver-policy"]
