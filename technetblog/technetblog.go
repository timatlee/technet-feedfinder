package technetblog

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	log "github.com/sirupsen/logrus"
)

type TechnetBlog struct {
	Name    string
	Url     string
	FeedUrl string
}

func (blog TechnetBlog) PopulateFeedUrl() {
	log.Debug("Determing the feed URL")
	if len(blog.Url) == 0 {
		return
	}

	log.Debug("URL has a length, so we can try to parse.")
	log.Debug(fmt.Sprintf("Reading URL:%s", blog.Url))

	res, err := http.Get(blog.Url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	log.Debug("Status code is 200, so reading content.")

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	contentString := string(content)
	log.Debug("contentString now contains the page's HTML content.")

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
