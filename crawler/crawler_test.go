package crawler

import "log"

func ExampleCrawler() {
	imageCh := make(chan string, 10)

	go func() {
		for img := range imageCh {
			log.Printf("img:%s", img)
		}
	}()

	ImageCrawl("https://logo-tank.net", `http://logo-tank\.net(\?page=.*)?$`, imageCh)
	// Output:
}
