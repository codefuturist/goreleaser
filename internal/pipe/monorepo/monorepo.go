// Package monorepo implements monorepo support — scoping the entire
// release pipeline to a subdirectory with prefixed tags.
package monorepo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/caarlos0/log"
	"github.com/goreleaser/goreleaser/v2/internal/git"
	"github.com/goreleaser/goreleaser/v2/pkg/context"
)

// Pipe configures monorepo scoping.
type Pipe struct{}

func (Pipe) String() string { return "monorepo" }

func (Pipe) Skip(ctx *context.Context) bool {
	return ctx.Config.Monorepo.Dir == ""
}

// Run rewrites the context for monorepo scoping:
//   - Resolves the current tag using the tag prefix
//   - Strips the tag prefix from the version
//   - Rewrites build directories to be relative to monorepo dir
//   - Scopes dist to monorepo dir
//   - Adds PrefixedTag and PrefixedPreviousTag to the context
func (Pipe) Run(ctx *context.Context) error {
	mono := ctx.Config.Monorepo
	dir := mono.Dir
	prefix := mono.TagPrefix
	if prefix == "" {
		prefix = filepath.Base(dir) + "/"
	}

	log.WithField("dir", dir).
		WithField("tag_prefix", prefix).
		Info("monorepo mode")

	// Resolve current tag with prefix
	tag, err := resolveTag(ctx, prefix)
	if err != nil && !ctx.Snapshot {
		return fmt.Errorf("monorepo: resolving tag with prefix %q: %w", prefix, err)
	}
	if tag != "" {
		ctx.Git.CurrentTag = tag
		ctx.Version = strings.TrimPrefix(strings.TrimPrefix(tag, prefix), "v")
		log.WithField("tag", tag).WithField("version", ctx.Version).Info("monorepo tag resolved")
	}

	// Resolve previous tag with prefix
	prevTag, err := resolvePreviousTag(ctx, prefix, tag)
	if err == nil && prevTag != "" {
		ctx.Git.PreviousTag = prevTag
	}

	// Rewrite build dirs
	for i := range ctx.Config.Builds {
		if ctx.Config.Builds[i].Dir == "" {
			ctx.Config.Builds[i].Dir = dir
		} else {
			ctx.Config.Builds[i].Dir = filepath.Join(dir, ctx.Config.Builds[i].Dir)
		}
	}

	// Scope dist
	if ctx.Config.Dist == "" || ctx.Config.Dist == "dist" {
		ctx.Config.Dist = filepath.Join(dir, "dist")
	}

	// Filter changelog paths to monorepo dir
	if len(ctx.Config.Changelog.Paths) == 0 {
		ctx.Config.Changelog.Paths = []string{dir + "/"}
	}

	return nil
}

func resolveTag(ctx *context.Context, prefix string) (string, error) {
	// Try to find the most recent tag matching the prefix
	out, err := git.Clean(git.Run(ctx, "tag", "--list", prefix+"*", "--sort=-version:refname"))
	if err != nil {
		return "", err
	}
	if out == "" {
		return "", fmt.Errorf("no tags found with prefix %q", prefix)
	}
	// git.Clean returns only the first line
	return strings.TrimSpace(out), nil
}

func resolvePreviousTag(ctx *context.Context, prefix, current string) (string, error) {
	if current == "" {
		return "", nil
	}
	// Find tags matching prefix, sorted by version descending
	out, err := git.Run(ctx, "tag", "--list", prefix+"*", "--sort=-version:refname")
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	found := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == current {
			found = true
			continue
		}
		if found {
			return line, nil
		}
	}
	return "", nil
}
