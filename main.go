package main

import (
	"hello_go/logger"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const pageURL = "https://interfacelift.com/wallpaper/downloads/date/wide_16:9/2560x1440/index5.html"
const savePath = "/Users/jinweiliu/Pictures/wallpaper/"

func main() {
	logger.Log("start check save path")
	pathInfo, err := os.Stat(savePath)
	if err == nil {
		if !pathInfo.IsDir() {
			logger.Error("save path is exists, and it's not a directory.")
			return
		}
	} else {
		if os.IsNotExist(err) {
			err := os.MkdirAll(savePath, 0777)
			if err != nil {
				logger.Error("create path error: " + err.Error())
				return
			}
		} else {
			logger.Error("get save path info error: " + err.Error())
			return
		}
	}

	logger.Log("start get pictures.")
	doc, err := goquery.NewDocument(pageURL)
	if err != nil {
		logger.Error("get page error: " + err.Error())
		return
	}

	urls := searchURL(doc)

	files := listFile(savePath)

	urlsToDownload := filterExistURL(files, urls)

	dispatchDownloadTask(urlsToDownload)
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

// listFile 列出已经下载的所有文件的文件名
func listFile(path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		logger.Error("list file error: " + err.Error())
		return nil
	}

	var nameList []string
	for _, v := range files {
		if !v.IsDir() {
			nameList = append(nameList, v.Name())
		}
	}
	return nameList
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

// dispatchDownloadTask 分发下载任务
func dispatchDownloadTask(urls []string) {
	runtime.GOMAXPROCS(len(urls))

	counter := make(chan int)

	for _, url := range urls {
		go download(url, counter)
		<-counter // 这里特意写成顺序下载，因为这个网站一旦开并行就挂了，估计他们是故意的
	}
}

// download 下载 url 所指定的图片
func download(url string, chanel chan int) {
	defer func() {
		chanel <- 0 // 下载完成或者出错后通知主线程
	}()

	logger.Log("start fetch " + "https://interfacelift.com" + url)
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
	filename := tmp[len(tmp)-1]
	file, err := os.Create(savePath + filename)
	if err != nil {
		logger.Error("create file " + filename + " error: " + err.Error())
		return
	}

	written, err := io.Copy(file, res.Body)
	if err != nil {
		logger.Error("write file " + filename + " error: " + err.Error())
		return
	}

	logger.Log("save status " + strconv.FormatInt(written, 10))
}
