package main

import (
	"cel-approver-policy-plugin/internal"
	"github.com/cert-manager/approver-policy/pkg/cmd"
	"github.com/cert-manager/approver-policy/pkg/registry"
)

func main() {
	cmd.ExecutePolicyApprover()
}

// Ensure that CEL plugin gets registered with the shared registry
func init() {
	registry.Shared.Store(&internal.CELPlugin{})
}
