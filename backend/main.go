package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"time"
)

func requestGem(gem string) string {
	url := fmt.Sprintf("https://rubygems.org/api/v1/gems/%v", gem)
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(body)
}

type Node struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// sid: source ID of node
// tid: target ID of node
type Link struct {
	Sid int `json:"sid"`
	Tid int `json:"tid"`
}

func nodes(gemResponse string) ([]byte, []byte) {
	var nodes = []Node{}
	var links = []Link{}
	gem := gjson.Get(gemResponse, "name").String()
	deps := gjson.Get(gemResponse, "dependencies.runtime").Array()
	parentNode := Node{
		0,
		gem,
	}
	nodes = append(nodes, parentNode)
	for index, element := range deps {
		link := Link{
			0,
			index + 1,
		}
		links = append(links, link)
		node := Node{
			index + 1,
			element.Map()["name"].String(),
		}
		nodes = append(nodes, node)
	}
	nodesJSON, err := json.Marshal(nodes)
	if err != nil {
		fmt.Println(err)
	}

	linksJSON, err := json.Marshal(links)
	if err != nil {
		fmt.Println(err)
	}
	return nodesJSON, linksJSON
}

// def gem_dependency_tree(gem, edges)
// for dependency in gem_dependencies(gem) do
// 	next if dependency.empty?
// 	next if edges.has_key? dependency['name']

// 	if !edges.has_key? gem
// 		edges[gem] = []
// 	end
// 	edges[gem] << dependency['name']

// 	gem_dependency_tree(dependency['name'], edges)
// 	end
// 	return edges
// end

// func dependencyTree(nodes links[]byte, gem string) ([]byte []byte) {
// 	for _, dependency in requestGem(gem)
// 	return []byte, []byte
// }

func main() {
	r := gin.New()

	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "User-Agent", "Referrer", "Host", "Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  false,
		AllowOriginFunc:  func(origin string) bool { return true },
		MaxAge:           86400,
	}))

	r.GET("/gem/:gemname", func(c *gin.Context) {
		gem := c.Param("gemname")
		dependencies := requestGem(gem)
		nodes, links := nodes(dependencies)
		c.JSON(200, gin.H{
			"nodes": string(nodes),
			"links": string(links),
		})
	})
	r.Run()
}
