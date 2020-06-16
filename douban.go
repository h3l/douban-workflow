package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"gopkg.in/resty.v1"
)

type UrlItem struct {
	Url string
	Cat string
}

type SearchResultItem struct {
	Title         string
	OriginScore   string
	Url           string
	FullStarCount int
	HalfStarCount int
	ImgUrl        string
}

type AlfredItem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Arg      string `json:"arg"`
	Icon     struct {
		Path string `json:"path"`
	} `json:"icon"`
}

var urlMapping = map[string]UrlItem{
	"book": {
		Url: "https://m.douban.com/search/?type=book&query=%s",
		Cat: "1001",
	},
	"movie": {
		Url: "https://m.douban.com/search/?type=movie&query=%s",
		Cat: "1002",
	},
}

func getNodeAttr(node *html.Node, attrName string) string {
	for _, a := range node.Attr {
		if a.Key == attrName {
			return a.Val
		}
	}
	return ""
}

func getItems(searchType string, searchString string) *[]SearchResultItem {
	if v, ok := urlMapping[searchType]; ok {
		resp, _ := resty.R().Get(fmt.Sprintf(v.Url, searchString))
		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
		s := doc.Find("ul.search_results_subjects > li")
		var node *goquery.Document
		var href, originScore, title, imgUrl string
		var fullStar, halfStar int
		r := make([]SearchResultItem, 0)
		for _, n := range s.Nodes {
			node = goquery.NewDocumentFromNode(n)
			href = getNodeAttr(node.Find("a").Nodes[0], "href")
			href = strings.ReplaceAll(href, "/"+searchType, "")

			originScore = node.Find("a > div > p > span").Text()
			title = node.Find("a > div > span").Text()
			imgUrl, _ = node.Find("a > img").Attr("src")
			fullStar = len(node.Find(".rating-star-small-full").Nodes)
			halfStar = len(node.Find(".rating-star-small-half").Nodes)
			r = append(r, SearchResultItem{
				Title:         title,
				OriginScore:   originScore,
				Url:           href,
				FullStarCount: fullStar,
				HalfStarCount: halfStar,
				ImgUrl:        imgUrl,
			})
		}
		return &r
	}
	return nil
}

var wg sync.WaitGroup

func downloadFile(path, url string) {
	out, _ := os.Create(path)
	defer out.Close()
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	io.Copy(out, resp.Body)
	wg.Done()
}

func getImgPath(url string) string {
	ps := strings.Split(url, "/")
	fn := ps[len(ps)-1]
	return fmt.Sprintf("/tmp/%v", fn)
}

func generateResponse(items *[]SearchResultItem, searchType string) {
	baseUrl := fmt.Sprintf("https://%s.douban.com", searchType)
	r := make([]AlfredItem, 0)
	for _, i := range *items {
		wg.Add(1)
		path := getImgPath(i.ImgUrl)
		go downloadFile(path, i.ImgUrl)
	}

	for _, i := range *items {
		path := getImgPath(i.ImgUrl)
		r = append(r, AlfredItem{
			Type:     "file",
			Title:    i.Title,
			Subtitle: strings.Repeat("⭐", i.FullStarCount) + strings.Repeat("⚡", i.HalfStarCount) + i.OriginScore,
			Arg:      fmt.Sprintf("%s%s", baseUrl, i.Url),
			Icon: struct {
				Path string `json:"path"`
			}{
				Path: path, //fmt.Sprintf("imgs/%s.png", searchType),
			},
		})
	}
	finalRes, _ := json.Marshal(struct {
		Items []AlfredItem `json:"items"`
	}{
		Items: r,
	})

	fmt.Println(string(finalRes))

	wg.Wait()
}

func main() {
	searchType := os.Args[1]
	query := strings.Join(os.Args[2:], " ")
	generateResponse(getItems(searchType, query), searchType)
}
