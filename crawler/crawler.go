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
type ExampleExtender struct {
	gocrawl.DefaultExtender // Will use the default implementation of all but Visit and Filter
	ImageCh                 chan string
	URLPattern              *regexp.Regexp
	TotalCount              int32
}

// Override Visit for our need.
func (x *ExampleExtender) Visit(ctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) (interface{}, bool) {
	// Use the goquery document or res.Body to manipulate the data
	// ...

	log.Printf("visit:%s", ctx.URL())
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		if val := s.AttrOr("src", ""); val != "alternative" {
			if !strings.HasPrefix(val, "http") {
				val = ctx.NormalizedURL().Scheme + "://" + ctx.NormalizedURL().Host + "/" + val
			}
			x.ImageCh <- val
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
	return !isVisited && x.URLPattern.MatchString(ctx.NormalizedURL().String())
}

func ImageCrawl(seed string, urlPattern string, cb func(string, int32) error) {
	imageCh := make(chan string, 100)
	// Set custom options
	extender := new(ExampleExtender)
	extender.ImageCh = imageCh
	// Only enqueue the root and paths beginning with an "a"
	extender.URLPattern = regexp.MustCompile(urlPattern)
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
	opts.MaxVisits = 2

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

	c.Run(seed)
	close(imageCh)

	wg.Wait()
}
