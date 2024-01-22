package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"gopkg.in/resty.v1"
)

type UrlItem struct {
	URL      string
	Category string
	Name     string
}

type SearchResultItem struct {
	Title         string
	OriginScore   string
	Url           string
	FullStarCount int
	HalfStarCount int
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
		URL:      "https://m.douban.com/search/?type=%s&query=%s",
		Category: "1001",
		Name:     "读书",
	},
	"movie": {
		URL:      "https://m.douban.com/search/?type=%s&query=%s",
		Category: "1002",
		Name:     "电影",
	},
	"music": {
		URL:      "https://m.douban.com/search/?type=%s&query=%s",
		Category: "1003",
		Name:     "音乐",
	},
	"game": {
		URL:      "https://m.douban.com/search/?type=%s&query=%s",
		Category: "1004",
		Name:     "游戏",
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
		resp, _ := resty.R().Get(fmt.Sprintf(v.URL, v.Category, searchString))
		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
		// 创建一个存储与搜索相关的li元素的切片
		var searchLis []*html.Node

		// 查找所有的li.search-module元素
		doc.Find("li.search-module").Each(func(i int, s *goquery.Selection) {
			// 检查这个li元素下是否有span包含文本指定的搜索类型
			if s.Find("span.search-results-modules-name").Text() == v.Name {
				// 如果找到，就将这个li元素下的ul.search_results_subjects > li元素添加到切片中
				s.Find("ul.search_results_subjects > li").Each(func(j int, li *goquery.Selection) {
					searchLis = append(searchLis, li.Nodes[0])
				})
			}
		})
		var node *goquery.Document
		var href, originScore, title string
		var fullStar, halfStar int
		r := make([]SearchResultItem, 0)
		for _, n := range searchLis {
			node = goquery.NewDocumentFromNode(n)
			href = getNodeAttr(node.Find("a").Nodes[0], "href")
			href = strings.ReplaceAll(href, "/"+searchType, "")

			originScore = node.Find("a > div > p > span").Text()
			title = node.Find("a > div > span").Text()
			fullStar = len(node.Find(".rating-star-small-full").Nodes)
			halfStar = len(node.Find(".rating-star-small-half").Nodes)
			r = append(r, SearchResultItem{
				Title:         title,
				OriginScore:   originScore,
				Url:           href,
				FullStarCount: fullStar,
				HalfStarCount: halfStar,
			})
		}
		return &r
	}
	return nil
}

func generateResponse(items *[]SearchResultItem, searchType string) {
	var baseUrl string
	if searchType == "game" {
		baseUrl = fmt.Sprintf("https://www.douban.com/game")
	} else {
		baseUrl = fmt.Sprintf("https://%s.douban.com", searchType)
	}

	var url string
	r := make([]AlfredItem, 0)
	for _, i := range *items {
		if searchType == "game" {
			url = i.Url[8:]
		} else {
			url = i.Url
		}

		r = append(r, AlfredItem{
			Type:     "file",
			Title:    i.Title,
			Subtitle: strings.Repeat("⭐", i.FullStarCount) + strings.Repeat("⚡", i.HalfStarCount) + i.OriginScore,
			Arg:      fmt.Sprintf("%s%s", baseUrl, url),
			Icon: struct {
				Path string `json:"path"`
			}{
				Path: fmt.Sprintf("imgs/%s.png", searchType),
			},
		})
	}
	finalRes, _ := json.Marshal(struct {
		Items []AlfredItem `json:"items"`
	}{
		Items: r,
	})
	fmt.Println(string(finalRes))
}

func main() {
	searchType := os.Args[1]
	query := strings.Join(os.Args[2:], " ")
	generateResponse(getItems(searchType, query), searchType)
}
