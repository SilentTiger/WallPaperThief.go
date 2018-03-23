package downloader

import (
	"io"
)

// IDownloader 下载器接口，所有的下载器都需要实现此接口
type IDownloader interface {
	Start()
	Stop()
}

// DataItem 单条数据的格式
type DataItem struct {
	FileName string
	Data     io.Reader
}

// Downloader 下载器结构，所有的下载器都需要继承此结构
type Downloader struct {
	SubDirectory  string
	finishChannel chan<- int
	dataChannel   chan<- DataItem
	existPictures []string
}
