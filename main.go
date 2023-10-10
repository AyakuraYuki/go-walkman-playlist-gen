package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/flytam/filenamify"
	"github.com/spf13/cast"
	"gopkg.in/vansante/go-ffprobe.v2"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var (
	baseDir             string
	filterPrefix        string
	filterSuffix        string
	filterContains      string
	filterTitleContains string
	selectExtension     string
	outputFilename      string
)

var (
	supportedExtension = []string{"flac", "mp3"}
	buildVersion       string
)

func init() {
	flag.StringVar(&baseDir, "dir", "", "an absolute path of directory to scan music")
	flag.StringVar(&selectExtension, "ext", "", fmt.Sprintf("select format for scanning music, supported extensions are %s", strings.Join(supportedExtension, ", ")))
	flag.StringVar(&filterPrefix, "filter-prefix", "", "a filter for prefixes of music filename")
	flag.StringVar(&filterSuffix, "filter-suffix", "", "a filter for suffixes of music filename, be warned this filter includes file extension")
	flag.StringVar(&filterContains, "filter-contains", "", "a filter for music filename containing partial string")
	flag.StringVar(&filterTitleContains, "filter-title-contains", "", "a filter for music title containing partial string, some music file doesn't present title tag, use filename")
	flag.StringVar(&outputFilename, "o", "", "output filename, the output file will be put into the scanning directory")

	flag.Usage = func() {
		w := flag.CommandLine.Output()

		_, _ = fmt.Fprintf(w, `A generator of Sony Walkman .m3u8 music playlist.
Version: %s

Flags:
`, buildVersion)

		flag.PrintDefaults()

		_, _ = fmt.Fprintf(w, `
Attention:
You can only use one of these four filters: -filter-prefix, -filter-suffix, -filter-contains, and -filter-title-contains.
They conflict with each other, and we don't currently support the use of mixed filters.

`)
	}

	flag.Parse()
	validateParams()

	if outputFilename == "" {
		outputFilename = "playlist.m3u8"
	} else if !strings.HasSuffix(outputFilename, ".m3u8") {
		outputFilename = fmt.Sprintf("%s.m3u8", outputFilename)
	}
	safetyFilename, _ := filenamify.Filenamify(outputFilename, filenamify.Options{Replacement: "_"})
	if safetyFilename != "" {
		outputFilename = safetyFilename
	}
}

func main() {
	res := walkBaseDir()
	if len(res) == 0 {
		fmt.Println("no music files found after filtering, exit")
		os.Exit(1)
	}

	m3u8ContentBuilder := strings.Builder{}
	m3u8ContentBuilder.WriteString("#EXTM3U\n")
	for _, v := range res {
		copiedV := v
		m3u8ContentBuilder.WriteString(fmt.Sprintf("#EXTINF:%d;%s\n", copiedV.DurationSeconds, copiedV.Title))
		m3u8Path := strings.ReplaceAll(copiedV.Path, baseDir, "")
		m3u8Path = strings.TrimPrefix(m3u8Path, string(os.PathSeparator))
		m3u8ContentBuilder.WriteString(fmt.Sprintf("%s\n", m3u8Path))
	}
	m3u8ContentBuilder.WriteString("\n")

	outputPath := filepath.Join(baseDir, outputFilename)
	if err := os.WriteFile(outputPath, []byte(m3u8ContentBuilder.String()), os.ModePerm); err != nil {
		panic(err)
	}
	fmt.Printf("done, %s has been generated\n", outputPath)
}

type WalkResult struct {
	Path            string
	DurationSeconds int32
	Title           string
}

func validateParams() {
	if baseDir == "" {
		flag.Usage()
		panic("missing base dir")
	}

	filterCount := 0
	if len(filterPrefix) > 0 {
		filterCount++
	}
	if len(filterSuffix) > 0 {
		filterCount++
	}
	if len(filterContains) > 0 {
		filterCount++
	}
	if len(filterTitleContains) > 0 {
		filterCount++
	}
	if filterCount > 1 {
		panic("You can only use one of these parameters: -filter-prefix, -filter-suffix, -filter-contains, and -filter-title-contains. You can't use two or more filtering parameters simultaneously.")
	}
}

func walkBaseDir() (res []*WalkResult) {
	res = make([]*WalkResult, 0)
	err := filepath.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		// skip dir
		if info.IsDir() {
			return nil
		}

		extension := strings.ToLower(filepath.Ext(info.Name()))
		extensionWithoutDot := strings.TrimPrefix(extension, ".")

		// filter by prefix
		if filterPrefix != "" && !strings.HasPrefix(info.Name(), filterPrefix) {
			return nil
		}

		// filter by suffix
		if filterSuffix != "" && !strings.HasSuffix(info.Name(), filterSuffix) {
			return nil
		}

		// filter by filename contains
		if filterContains != "" && !strings.Contains(info.Name(), filterContains) {
			return nil
		}

		// filter supported extensions, or select extension
		if selectExtension != "" && extensionWithoutDot != strings.ToLower(selectExtension) {
			return nil
		} else if !slices.Contains(supportedExtension, extensionWithoutDot) {
			return nil
		}

		// probe info
		data, err0 := ffprobe.ProbeURL(context.Background(), path)
		if err0 != nil {
			panic(err0)
		}
		title, _ := data.Format.TagList.GetString("TITLE")
		if title == "" {
			title = strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
		}

		// filter by music title tag
		if filterTitleContains != "" && !strings.Contains(title, filterTitleContains) {
			return nil
		}

		res = append(res, &WalkResult{
			Path:            path,
			DurationSeconds: cast.ToInt32(math.Round(data.Format.DurationSeconds)),
			Title:           title,
		})
		return nil
	})
	if err != nil {
		panic(err)
	}
	return res
}
