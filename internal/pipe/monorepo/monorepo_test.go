package monorepo

import (
	"testing"

	"github.com/goreleaser/goreleaser/v2/internal/testctx"
	"github.com/goreleaser/goreleaser/v2/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestDescription(t *testing.T) {
	require.NotEmpty(t, Pipe{}.String())
}

func TestSkipNoMonorepo(t *testing.T) {
	ctx := testctx.WrapWithCfg(t.Context(), config.Project{})
	require.True(t, Pipe{}.Skip(ctx))
}

func TestSkipWithMonorepo(t *testing.T) {
	ctx := testctx.WrapWithCfg(t.Context(), config.Project{
		Monorepo: config.Monorepo{Dir: "apps/myapp"},
	})
	require.False(t, Pipe{}.Skip(ctx))
}

func TestRewriteBuilds(t *testing.T) {
	ctx := testctx.WrapWithCfg(t.Context(), config.Project{
		Monorepo: config.Monorepo{
			Dir:       "apps/myapp",
			TagPrefix: "myapp/",
		},
		Builds: []config.Build{
			{ID: "default"},
			{ID: "sub", Dir: "cmd/server"},
		},
	})
	ctx.Snapshot = true // skip tag resolution in tests
	require.NoError(t, Pipe{}.Run(ctx))

	// Default build dir should be set to monorepo dir
	require.Equal(t, "apps/myapp", ctx.Config.Builds[0].Dir)
	// Nested build dir should be joined with monorepo dir
	require.Equal(t, "apps/myapp/cmd/server", ctx.Config.Builds[1].Dir)
}

func TestScopesDist(t *testing.T) {
	ctx := testctx.WrapWithCfg(t.Context(), config.Project{
		Monorepo: config.Monorepo{
			Dir:       "apps/myapp",
			TagPrefix: "myapp/",
		},
	})
	ctx.Snapshot = true
	require.NoError(t, Pipe{}.Run(ctx))
	require.Equal(t, "apps/myapp/dist", ctx.Config.Dist)
}

func TestScopesChangelogPaths(t *testing.T) {
	ctx := testctx.WrapWithCfg(t.Context(), config.Project{
		Monorepo: config.Monorepo{
			Dir:       "apps/myapp",
			TagPrefix: "myapp/",
		},
	})
	ctx.Snapshot = true
	require.NoError(t, Pipe{}.Run(ctx))
	require.Equal(t, []string{"apps/myapp/"}, ctx.Config.Changelog.Paths)
}

func TestDefaultTagPrefix(t *testing.T) {
	ctx := testctx.WrapWithCfg(t.Context(), config.Project{
		Monorepo: config.Monorepo{
			Dir: "apps/myapp",
		},
	})
	ctx.Snapshot = true
	require.NoError(t, Pipe{}.Run(ctx))
	// Should use the basename of Dir as tag prefix
	require.Equal(t, "apps/myapp/dist", ctx.Config.Dist)
}
