package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	log.Info("Hello. Starting the program.")

	blogsList := getTechnetBlogs("https://techcommunity.microsoft.com/t5/custom/page/page-id/Blogs")
	log.Debug("Found blogs:")
	log.Debug(fmt.Sprintln(blogsList))

	for i := 0; i < len(blogsList); i++ {
		blogsList[i].populateFeedUrl()
	}
}

func getTechnetBlogs(rootURL string) []TechnetBlog {
	res, err := http.Get(rootURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	u, err := url.Parse(rootURL)
	if err != nil {
		log.Fatal(err)
	}

	var blogsList []TechnetBlog = make([]TechnetBlog, 1)

	doc.Find(".blogs-all-list li").Each(func(i int, s *goquery.Selection) {
		linkTitle := s.Find("a").Text()
		linkUrl, _ := s.Find("a").Attr("href")
		log.Debug(fmt.Sprintf("Row: %d: Title: %s URL: %s", i, linkTitle, linkUrl))

		if len(linkTitle) > 0 && len(linkUrl) > 0 {
			url := url.URL{
				Scheme: u.Scheme,
				Host:   u.Host,
				Path:   linkUrl,
			}
			b := TechnetBlog{Name: linkTitle, Url: url.String()}
			blogsList = append(blogsList, b)
		}
	})

	return blogsList
}

/*
 Trying to get a list of all the communities by reading the website, then parsing it all out
 We also found the JSON endpoint at https://techcommunity.microsoft.com/plugins/custom/microsoft/o365/filter-hubs?allhubs=true&sortBy=recent
 but still only feeds back the first 50 communities.

func getTechnetCommunities(rootURL string) string {
	resp, err := http.Get(rootURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(content)
}
*/

type TechnetBlog struct {
	Name    string
	Url     string
	FeedUrl string
}

func (blog TechnetBlog) populateFeedUrl() {
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
