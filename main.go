package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"technetfeedfinder/technetblog"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/beevik/etree"
	"github.com/pbenner/threadpool"
	log "github.com/sirupsen/logrus"
)

var cacheFileJson string = "bloglistcache.json"
var opmlOutputFile string = "output/technetblogs.opml"
var threadpoolSize int = 4

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

	// Array to hold the blogs
	var blogsList []technetblog.TechnetBlog = make([]technetblog.TechnetBlog, 0)

	// Test if the file exists
	_, err := os.Stat(cacheFileJson)
	if errors.Is(err, os.ErrNotExist) {
		log.Info("Could not find the cache list. Generating a new one.")
		// Does not exist, so we need to go out the internet to build it.
		log.Info("Finding blogs on the technet site.")
		blogsList := getTechnetBlogs("https://techcommunity.microsoft.com/t5/custom/page/page-id/Blogs")
		log.Debug("Found blogs:")
		log.Debug(fmt.Sprintln(blogsList))

		pool := threadpool.New(threadpoolSize, threadpoolSize*25) // why not?
		g := pool.NewJobGroup()

		log.Info("Finished finding blogs. Parsing the pages to find category and feed URL.")
		for i := 0; i < len(blogsList); i++ {
			index := i
			pool.AddJob(g, func(pool threadpool.ThreadPool, erf func() error) error {
				log.Debug(fmt.Sprintf("Thread ID %d array index %d is blog name:%s url:%s **START**",
					pool.GetThreadId(),
					index,
					blogsList[index].Name,
					blogsList[index].Url,
				))
				blogsList[index].PopulateMembers()
				log.Debug(fmt.Sprintf("Thread ID %d array index %d is blog name:%s url:%s **END**",
					pool.GetThreadId(),
					index,
					blogsList[index].Name,
					blogsList[index].Url,
				))
				return nil
			})
		}
		log.Debug("Waiting for threads to complete.")
		pool.Wait(g)
		log.Debug("Threads done.")
		log.Info("Done finding category and feed URL's for each blog.  Dumping this to a cache file.")
		file, _ := json.MarshalIndent(blogsList, "", " ")
		_ = ioutil.WriteFile(cacheFileJson, file, 0644)
	} else {
		log.Info("Found a cache list. Using that. If you want a fresh download, delete 'bloglistcache.json'.")
		jsonFile, err := os.Open(cacheFileJson)
		if err != nil {
			log.Fatal(err)
		}
		byteValue, _ := ioutil.ReadAll(jsonFile)
		json.Unmarshal(byteValue, &blogsList)

		defer jsonFile.Close()
	}

	// Convert the list of blogs into a map[category]blog.
	log.Debug("Start work torwards generating the OPML file")
	log.Debug("Convert the list into a map of string -> []TechnetBlog")
	blogsMap := make(map[string][]technetblog.TechnetBlog)
	for i := 0; i < len(blogsList); i++ {
		blogsMap[blogsList[i].Category] = append(blogsMap[blogsList[i].Category], blogsList[i])
	}

	// Generate OPML file
	log.Debug("Generate the OPML")
	generateOPMLFile(blogsMap, opmlOutputFile)
}

func generateOPMLFile(blogs map[string][]technetblog.TechnetBlog, filepath string) {
	currentTime := time.Now()
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="ISO-8859-1"`)
	opml := doc.CreateElement("opml")
	opml.CreateAttr("version", "2.0")

	head := opml.CreateElement("head")
	title := head.CreateElement("Title")
	title.SetText("Microsoft Tech Community Blogs")
	dateCreated := head.CreateElement("dateCreated")
	dateCreated.SetText(currentTime.Format(time.RFC822))
	body := opml.CreateElement("body")

	// Get a list of keys from the map, and sort themm
	keys := make([]string, 0, len(blogs))
	for k := range blogs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// With the sorted list of keys, get each key and it's values from the map.
	for i := 0; i < len(keys); i++ {
		catOutline := body.CreateElement("outline")
		catOutline.CreateAttr("text", keys[i])

		categoryBlogs := blogs[keys[i]]
		for j := 0; j < len(categoryBlogs); j++ {
			catBlog := catOutline.CreateElement("outline")
			catBlog.CreateAttr("type", "rss")
			catBlog.CreateAttr("text", categoryBlogs[j].Name)
			catBlog.CreateAttr("xmlUrl", categoryBlogs[j].FeedUrl)
			catBlog.CreateAttr("htmlUrl", categoryBlogs[j].Url)
		}

	}

	doc.Indent(2)
	doc.WriteToFile(filepath)
}

func getTechnetBlogs(rootURL string) []technetblog.TechnetBlog {
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

	var blogsList []technetblog.TechnetBlog = make([]technetblog.TechnetBlog, 0)

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
			b := technetblog.TechnetBlog{Name: linkTitle, Url: url.String()}
			blogsList = append(blogsList, b)
		}
	})

	return blogsList
}
