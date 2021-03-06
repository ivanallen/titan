package crawler

import (
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/gocrawl"
	"github.com/PuerkitoBio/goquery"
)

// Create the Extender implementation, based on the gocrawl-provided DefaultExtender,
// because we don't want/need to override all methods.

type CrawlImageOption struct {
	Seed      string `json:"seed" binding:"required"`
	Pattern   string `json:"pattern" binding:"required"`
	MaxVisits int32  `json:"max_visits"`
}

type ImageInfo struct {
	ImageURL string // 图片地址
	FromURL  string // 图片来源
}
type ExampleExtender struct {
	gocrawl.DefaultExtender // Will use the default implementation of all but Visit and Filter
	ImageCh                 chan *ImageInfo
	URLPattern              *regexp.Regexp
	TotalCount              int32
	Images                  map[string]struct{}
}

// Override Visit for our need.
func (x *ExampleExtender) Visit(ctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) (interface{}, bool) {
	// Use the goquery document or res.Body to manipulate the data
	// ...

	if ctx == nil || res == nil || doc == nil {
		log.Printf("ctx, res or doc is nil")
		return nil, true
	}
	log.Printf("visit:%s", ctx.URL())
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		if val := s.AttrOr("src", ""); val != "alternative" {
			if !strings.HasPrefix(val, "http") {
				val = ctx.NormalizedURL().Scheme + "://" + ctx.NormalizedURL().Host + "/" + val
			}
			if _, ok := x.Images[val]; ok {
				log.Printf("repeated, ignore:" + val)
				return
			}
			x.Images[val] = struct{}{}
			x.ImageCh <- &ImageInfo{
				val,
				ctx.URL().String(),
			}
			x.TotalCount += 1
		} else {
			log.Printf("error")
		}
	})
	// Return nil and true - let gocrawl find the links
	return nil, true
}

// Override Filter for our need.
func (x *ExampleExtender) Filter(ctx *gocrawl.URLContext, isVisited bool) bool {
	log.Printf("URL:%s", ctx.URL())
	log.Printf("NormalizedURL:%s", ctx.NormalizedURL())
	log.Printf("SourceURL:%s", ctx.SourceURL())
	log.Printf("NormalizedSourceURL:%s", ctx.NormalizedSourceURL())
	return !isVisited && x.URLPattern.MatchString(ctx.NormalizedURL().String())
}

func ImageCrawl(imgOpts *CrawlImageOption, cb func(*ImageInfo, int32) error) {
	imageCh := make(chan *ImageInfo, 100)
	// Set custom options
	extender := new(ExampleExtender)
	extender.ImageCh = imageCh
	extender.Images = make(map[string]struct{})
	// Only enqueue the root and paths beginning with an "a"
	extender.URLPattern = regexp.MustCompile(imgOpts.Pattern)
	opts := gocrawl.NewOptions(extender)

	// should always set your robot name so that it looks for the most
	// specific rules possible in robots.txt.
	opts.RobotUserAgent = "Cat"
	// and reflect that in the user-agent string used to make requests,
	// ideally with a link so site owners can contact you if there's an issue
	opts.UserAgent = "Mozilla/5.0 (compatible; Cat/1.0; +http://cat.com)"

	opts.CrawlDelay = 1 * time.Second
	opts.LogFlags = gocrawl.LogAll

	// Play nice with ddgo when running the test!
	opts.MaxVisits = 1

	if imgOpts.MaxVisits != 0 {
		opts.MaxVisits = int(imgOpts.MaxVisits)

	}

	// Create crawler and start at root of duckduckgo
	c := gocrawl.NewCrawlerWithOptions(opts)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for img := range imageCh {
			if err := cb(img, extender.TotalCount); err != nil {
				log.Printf("cb error:%v", err)
				c.Stop()
				break
			}
		}
		wg.Done()
	}()

	c.Run(imgOpts.Seed)
	close(imageCh)

	wg.Wait()
}
