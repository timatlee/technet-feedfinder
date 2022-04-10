package technetblog

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

type TechnetBlog struct {
	Name        string
	Url         string
	FeedUrl     string
	Category    string
	bodyContent string
}

func (blog *TechnetBlog) getPageContent() {
	log.Debug("URL has a length, so we can try to parse.")
	log.Debug(fmt.Sprintf("Reading URL:%s", blog.Url))

	res, err := http.Get(blog.Url)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Read content fron the URL.")

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	log.Debug("Status code is 200, so reading content.")

	// May need to Tee the IO reader. https://golang.cafe/blog/how-to-read-multiple-times-from-an-io-reader-golang.html
	// blog.httpResponse = *res
	log.Debug("Reading body content.")
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	blog.bodyContent = string(content)

	log.Debug("Closing the body reader, since we have all content local now")
	defer res.Body.Close()
}

func (blog *TechnetBlog) populatFeedURL() {
	// content, err := ioutil.ReadAll(blog.httpResponse.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	contentString := blog.bodyContent

	r := regexp.MustCompile("href=\"(?P<URL>\\/gxcuf89792\\/rss\\/board\\?board\\.id=[\\w\\d-]+)\"><\\/link>")
	result := r.FindStringSubmatch(contentString)

	if len(result) == 0 {
		log.Fatalf("Could not find the feel URL in %s. Please check that it exists and handle this condition.", blog.Url)
	}

	log.Debug("Matches:")
	for k, v := range result {
		log.Debug(fmt.Sprintf("%d: %s\n", k, v))
	}

	// when using url.String(), it's encoding the path. We don't want this, so we're going to clear the path then manually add it. Gross.
	u, _ := url.Parse(blog.Url)
	u.Path = ""
	urlString := fmt.Sprintf("%s/%s", u.String(), result[1])

	log.Info(fmt.Sprintf("Found URL: %s", urlString))

	blog.FeedUrl = urlString
}

func (blog *TechnetBlog) populateFeedCategory() {
	reader := strings.NewReader(blog.bodyContent)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("a.crumb-category").Each(func(i int, s *goquery.Selection) {
		fmt.Println(s.Text())
		blog.Category = s.Text()
	})

}

func (blog *TechnetBlog) PopulateMembers() {
	if len(blog.Url) == 0 {
		return
	}
	blog.getPageContent()
	blog.populatFeedURL()
	blog.populateFeedCategory()
}
