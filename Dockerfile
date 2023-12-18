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

FROM gcr.io/distroless/static@sha256:9be3fcc6abeaf985b5ecce59451acbcbb15e7be39472320c538d0d55a0834edc
LABEL description="Experimental CEL cert-manager approver-policy plugin"

WORKDIR /
USER 1001
COPY --from=builder /workspace/cert-manager-cel-approver-policy-plugin /usr/bin/cert-manager-approver-policy

ENTRYPOINT ["/usr/bin/cert-manager-approver-policy"]
