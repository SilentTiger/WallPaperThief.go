package main

import (
	"WallPaperThief/downloader"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"sync"

	"WallPaperThief/logger"
)

const rootPath = "/Users/jinweiliu/Pictures/wallpaper/"

func main() {

	runtime.GOMAXPROCS(2)

	var finishChannel = make(chan int)
	var dataChannel = make(chan downloader.DataItem, 100)

	directoryStatus := true
	directoryStatus = directoryStatus && initDirectories(rootPath+"a/")
	directoryStatus = directoryStatus && initDirectories(rootPath+"b/")

	if !directoryStatus {
		logger.Info("init directories failed, exit.")
		return
	}

	logger.Info("start init downloaders")
	var downloaders []downloader.IDownloader
	res, err := downloaderFactory("Interfacelift", "a/", finishChannel, dataChannel, listFile(rootPath+"a/"))
	if err == nil {
		downloaders = append(downloaders, res)
	}

	res, err = downloaderFactory("Interfacelift", "b/", finishChannel, dataChannel, listFile(rootPath+"b/"))
	if err == nil {
		downloaders = append(downloaders, res)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		logger.Info("start goroutine write")

		defer func() {
			wg.Done()
			logger.Info("goroutine defer")
		}()

		for dataItem := range dataChannel {
			logger.Info("start write " + dataItem.FileName)
			logger.Info("current data length " + strconv.Itoa(len(dataChannel)))
			writeFile(rootPath+dataItem.FileName, dataItem.Data)
			logger.Info("finish write " + dataItem.FileName)
		}
		logger.Info("finish goroutine write")
	}()

	logger.Info("start all downloaders")
	for _, d := range downloaders {
		go d.Start()
	}

	finishCount := 0
	for {
		if finishCount < len(downloaders) {
			<-finishChannel
			finishCount++
			logger.Info("one downloader finished")
		} else {
			// 关闭 datachannel
			close(dataChannel)
			break
		}
	}

	wg.Wait()
}

// downloaderFactory 下载器工厂方法
func downloaderFactory(downloaderType string, subDirectory string, finishChannel chan int, dataChannel chan downloader.DataItem, existPictures []string) (res downloader.IDownloader, err error) {
	switch downloaderType {
	case "Interfacelift":
		res = downloader.NewInterfacelift(subDirectory, finishChannel, dataChannel, existPictures)
		break
	default:
		err = errors.New("wrong downloaderType value")
		logger.Error("Init downloader error:")
		logger.Error(err)
	}

	return res, err
}

// 初始化目录
func initDirectories(path string) bool {
	logger.Info("start check save path")
	pathInfo, err := os.Stat(path)
	if err == nil {
		if !pathInfo.IsDir() {
			logger.Error("save path is exists, and it's not a directory.")
			return false
		}
	} else {
		if os.IsNotExist(err) {
			err := os.MkdirAll(path, 0777)
			if err != nil {
				logger.Error("create path error: " + err.Error())
				return false
			}
		} else {
			logger.Error("get save path info error: " + err.Error())
			return false
		}
	}

	return true
}

func writeFile(absoluteFilename string, data []byte) {
	file, err := os.Create(absoluteFilename)
	if err != nil {
		logger.Error("create file " + absoluteFilename + " error: " + err.Error())
		return
	}

	_, err = io.Copy(file, bytes.NewReader(data))
	if err != nil {
		logger.Error("write file " + absoluteFilename + " error: " + err.Error())
		return
	}
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
