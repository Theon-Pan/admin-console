package static // import "admin-console/internal/ui/static"
import (
	"admin-console/internal/crypto"
	"bytes"
	"embed"
	"log/slog"
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/svg"
)

type asset struct {
	Data     []byte
	Checksum string
}

// Static asserts.
var (
	StylesheetBundles map[string]asset
	JavascriptBundles map[string]asset
	BinaryBundles     map[string]asset
)

//go:embed bin/*
var binaryFiles embed.FS

//go:embed css/*.css
var stylesheetFiles embed.FS

//go:embed js/*.js
var javascriptFiles embed.FS

func GenerateBinaryBundles() error {
	dirEntries, err := binaryFiles.ReadDir("bin")
	if err != nil {
		return err
	}
	BinaryBundles = make(map[string]asset, len(dirEntries))

	minifier := minify.New()
	minifier.Add("image/svg+xml", &svg.Minifier{
		KeepComments: true, // needed to keep the license
	})

	for _, dirEntry := range dirEntries {
		name := dirEntry.Name()
		data, err := binaryFiles.ReadFile("bin/" + name)
		if err != nil {
			return err
		}

		if strings.HasSuffix(name, ".svg") {
			// minifier.Bytes returns the data unchanged in case of error.
			data, err = minifier.Bytes("image/svg+xml", data)
			if err != nil {
				slog.Error("Unable to minimize the svg file", slog.String("filename", name), slog.Any("error", err))
			}
		}

		BinaryBundles[name] = asset{
			Data:     data,
			Checksum: crypto.HashFromBytes(data),
		}
	}

	return nil
}

// GenerateStylesheetBundles creates CSS bundles.
func GenerateStylesheetBundles() error {
	var bundles = map[string][]string{
		"light_serif":       {"css/light.css", "css/serif.css", "css/common.css"},
		"light_sans_serif":  {"css/light.css", "css/sans_serif.css", "css/common.css"},
		"dark_serif":        {"css/dark.css", "css/serif.css", "css/common.css"},
		"dark_sans_serif":   {"css/dark.css", "css/sans_serif.css", "css/common.css"},
		"system_serif":      {"css/system.css", "css/serif.css", "css/common.css"},
		"system_sans_serif": {"css/sytem.css", "css/sans_serif.css", "css/common.css"},
	}

	StylesheetBundles = make(map[string]asset, len(bundles))

	minifier := minify.New()
	minifier.AddFunc("text/css", css.Minify)

	for bundleName, srcFiles := range bundles {
		var buffer bytes.Buffer

		for _, srcFile := range srcFiles {
			fileData, err := stylesheetFiles.ReadFile(srcFile)
			if err != nil {
				return err
			}
			buffer.Write(fileData)
		}
		minifiedData, err := minifier.Bytes("text/css", buffer.Bytes())
		if err != nil {
			return err
		}

		StylesheetBundles[bundleName+".css"] = asset{
			Data:     minifiedData,
			Checksum: crypto.HashFromBytes(minifiedData),
		}
	}

	return nil
}
