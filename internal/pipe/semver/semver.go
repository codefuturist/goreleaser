// Package semver handles semver parsing.
package semver

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/caarlos0/log"
	"github.com/goreleaser/goreleaser/v2/internal/pipe"
	"github.com/goreleaser/goreleaser/v2/internal/skips"
	"github.com/goreleaser/goreleaser/v2/pkg/context"
)

// Pipe is a global hook pipe.
type Pipe struct{}

// String is the name of this pipe.
func (Pipe) String() string {
	return "parsing tag"
}

// Run executes the hooks.
func (Pipe) Run(ctx *context.Context) error {
	// ctx.Version is always set by a prior pipe to the correct semver string
	// without any tag prefix. The git pipe sets it by stripping the leading "v"
	// from the tag; the monorepo pipe overwrites it by stripping the full
	// monorepo prefix (e.g. "git-patrol/v0.2.0" → "0.2.0"). Using ctx.Version
	// here avoids failures when ctx.Git.CurrentTag contains a monorepo prefix
	// that is not valid semver.
	sv, err := semver.NewVersion("v" + ctx.Version)
	if err != nil {
		if skips.Any(ctx, skips.Validate) {
			log.WithError(err).
				WithField("tag", ctx.Git.CurrentTag).
				WithField("version", ctx.Version).
				Warn("current tag is not semver")
			return pipe.ErrSkipValidateEnabled
		}
		return fmt.Errorf("failed to parse tag '%s' (version %q) as semver: %w", ctx.Git.CurrentTag, ctx.Version, err)
	}
	ctx.Semver = context.Semver{
		Major:      sv.Major(),
		Minor:      sv.Minor(),
		Patch:      sv.Patch(),
		Prerelease: sv.Prerelease(),
	}
	return nil
}
