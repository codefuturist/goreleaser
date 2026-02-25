# GoReleaser Pro Features — Surgical Changes Documentation

> **Fork:** [`codefuturist/goreleaser`](https://github.com/codefuturist/goreleaser)  
> **Upstream:** [`goreleaser/goreleaser`](https://github.com/goreleaser/goreleaser) (OSS, MIT)  
> **Config source:** [`goreleaser/goreleaser-pro`](https://github.com/goreleaser/goreleaser-pro) (public Go library)

This document catalogs every surgical change made to the upstream GoReleaser OSS
repository to enable Pro-level features. The goal is a minimal, auditable diff
that can be rebased onto future upstream releases with minimal conflicts.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Phase 1 — Config Type Swap](#phase-1--config-type-swap)
  - [1.1 Replace config.go with Pro version](#11-replace-configgo-with-pro-version)
  - [1.2 Add include.go](#12-add-includego)
  - [1.3 Add matrix.go](#13-add-matrixgo)
  - [1.4 Update marshaler.go](#14-update-marshalergo)
  - [1.5 Fix before.go — Hook type change](#15-fix-beforego--hook-type-change)
  - [1.6 Fix body.go — IncludedMarkdown type](#16-fix-bodygo--includedmarkdown-type)
  - [1.7 Remove Pro config warning from cmd/config.go](#17-remove-pro-config-warning-from-cmdconfiggo)
- [Phase 2 — Feature Pipes](#phase-2--feature-pipes)
  - [2.1 Includes pipe](#21-includes-pipe)
  - [2.2 Variables pipe + template engine](#22-variables-pipe--template-engine)
  - [2.3 Monorepo pipe](#23-monorepo-pipe)
  - [2.4 Pipeline registration](#24-pipeline-registration)
- [Summary of All Changed Files](#summary-of-all-changed-files)
- [New Pro Types Added](#new-pro-types-added)
- [New Pro Project Fields Added](#new-pro-project-fields-added)
- [Rebasing Guide](#rebasing-guide)

---

## Architecture Overview

GoReleaser uses a **pipeline architecture**: an ordered list of `Pipe` structs,
each implementing `Run(ctx *context.Context) error`. The pipeline is defined in
`internal/pipeline/pipeline.go`. Config is loaded into `pkg/config/config.go`
structs, and the template engine in `internal/tmpl/tmpl.go` provides `{{ .Field }}`
expansion throughout.

The Pro version ships as a separate closed-source binary. Its config types are
published as a public Go library at `goreleaser/goreleaser-pro`. The Pro binary
has additional pipes that implement features like includes, variables, monorepo,
DMG, MSI, NSIS, etc. — none of which exist in the OSS codebase.

**Our strategy:** Replace the OSS config types with the Pro config types (giving
us full YAML compatibility), then implement the most valuable Pro pipes natively
in the OSS pipeline.

---

## Phase 1 — Config Type Swap

**Commit:** `710d475`  
**Goal:** Make the OSS binary accept and parse all Pro config fields without errors.

### 1.1 Replace config.go with Pro version

| | |
|---|---|
| **File** | `pkg/config/config.go` |
| **Operation** | Full replacement |
| **Lines** | 1,514 → 1,932 (+418 lines) |
| **Source** | `goreleaser/goreleaser-pro` v2 `pkg/config/config.go` |

The Pro `config.go` is a **superset** of the OSS version — every OSS type and field
is preserved, with 28 new types and 18 new `Project` fields added. This is a clean
drop-in replacement because the Pro library is kept in sync with the OSS config by
the GoReleaser maintainers.

### 1.2 Add include.go

| | |
|---|---|
| **File** | `pkg/config/include.go` |
| **Operation** | New file (85 lines) |
| **Source** | `goreleaser/goreleaser-pro` v2 `pkg/config/include.go` |

Defines `Include`, `IncludeFromFile`, `IncludeFromURL`, and `IncludedMarkdown` types
with a `Load() (io.ReadCloser, error)` method that handles:
- Local file paths (`FromFile.Path`)
- HTTP/HTTPS URLs (`FromURL.URL` with optional headers)
- GitHub shorthand (`owner/repo/path` → `https://raw.githubusercontent.com/...`)
- Environment variable expansion in URL headers

### 1.3 Add matrix.go

| | |
|---|---|
| **File** | `pkg/config/matrix.go` |
| **Operation** | New file (4 lines) |
| **Source** | `goreleaser/goreleaser-pro` v2 `pkg/config/matrix.go` |

Defines `Matrix` type alias (`map[string][]string`) for future matrix build support.

### 1.4 Update marshaler.go

| | |
|---|---|
| **File** | `pkg/config/marshaler.go` |
| **Operation** | Added 2 new unmarshalers (+39 lines) |
| **Source** | `goreleaser/goreleaser-pro` v2 `pkg/config/marshaler.go` |

Added custom YAML/JSON unmarshalers for:

- **`ExtraFile`** — Accepts both string shorthand (`"path/to/file"`) and full struct
  (`{glob: "*.txt", name_template: "..."}`)
- **`IncludedMarkdown`** — Accepts both inline string content and struct with
  `from_file`/`from_url` fields

These maintain backward compatibility with existing configs that use the string forms.

### 1.5 Fix before.go — Hook type change

| | |
|---|---|
| **File** | `internal/pipe/before/before.go` |
| **Operation** | 1-line change |

```diff
-		s, err := tmpl.Apply(step)
+		s, err := tmpl.Apply(step.Cmd)
```

**Why:** In OSS, `Before.Hooks` was `[]string`. In Pro, it's `Hooks` (alias for
`[]Hook`), where `Hook` is a struct with `Cmd`, `Dir`, `Env`, `If`, and `Output`
fields. The custom YAML unmarshaler on `Hooks` still accepts plain strings for
backward compatibility, but the Go code must access `.Cmd` to get the command string.

The corresponding test file (`before_test.go`) was updated to construct `Hook{Cmd: "..."}` structs instead of raw strings.

### 1.6 Fix body.go — IncludedMarkdown type

| | |
|---|---|
| **File** | `internal/pipe/release/body.go` |
| **Operation** | Added `loadIncludedMarkdown()` helper (+15 lines) |

```diff
+func loadIncludedMarkdown(im config.IncludedMarkdown) (string, error) {
+	rc, err := im.Load()
+	if err != nil {
+		return "", err
+	}
+	defer rc.Close()
+	data, err := io.ReadAll(rc)
+	if err != nil {
+		return "", err
+	}
+	return string(data), nil
+}
```

**Why:** `Release.Header` and `Release.Footer` changed from `string` to
`IncludedMarkdown`. The new type supports inline content (`Content` field), local
files (`FromFile`), or remote URLs (`FromURL`). The helper function uses the
existing `Load()` method to transparently resolve all three sources.

The corresponding test file (`body_test.go`) was updated to use
`config.IncludedMarkdown{Content: "..."}` instead of raw strings.

### 1.7 Remove Pro config warning from cmd/config.go

| | |
|---|---|
| **File** | `cmd/config.go` |
| **Operation** | Removed dead code (−16 lines) |

Removed the `proExplain` warning message and the `ErrProConfig` error handling
that would warn users when Pro fields were detected. Since we now support Pro
config natively, this check is no longer needed (and would never trigger anyway,
since the Pro config types don't emit that error).

---

## Phase 2 — Feature Pipes

**Commit:** `47963b8`  
**Goal:** Implement the three most valuable Pro features as native pipeline pipes.

### 2.1 Includes pipe

| | |
|---|---|
| **File** | `internal/pipe/includes/includes.go` (92 lines) |
| **Test** | `internal/pipe/includes/includes_test.go` (85 lines, 5 tests) |

**What it does:**
1. Iterates over `ctx.Config.Includes` (list of `Include` structs)
2. For each include, calls `Include.Load()` to fetch content (file or URL)
3. Unmarshals the YAML into a partial `config.Project`
4. Deep-merges partials into an accumulator using YAML round-trip merging
5. Merges the main config on top (main always wins)
6. Clears `Includes` to prevent re-processing

**Merge strategy:** YAML marshal → unmarshal overlay. Non-empty fields from later
sources override earlier ones. Slices replace (no append). This matches the Pro
behavior where main config fields take precedence over includes.

**Pipeline position:** FIRST — before `dist.CleanPipe{}`, so the full merged config
is available to all downstream pipes.

### 2.2 Variables pipe + template engine

| | |
|---|---|
| **File** | `internal/pipe/variables/variables.go` (27 lines) |
| **Test** | `internal/pipe/variables/variables_test.go` (39 lines, 4 tests) |
| **Also modified** | `internal/tmpl/tmpl.go` (+2 lines) |

**What it does:**
- The pipe itself is minimal — it logs variables and ensures they're present on
  `ctx.Config.Variables` for the template engine
- The real work is a 2-line change in `tmpl.go`:

```diff
+	varK            = "Var"
 	...
+		varK:            ctx.Config.Variables,
```

This adds a `"Var"` key to the template engine's `Fields` map, populated from
`ctx.Config.Variables` (`map[string]any`). Templates can now use:

```yaml
variables:
  app_name: myapp
  registry: ghcr.io

builds:
  - binary: "{{ .Var.app_name }}"

dockers:
  - image_templates:
      - "{{ .Var.registry }}/{{ .Var.app_name }}:{{ .Version }}"
```

**Pipeline position:** After includes, before `env` — so variables from included
configs are available, and all environment resolution can reference them.

### 2.3 Monorepo pipe

| | |
|---|---|
| **File** | `internal/pipe/monorepo/monorepo.go` (119 lines) |
| **Test** | `internal/pipe/monorepo/monorepo_test.go` (81 lines, 6 tests) |

**What it does:**
1. Reads `ctx.Config.Monorepo.Dir` and `TagPrefix`
2. If `TagPrefix` is empty, defaults to `basename(Dir) + "/"`
3. Resolves the current git tag matching the prefix (e.g., `myapp/v1.2.3`)
4. Strips the prefix to extract the version
5. Resolves the previous tag with the same prefix for changelog diffing
6. Rewrites all `Build.Dir` paths to be relative to the monorepo dir
7. Scopes `Dist` to `<dir>/dist`
8. Sets `Changelog.Paths` to scope the changelog to the subdirectory

**Config example:**
```yaml
monorepo:
  dir: apps/myapp
  tag_prefix: myapp/

builds:
  - main: ./cmd/server  # resolved as apps/myapp/cmd/server
```

**Pipeline position:** After `git` pipe, before `semver` — needs git info to
resolve tags, and must modify the version before semver parsing.

### 2.4 Pipeline registration

| | |
|---|---|
| **File** | `internal/pipeline/pipeline.go` (+10 lines) |

Three imports added and three pipe entries inserted into `BuildPipeline`:

```go
var BuildPipeline = []Piper{
	includes.Pipe{},    // NEW — load and merge config includes
	variables.Pipe{},   // NEW — inject template variables
	dist.CleanPipe{},
	env.Pipe{},
	git.Pipe{},
	monorepo.Pipe{},    // NEW — monorepo scoping (after git, before semver)
	semver.Pipe{},
	// ... rest unchanged
}
```

---

## Summary of All Changed Files

| File | Operation | Lines Changed | Phase |
|------|-----------|:---:|:---:|
| `pkg/config/config.go` | Replaced with Pro version | +418 | 1 |
| `pkg/config/include.go` | **New** — Include types + Load() | +85 | 1 |
| `pkg/config/matrix.go` | **New** — Matrix type | +4 | 1 |
| `pkg/config/marshaler.go` | Added ExtraFile + IncludedMarkdown unmarshalers | +39 | 1 |
| `internal/pipe/before/before.go` | `step` → `step.Cmd` | +1 −1 | 1 |
| `internal/pipe/before/before_test.go` | Updated for Hook struct | ~20 | 1 |
| `internal/pipe/release/body.go` | Added `loadIncludedMarkdown()` | +15 | 1 |
| `internal/pipe/release/body_test.go` | Updated for IncludedMarkdown | ~16 | 1 |
| `cmd/config.go` | Removed Pro warning | −16 | 1 |
| `internal/pipe/includes/includes.go` | **New** — Includes pipe | +92 | 2 |
| `internal/pipe/includes/includes_test.go` | **New** — 5 tests | +85 | 2 |
| `internal/pipe/variables/variables.go` | **New** — Variables pipe | +27 | 2 |
| `internal/pipe/variables/variables_test.go` | **New** — 4 tests | +39 | 2 |
| `internal/pipe/monorepo/monorepo.go` | **New** — Monorepo pipe | +119 | 2 |
| `internal/pipe/monorepo/monorepo_test.go` | **New** — 6 tests | +81 | 2 |
| `internal/pipeline/pipeline.go` | Registered 3 new pipes | +10 | 2 |
| `internal/tmpl/tmpl.go` | Added `.Var` template key | +2 | 2 |
| **Total** | **17 files** | **+1,051 −56** | |

---

## New Pro Types Added

These 28 types exist in the Pro config but not in OSS:

| Type | Purpose |
|------|---------|
| `After` | Post-release hooks |
| `AppBundle` | macOS `.app` bundle packaging |
| `ArchiveHooks` | Before/after hooks on archive creation |
| `BeforePublishHook` | Pre-publish hooks |
| `ChangelogAI` | AI-powered changelog generation (OpenAI/Anthropic/Ollama) |
| `ChangelogSubgroup` | Nested changelog grouping |
| `Cloudsmith` | Cloudsmith package registry publishing |
| `DMG` | macOS `.dmg` disk image creation |
| `DockerHub` | Docker Hub publishing |
| `Fury` | Gemfury package publishing |
| `MacOSNotarizeNative` | Native macOS notarization (without quill) |
| `MacOSPkg` | macOS `.pkg` installer creation |
| `MacOSSignNative` | Native macOS code signing |
| `MacOSSignNotarizeNative` | Combined sign + notarize |
| `MakeselfTemplatedFile` | Templated files for makeself archives |
| `Matrix` | Build matrix support |
| `Monorepo` | Monorepo directory + tag prefix config |
| `MSI` | Windows `.msi` installer creation |
| `MSIHooks` | Before/after hooks on MSI creation |
| `Nightly` | Nightly release configuration |
| `NPM` | NPM package publishing |
| `NSIS` | Windows NSIS installer creation |
| `Partial` | Partial build/release configuration |
| `PreBuiltOptions` | Pre-built binary options |
| `TemplatedExtraFile` | Extra files with template support |
| `TemplatedExtraFileWithMode` | Extra files with template + file mode |
| `TemplatedFile` | Standalone templated file generation |
| `TemplateFile` | Template file reference |

---

## New Pro Project Fields Added

These 18 fields were added to the `Project` struct:

| Field | YAML Key | Type | Description |
|-------|----------|------|-------------|
| `After` | `after` | `After` | Post-release hooks |
| `AppBundles` | `app_bundles` | `[]AppBundle` | macOS app bundles |
| `BeforePublish` | `before_publish` | `[]BeforePublishHook` | Pre-publish hooks |
| `Cloudsmiths` | `cloudsmiths` | `[]Cloudsmith` | Cloudsmith publishing |
| `DMG` | `dmg` | `[]DMG` | macOS DMG images |
| `DockerHubs` | `dockerhub` | `[]DockerHub` | Docker Hub publishing |
| `Furies` | `furies` | `[]Fury` | Gemfury publishing (deprecated) |
| `Gemfury` | `gemfury` | `[]Fury` | Gemfury publishing |
| `Includes` | `includes` | `[]Include` | Config file includes ✅ **Implemented** |
| `Monorepo` | `monorepo` | `Monorepo` | Monorepo scoping ✅ **Implemented** |
| `MSI` | `msi` | `[]MSI` | Windows MSI installers |
| `Nightly` | `nightly` | `Nightly` | Nightly releases |
| `NPMs` | `npms` | `[]NPM` | NPM publishing |
| `NSIs` | `nsis` | `[]NSIS` | NSIS installers |
| `Partial` | `partial` | `Partial` | Partial builds |
| `Pkgs` | `pkgs` | `[]MacOSPkg` | macOS PKG installers |
| `TemplateFiles` | `template_files` | `[]TemplateFile` | Templated file generation |
| `Variables` | `variables` | `map[string]any` | Template variables ✅ **Implemented** |

---

## Rebasing Guide

When a new upstream GoReleaser release is published:

```bash
# Add upstream remote (if not already done)
git remote add upstream https://github.com/goreleaser/goreleaser.git

# Fetch latest upstream
git fetch upstream

# Rebase our changes onto the latest release tag
git rebase upstream/main
```

**Expected conflicts:**

| File | Likelihood | Resolution |
|------|-----------|------------|
| `pkg/config/config.go` | **High** | Re-apply the Pro config from `goreleaser-pro` latest release |
| `pkg/config/marshaler.go` | Medium | Keep our additions, merge upstream changes |
| `internal/pipe/before/before.go` | Low | Single-line change, easy to re-apply |
| `internal/pipe/release/body.go` | Low | Keep our `loadIncludedMarkdown()` helper |
| `cmd/config.go` | Low | Keep Pro warning removed |
| `internal/pipeline/pipeline.go` | Medium | Re-add our 3 pipe registrations |
| `internal/tmpl/tmpl.go` | Low | Re-add the 2-line `.Var` addition |

**New pipe files will never conflict** since they are in new directories that don't
exist upstream.

**Recommended sync procedure:**
1. Update `pkg/config/config.go` from latest `goreleaser/goreleaser-pro`
2. Check if any new type changes cause compile errors (like Phase 1 fixes)
3. Run `go build ./...` and fix any breakage
4. Run `go test ./internal/pipe/includes/... ./internal/pipe/variables/... ./internal/pipe/monorepo/...`
5. Run full test suite to verify no regressions
