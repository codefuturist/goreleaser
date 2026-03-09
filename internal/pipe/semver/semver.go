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
	// If the monorepo pipe already resolved the version (e.g. stripped a tag
	// prefix like "git-patrol/"), construct a semver-parseable string from it
	// instead of parsing the raw tag which may contain slashes.
	tagToParse := ctx.Git.CurrentTag
	if ctx.Version != "" {
		tagToParse = "v" + ctx.Version
	}
	sv, err := semver.NewVersion(tagToParse)
	if err != nil {
		if skips.Any(ctx, skips.Validate) {
			log.WithError(err).
				WithField("tag", ctx.Git.CurrentTag).
				Warn("current tag is not semver")
			return pipe.ErrSkipValidateEnabled
		}
		return fmt.Errorf("failed to parse tag '%s' as semver: %w", ctx.Git.CurrentTag, err)
	}
	ctx.Semver = context.Semver{
		Major:      sv.Major(),
		Minor:      sv.Minor(),
		Patch:      sv.Patch(),
		Prerelease: sv.Prerelease(),
	}
	return nil
}
