# Build the approver-policy binary
FROM docker.io/library/golang:1.20 as builder
ARG GOPROXY

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source files
COPY cmd/ cmd/
COPY internal/ internal/

# Build
RUN go build -o cert-manager-cel-approver-policy-plugin cmd/main.go

FROM gcr.io/distroless/static:nonroot
LABEL description="Experimental CEL cert-manager approver-policy plugin"

WORKDIR /
USER 65532:65532
COPY --from=builder /workspace/cert-manager-cel-approver-policy-plugin /usr/bin/cert-manager-approver-policy

ENTRYPOINT ["/usr/bin/cert-manager-approver-policy"]
