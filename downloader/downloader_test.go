package downloader

import (
	"log"
	"os"
)

func ExampleDownload() {
	dirname := "./tmp"
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		os.Mkdir(dirname, 0755)
	} else {
		os.RemoveAll(dirname)
		os.Mkdir(dirname, 0755)
	}
	err := Download("http://logo-tank.net/logo_data/13411.gif", "", dirname)
	if err != nil {
		log.Printf("error:%v", err)
	}
	// Output:
}
