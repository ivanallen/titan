package downloader

import (
	"io"
	"net/http"
	"os"
	"strings"
)

// 有些网站没有 referer 是无法下载的。
func Download(url string, referer string, dirname string) error {
	parts := strings.Split(url, "/")
	fileName := parts[len(parts)-1]

	f, err := os.OpenFile(dirname+"/"+fileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return err
	}
	req.Header.Set("Referer", referer)

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	io.Copy(f, res.Body)
	return nil
}
