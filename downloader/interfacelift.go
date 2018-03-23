package downloader

import (
	"WallPaperThief/logger"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const pageURL = "https://interfacelift.com/wallpaper/downloads/date/wide_16:9/2560x1440/index5.html"

// Interfacelift 下载器
type Interfacelift struct {
	Downloader
}

// NewInterfacelift Interfacelift构造函数
func NewInterfacelift(subDirectory string, finishChannel chan<- int, dataChannel chan<- DataItem, existPictures []string) Interfacelift {
	res := Interfacelift{}
	res.SubDirectory = subDirectory
	res.finishChannel = finishChannel
	res.dataChannel = dataChannel
	res.existPictures = existPictures
	return res
}

// Start 开始下载图片
func (instance Interfacelift) Start() {
	defer func() {
		instance.finishChannel <- 1
	}()

	logger.Info("start interfacelift")

	doc, err := goquery.NewDocument(pageURL)
	if err != nil {
		logger.Error("get page error: " + err.Error())
		return
	}

	urls := searchURL(doc)

	urlsToDownload := filterExistURL(instance.existPictures, urls)

	instance.batDownload(urlsToDownload)
}

// Stop 中断下载图片
func (instance Interfacelift) Stop() {
	instance.finishChannel <- 1
	// todo
}

// searchURL 在html文档中搜索用于下载的url.
func searchURL(document *goquery.Document) []string {
	rootNodeDoc := goquery.NewDocumentFromNode(document.Find("#wallpaper").Get(1))

	return rootNodeDoc.Find("div[id]").FilterFunction(func(_ int, node *goquery.Selection) bool {
		itemID, _ := node.Attr("id")
		return strings.Contains(itemID, "download_")
	}).Map(func(_ int, node *goquery.Selection) string {
		href, exist := node.Children().Eq(0).Attr("href")
		if !exist {
			logger.Error("can not get href")
			return ""
		}
		return href
	})
}

// filterExistURL 过滤掉已经下载过的 url
func filterExistURL(fileList []string, urls []string) []string {
	for index := len(urls) - 1; index >= 0; index-- {
		for _, filename := range fileList {
			if strings.Index(urls[index], filename) != -1 {
				urls = append(urls[:index], urls[index+1:]...)
				break
			}
		}
	}
	return urls
}

// batDownload 开始下载任务
func (instance *Interfacelift) batDownload(urls []string) {
	for _, url := range urls {
		res, err := instance.download(url)
		if err == nil {
			instance.dataChannel <- res
		}
	}
}

// download 下载 url 所指定的图片
func (instance *Interfacelift) download(url string) (dataItem DataItem, err error) {
	logger.Info("start fetch " + "https://interfacelift.com" + url)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://interfacelift.com"+url, nil)
	req.Header.Set("Host", "interfacelift.com")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.186 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7,ja;q=0.6")
	res, err := client.Do(req)

	if err != nil {
		logger.Error("fetch " + url + " error: " + err.Error())
		return
	}

	tmp := strings.Split(url, "/")
	dataItem.FileName = instance.SubDirectory + tmp[len(tmp)-1]
	dataItem.Data = res.Body

	return
}
