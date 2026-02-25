package release

import (
	"bytes"
	"io"
	"os"
	"text/template"

	"github.com/goreleaser/goreleaser/v2/internal/artifact"
	"github.com/goreleaser/goreleaser/v2/internal/tmpl"
	"github.com/goreleaser/goreleaser/v2/pkg/config"
	"github.com/goreleaser/goreleaser/v2/pkg/context"
)

const bodyTemplateText = `{{ with .Header }}{{ . }}{{ "\n" }}{{ end }}
{{- .ReleaseNotes }}
{{- with .Footer }}{{ "\n" }}{{ . }}{{ end }}
`

func describeBody(ctx *context.Context) (bytes.Buffer, error) {
	var out bytes.Buffer
	fields := tmpl.Fields{}

	checksums := ctx.Artifacts.Filter(artifact.ByType(artifact.Checksum))

	checksumsList := checksums.List()
	switch len(checksumsList) {
	case 0:
		// do nothing
	case 1:
		bts, err := os.ReadFile(checksumsList[0].Path)
		if err != nil {
			return out, err
		}
		fields["Checksums"] = string(bts)
	default:
		checkMap := map[string]string{}
		for _, check := range checksumsList {
			bts, err := os.ReadFile(check.Path)
			if err != nil {
				return out, err
			}
			checkMap[artifact.ExtraOr(*check, artifact.ExtraChecksumOf, "")] = string(bts)
		}
		fields["Checksums"] = checkMap
	}

	t := tmpl.New(ctx).WithExtraFields(fields)

	headerStr, err := loadIncludedMarkdown(ctx.Config.Release.Header)
if err != nil {
return out, err
}
header, err := t.Apply(headerStr)
	if err != nil {
		return out, err
	}
	footerStr, err := loadIncludedMarkdown(ctx.Config.Release.Footer)
if err != nil {
return out, err
}
footer, err := t.Apply(footerStr)
	if err != nil {
		return out, err
	}

	bodyTemplate := template.Must(template.New("release").Parse(bodyTemplateText))
	err = bodyTemplate.Execute(&out, struct {
		Header       string
		Footer       string
		ReleaseNotes string
	}{
		Header:       header,
		Footer:       footer,
		ReleaseNotes: ctx.ReleaseNotes,
	})
	return out, err
}

func loadIncludedMarkdown(md config.IncludedMarkdown) (string, error) {
if md.Content != "" {
return md.Content, nil
}
rc, err := md.Load()
if err != nil {
return "", err
}
defer rc.Close()
bts, err := io.ReadAll(rc)
if err != nil {
return "", err
}
return string(bts), nil
}
