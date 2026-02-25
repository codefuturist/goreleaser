package includes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goreleaser/goreleaser/v2/internal/testctx"
	"github.com/goreleaser/goreleaser/v2/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestDescription(t *testing.T) {
	require.NotEmpty(t, Pipe{}.String())
}

func TestSkip(t *testing.T) {
	ctx := testctx.WrapWithCfg(t.Context(), config.Project{})
	require.True(t, Pipe{}.Skip(ctx))
}

func TestRunWithFileInclude(t *testing.T) {
	dir := t.TempDir()

	// Write a base config file
	base := `
builds:
  - id: base-build
    goos:
      - linux
    goarch:
      - amd64
`
	basePath := filepath.Join(dir, "base.yaml")
	require.NoError(t, os.WriteFile(basePath, []byte(base), 0o644))

	ctx := testctx.WrapWithCfg(t.Context(), config.Project{
		Includes: []config.Include{
			{FromFile: config.IncludeFromFile{Path: basePath}},
		},
		ProjectName: "myproject",
	})

	require.NoError(t, Pipe{}.Run(ctx))
	// Includes should be cleared after processing
	require.Nil(t, ctx.Config.Includes)
	// Project name from main config should be preserved
	require.Equal(t, "myproject", ctx.Config.ProjectName)
	// Build from include should be present
	require.Len(t, ctx.Config.Builds, 1)
	require.Equal(t, "base-build", ctx.Config.Builds[0].ID)
}

func TestRunMainOverridesInclude(t *testing.T) {
	dir := t.TempDir()

	base := `
project_name: from-include
builds:
  - id: included-build
`
	basePath := filepath.Join(dir, "base.yaml")
	require.NoError(t, os.WriteFile(basePath, []byte(base), 0o644))

	ctx := testctx.WrapWithCfg(t.Context(), config.Project{
		Includes: []config.Include{
			{FromFile: config.IncludeFromFile{Path: basePath}},
		},
		ProjectName: "main-project",
	})

	require.NoError(t, Pipe{}.Run(ctx))
	// Main config's project_name should win
	require.Equal(t, "main-project", ctx.Config.ProjectName)
}

func TestRunInvalidInclude(t *testing.T) {
	ctx := testctx.WrapWithCfg(t.Context(), config.Project{
		Includes: []config.Include{
			{FromFile: config.IncludeFromFile{Path: "/nonexistent/file.yaml"}},
		},
	})

	require.Error(t, Pipe{}.Run(ctx))
}
