package variables

import (
	"testing"

	"github.com/goreleaser/goreleaser/v2/internal/testctx"
	"github.com/goreleaser/goreleaser/v2/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestDescription(t *testing.T) {
	require.NotEmpty(t, Pipe{}.String())
}

func TestSkipNoVariables(t *testing.T) {
	ctx := testctx.WrapWithCfg(t.Context(), config.Project{})
	require.True(t, Pipe{}.Skip(ctx))
}

func TestSkipWithVariables(t *testing.T) {
	ctx := testctx.WrapWithCfg(t.Context(), config.Project{
		Variables: map[string]any{"foo": "bar"},
	})
	require.False(t, Pipe{}.Skip(ctx))
}

func TestRun(t *testing.T) {
	vars := map[string]any{
		"app_name": "myapp",
		"version":  42,
	}
	ctx := testctx.WrapWithCfg(t.Context(), config.Project{
		Variables: vars,
	})
	require.NoError(t, Pipe{}.Run(ctx))
	// Variables should remain on config for template engine to pick up
	require.Equal(t, "myapp", ctx.Config.Variables["app_name"])
	require.Equal(t, 42, ctx.Config.Variables["version"])
}
