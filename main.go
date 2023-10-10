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
	selectFormat        string
	outputFilename      string
)

var (
	supportedFormat = []string{"flac", "mp3"}
	buildVersion    string
)

func init() {
	flag.StringVar(&baseDir, "dir", "", "待扫描音乐文件的目录，要用绝对路径")
	flag.StringVar(&selectFormat, "format", "", fmt.Sprintf("筛选特定格式的音乐文件，受支持的格式有：%s", strings.Join(supportedFormat, ", ")))
	flag.StringVar(&filterPrefix, "filter-prefix", "", "音乐文件的文件名的前缀过滤器")
	flag.StringVar(&filterSuffix, "filter-suffix", "", "音乐文件的文件名的后缀过滤器，请注意后缀过滤器会筛选包含文件名扩展名的内容")
	flag.StringVar(&filterContains, "filter-contains", "", "音乐文件的文件名的部分匹配过滤器")
	flag.StringVar(&filterTitleContains, "filter-title-contains", "", "音乐标题的部分匹配过滤器，这个过滤器会查询音乐文件的 Title Tag，如果音乐文件缺少 Title Tag 则会使用文件名来筛选")
	flag.StringVar(&outputFilename, "o", "", "播放列表输出文件的文件名，默认文件名是 playlist.m3u8，输入的内容如果缺少 .m3u8 则会自动补充")

	flag.Usage = func() {
		w := flag.CommandLine.Output()

		_, _ = fmt.Fprintf(w, `这是一个用来生成 Sony Walkman 的音乐播放列表的工具。
Version: %s

参数说明:
`, buildVersion)

		flag.PrintDefaults()

		_, _ = fmt.Fprintf(w, `
注意:
虽然我们提供了这四种过滤器【-filter-prefix】、【-filter-suffix】、【-filter-contains】和【-filter-title-contains】但是它们之间相互冲突，并且目前我暂时不打算支持混合过滤器。

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
		fmt.Println("筛选结束后没有可用的音乐文件，结束")
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
	fmt.Printf("完成，播放列表生成在 %s\n", outputPath)
}

type WalkResult struct {
	Path            string
	DurationSeconds int32
	Title           string
}

func validateParams() {
	if baseDir == "" {
		flag.Usage()
		panic("缺少待扫描文件夹路径")
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
		panic("你只能使用以下参数之一：-filter-prefix、-filter-suffix、-filter-contains 和 -filter-title-contains。不能同时使用两个以上的过滤参数。")
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

		// probe info
		data, err0 := ffprobe.ProbeURL(context.Background(), path)
		if err0 != nil {
			return nil // unsupported file
		}

		// filter supported formats, or select format
		formatName := strings.ToLower(data.Format.FormatName)
		if selectFormat != "" && formatName != strings.ToLower(selectFormat) {
			return nil
		} else if !slices.Contains(supportedFormat, formatName) {
			return nil
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
