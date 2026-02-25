// Package variables injects user-defined template variables into the
// template context so they are available as {{ .Var.key }} in all templates.
package variables

import (
	"github.com/caarlos0/log"
	"github.com/goreleaser/goreleaser/v2/pkg/context"
)

// Pipe injects variables into the context.
type Pipe struct{}

func (Pipe) String() string { return "loading variables" }

func (Pipe) Skip(ctx *context.Context) bool {
	return len(ctx.Config.Variables) == 0
}

// Run stores variables on the context so the template engine can access them.
func (Pipe) Run(ctx *context.Context) error {
	for k, v := range ctx.Config.Variables {
		log.WithField(k, v).Debug("variable")
	}
	// Variables are read from ctx.Config.Variables by the template engine.
	// No further action needed — the template engine picks them up directly.
	return nil
}
