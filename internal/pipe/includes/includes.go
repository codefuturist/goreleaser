// Package includes implements config includes — loading and merging
// partial configs from files or URLs into the main project config.
package includes

import (
	"fmt"
	"io"

	"github.com/caarlos0/log"
	"github.com/goreleaser/goreleaser/v2/internal/yaml"
	"github.com/goreleaser/goreleaser/v2/pkg/config"
	"github.com/goreleaser/goreleaser/v2/pkg/context"
)

// Pipe loads and merges included config files.
type Pipe struct{}

func (Pipe) String() string { return "loading includes" }

func (Pipe) Skip(ctx *context.Context) bool {
	return len(ctx.Config.Includes) == 0
}

// Run loads each include and deep-merges it into the project config.
// Includes are processed in order; later includes override earlier ones.
// The main config always wins over included values.
func (Pipe) Run(ctx *context.Context) error {
	main := ctx.Config
	var base config.Project

	for _, inc := range main.Includes {
		log.WithField("from", inc.String()).Info("loading include")
		partial, err := loadInclude(inc)
		if err != nil {
			return fmt.Errorf("loading include %q: %w", inc.String(), err)
		}
		base = mergeConfigs(base, partial)
	}

	// Merge main on top of the accumulated base — main fields always win.
	ctx.Config = mergeConfigs(base, main)
	// Clear includes so they are not re-processed.
	ctx.Config.Includes = nil
	return nil
}

func loadInclude(inc config.Include) (config.Project, error) {
	rc, err := inc.Load()
	if err != nil {
		return config.Project{}, err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return config.Project{}, err
	}

	var proj config.Project
	if err := yaml.Unmarshal(data, &proj); err != nil {
		return config.Project{}, err
	}
	return proj, nil
}

// mergeConfigs merges src into dst. Non-zero/non-empty src fields override dst.
// Slices from src replace dst slices (they don't append).
// This is a shallow merge at the top-level Project fields — deeper nesting
// uses the standard YAML struct unmarshaling behavior.
func mergeConfigs(dst, src config.Project) config.Project {
	// Re-marshal both to YAML, then unmarshal src on top of dst.
	// This gives us a field-level merge without writing 100+ field checks.
	dstBytes, err := yaml.Marshal(dst)
	if err != nil {
		return src
	}
	srcBytes, err := yaml.Marshal(src)
	if err != nil {
		return dst
	}

	var merged config.Project
	// First load dst as base
	if err := yaml.Unmarshal(dstBytes, &merged); err != nil {
		return src
	}
	// Then overlay src — non-empty fields from src win
	if err := yaml.Unmarshal(srcBytes, &merged); err != nil {
		return dst
	}
	return merged
}
