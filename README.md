# go-walkman-playlist-gen

Sony Walkman 播放列表生成工具。

目前这个工具只支持 flac 和 mp3，因为这两个格式的音乐文件是最有可能包含 TAG 数据的。

## 下载

前往 [Release](https://github.com/AyakuraYuki/go-walkman-playlist-gen/releases) 页面下载可执行文件。

注意，下载项中的 `386` 版本未经测试，会有意料之外的情况，并且目前本程序不对 32 位操作系统做技术支持。建议使用 64 位操作系统。

## 使用方式

```text
参数说明:
  -dir string
        待扫描音乐文件的目录，要用绝对路径
  -filter-contains string
        音乐文件的文件名的部分匹配过滤器
  -filter-prefix string
        音乐文件的文件名的前缀过滤器
  -filter-suffix string
        音乐文件的文件名的后缀过滤器，请注意后缀过滤器会筛选包含文件名扩展名的内容
  -filter-title-contains string
        音乐标题的部分匹配过滤器，这个过滤器会查询音乐文件的 Title Tag，如果音乐文件缺少 Title Tag 则会使用文件名来筛选
  -format string
        筛选特定格式的音乐文件，受支持的格式有：flac, mp3
  -o string
        播放列表输出文件的文件名，默认文件名是 playlist.m3u8，输入的内容如果缺少 .m3u8 则会自动补充
```

注意：虽然我们提供了这四种过滤器【-filter-prefix】、【-filter-suffix】、【-filter-contains】和【-filter-title-contains】但是它们之间相互冲突，并且目前我暂时不打算支持混合过滤器。

## 关于生成播放列表的顺序问题

目前，这个工具暂时不提供排序功能。相较于其他结构化数据，音乐的排序有比较多的维度可以参考，例如使用发行年，或者光盘编号配合歌曲编号，又或者有些烧友偏爱使用文件名排序，这些都有很大的人为因素掺入其中。

m3u8 文件天然是可编辑的文本文件，特别是针对 Walkman 的播放列表，它的播放列表格式更是简单，仅用两行确定一首音乐的信息，所以支持手动进行自定义排序。

所以这里我建议使用文本编辑器打开生成的播放列表，调整文本顺序来进行播放列表的排序工作。

需要移动的数据大概长得像下面的这两行文本：

```m3u
#EXTINF:114;先辈的鸣奏曲
田所浩二\先辈的鸣奏曲.mp3
```

这两行文本的特征是，它们第一行以 `#EXTINF` 开头，第二行是相对于 m3u8 文件的音乐文件路径（这个路径不需要做过多理解，只要保持不变就行了）。
