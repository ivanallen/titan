package downloader

import (
	"io"
	"net/http"
	"os"
	"strings"
)

func Download(url string, dirname string) error {
	parts := strings.Split(url, "/")
	fileName := parts[len(parts)-1]

	f, err := os.OpenFile(dirname+"/"+fileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer f.Close()
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	io.Copy(f, res.Body)
	return nil
}
