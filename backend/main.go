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
	// fmt.Println(url)
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

func nodeExists(name string, nodes []Node) (result bool, node Node) {
	result = false
	foundNode := Node{}
	for _, node := range nodes {
		if node.Name == name {
			result = true
			foundNode = node
			break
		}
	}
	return result, foundNode
}

func dependencyTree(nodes []Node, links []Link, curNode Node) ([]Node, []Link) {
	resp := requestGem(curNode.Name)
	deps := gjson.Get(resp, "dependencies.runtime").Array()
	for _, dependency := range deps {
		gemName := gjson.Get(dependency.String(), "name").String()
		// Guard, this node has already been found.
		result, foundNode := nodeExists(gemName, nodes)
		if result {
			fmt.Println("already found", foundNode)
			newLink := Link{curNode.Id, foundNode.Id}
			links = append(links, newLink)
			continue
		}
		newID := len(nodes)
		newNode := Node{newID, gemName}
		newLink := Link{curNode.Id, newNode.Id,}
		nodes = append(nodes, newNode)
		links = append(links, newLink)
		nodes, links = dependencyTree(nodes, links, newNode)
	}
	return nodes, links
}

func fullDependencyTree(gem string) (string, string) {
	var nodes = []Node{}
	var links = []Link{}
	baseNode := Node{0, gem}
	nodes = append(nodes, baseNode)
	nodeArray, linkArray := dependencyTree(nodes, links, baseNode)

	nodesJSON, err := json.Marshal(nodeArray)
	if err != nil {
		fmt.Println(err)
	}

	linksJSON, err := json.Marshal(linkArray)
	if err != nil {
		fmt.Println(err)
	}
	return string(nodesJSON), string(linksJSON)
}

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
		nodes, links := fullDependencyTree(gem)
		c.JSON(200, gin.H{
			"nodes": nodes,
			"links": links,
		})
	})
	r.Run()
}
