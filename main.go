package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"tools/archiver"
	"tools/crawler"
	"tools/downloader"
)

type CrawlImageOption struct {
	Seed    string `json:"seed" binding:"required"`
	Pattern string `json:"pattern" binding:"required"`
}

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Static("/static", "./static")
	r.Static("/assets", "./assets")
	r.POST("/crawlimage", func(c *gin.Context) {
		var opts CrawlImageOption
		if err := c.ShouldBindJSON(&opts); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Printf(`request:{"seed":"%v","pattern":"%v"}`, opts.Seed, opts.Pattern)
		/*
			opts := &CrawlImageOption{
				Seed:    "https://logo-tank.net",
				Pattern: `http://logo-tank\.net(\?page=.*)?$`,
			}
		*/
		multipartWriter := multipart.NewWriter(c.Writer)
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="metadata"`)
		header.Set("Content-Type", "application/json; charset=UTF-8")

		c.Writer.Header().Set("Content-Type", multipartWriter.FormDataContentType())
		c.Writer.WriteHeader(http.StatusOK)

		if f, ok := c.Writer.(http.Flusher); ok {
			f.Flush()
		} else {
			log.Print("Damn, no flush")
		}

		crawlImage(&opts, func(img string, index int32, total int32) error {
			part, err := multipartWriter.CreatePart(header)
			if err != nil {
				return err
			}
			report := fmt.Sprintf(`{"img":"%s", "index":%d, "total":%d}`, img, index, total)
			part.Write([]byte(report))
			log.Printf("report:%v", report)

			if f, ok := c.Writer.(http.Flusher); ok {
				f.Flush()
			} else {
				log.Print("Damn, no flush")
			}
			return nil
		})
		multipartWriter.Close()
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}

func crawlImage(opts *CrawlImageOption, report func(string, int32, int32) error) {

	dirname := "./tmp"

	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		os.Mkdir(dirname, 0755)
	} else {
		os.RemoveAll(dirname)
		os.Mkdir(dirname, 0755)
	}

	var processedCount int32
	crawler.ImageCrawl(opts.Seed, opts.Pattern, func(img string, totalCount int32) error {
		downloader.Download(img, dirname)
		processedCount += 1
		if err := report(img, processedCount, totalCount); err != nil {
			return err
		}
		return nil
	})

	if _, err := os.OpenFile(dirname+".zip", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600); os.IsExist(err) {
		os.Remove(dirname + ".zip")
	}

	err := archiver.Zip(dirname, dirname+".zip")
	if err != nil {
		log.Printf("zip failed:%v", err)
	}
}
