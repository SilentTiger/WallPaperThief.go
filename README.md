# WallPaperThief.go

golang 版本的 WallPaperThief, 用来爬取特定网站上的壁纸。

主流程在 main.go 中，各网站的下载器需要继承自 Downloader 结构，参考 interfacelift.go 的代码。
下载器只用实现获取图片数据的逻辑，初始化目录以及写文件的操作 main.go 里面已经实现了。