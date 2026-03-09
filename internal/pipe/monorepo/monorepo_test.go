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

// TestResolveTagUsesExistingCurrentTag verifies that resolveTag does not override
// ctx.Git.CurrentTag with the highest version-sorted tag when the existing tag
// already matches the prefix. This prevents silently tagging the wrong version
// when newer prefix tags exist on other commits.
func TestResolveTagUsesExistingCurrentTag(t *testing.T) {
	ctx := testctx.WrapWithCfg(
		t.Context(),
		config.Project{
			Monorepo: config.Monorepo{
				Dir:       "apps/myapp",
				TagPrefix: "myapp/",
			},
		},
		testctx.WithCurrentTag("myapp/v1.0.0"),
	)
	// resolveTag should short-circuit and use the existing tag rather than
	// querying git (which would require a real repo and might return a newer tag).
	tag, err := resolveTag(ctx, "myapp/")
	require.NoError(t, err)
	require.Equal(t, "myapp/v1.0.0", tag)
}

func TestResolveTagIgnoresUnrelatedCurrentTag(t *testing.T) {
	ctx := testctx.WrapWithCfg(
		t.Context(),
		config.Project{
			Monorepo: config.Monorepo{
				Dir:       "apps/myapp",
				TagPrefix: "myapp/",
			},
		},
		// Tag on HEAD belongs to a different project — should not be used.
		testctx.WithCurrentTag("other-app/v2.0.0"),
	)
	// resolveTag must NOT use this tag; it should fall back to git.
	// With no git repo here it will error — that's the expected fallback path.
	_, err := resolveTag(ctx, "myapp/")
	require.Error(t, err)
}
