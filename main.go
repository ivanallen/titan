package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"titan/archiver"
	"titan/crawler"
	"titan/downloader"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Static("/static", "./static")
	r.Static("/assets", "./assets")
	r.Static("/resource", "./resource")
	r.POST("/crawlimage", func(c *gin.Context) {
		var opts crawler.CrawlImageOption
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

		crawlImage(&opts, func(imgInfo *crawler.ImageInfo, index int32, total int32) error {
			part, err := multipartWriter.CreatePart(header)
			if err != nil {
				return err
			}
			report := fmt.Sprintf(`{"img":"%s", "from":"%s", "index":%d, "total":%d}`, imgInfo.ImageURL, imgInfo.FromURL, index, total)
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

	r.POST("/crawlimage2", func(c *gin.Context) {
		var opts crawler.CrawlImageOption
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
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.WriteHeader(http.StatusOK)

		if f, ok := c.Writer.(http.Flusher); ok {
			f.Flush()
		} else {
			log.Print("Damn, no flush")
		}

		crawlImage(&opts, func(imgInfo *crawler.ImageInfo, index int32, total int32) error {
			report := fmt.Sprintf(`{"img":"%s", "from":"%s", "index":%d, "total":%d}`, imgInfo.ImageURL, imgInfo.FromURL, index, total)

			// Write 可能不会完全写出。需要 Writen
			c.Writer.Write([]byte("event: progress\n"))
			c.Writer.Write([]byte("data: " + report + "\n\n"))

			log.Printf("report:%v", report)

			if f, ok := c.Writer.(http.Flusher); ok {
				f.Flush()
			} else {
				log.Print("Damn, no flush")
			}
			return nil
		})

		c.Writer.Write([]byte("event: end\n"))
		c.Writer.Write([]byte(`data: {"download_url:"/resource/tmp.zip"}` + "\n\n"))
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}

func crawlImage(opts *crawler.CrawlImageOption, report func(*crawler.ImageInfo, int32, int32) error) {

	dirname := "tmp"

	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		os.Mkdir(dirname, 0755)
	} else {
		os.RemoveAll(dirname)
		os.Mkdir(dirname, 0755)
	}

	var processedCount int32
	crawler.ImageCrawl(opts, func(img *crawler.ImageInfo, totalCount int32) error {
		downloader.Download(img.ImageURL, img.FromURL, dirname)
		processedCount += 1
		if err := report(img, processedCount, totalCount); err != nil {
			return err
		}
		return nil
	})

	if _, err := os.OpenFile("./resource/"+dirname+".zip", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600); os.IsExist(err) {
		os.Remove("./resource/" + dirname + ".zip")
	}

	err := archiver.Zip(dirname, "./resource/"+dirname+".zip")
	if err != nil {
		log.Printf("zip failed:%v", err)
	}
}
