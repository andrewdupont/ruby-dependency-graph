package main

import (
	"fmt"
	"net/http"
	"io"
	"github.com/tidwall/gjson"
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

// func nodes(gemResponse string) ([]byte, []byte) {
// 	var nodes = []Node{}
// 	var links = []Link{}
// 	gem := gjson.Get(gemResponse, "name").String()
// 	deps := gjson.Get(gemResponse, "dependencies.runtime").Array()
// 	parentNode := Node{
// 		0,
// 		gem,
// 	}
// 	nodes = append(nodes, parentNode)
// 	for index, element := range deps {
// 		link := Link{
// 			0,
// 			index + 1,
// 		}
// 		links = append(links, link)
// 		node := Node{
// 			index + 1,
// 			element.Map()["name"].String(),
// 		}
// 		nodes = append(nodes, node)
// 	}
// 	nodesJSON, err := json.Marshal(nodes)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	linksJSON, err := json.Marshal(links)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return nodesJSON, linksJSON
// }

// def gem_dependency_tree(gem, edges)
//   for dependency in gem_dependencies(gem) do
// 	   next if dependency.empty?
// 	   next if edges.has_key? dependency['name']

// 	   if !edges.has_key? gem
// 	  	edges[gem] = []
// 	   end
// 	   edges[gem] << dependency['name']

// 	   gem_dependency_tree(dependency['name'], edges)
// 	 end
// 	 return edges
// end


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

func main() {
	var nodes = []Node{}
	var links = []Link{}
	gem := "devise"
	baseNode := Node{0, gem}
	nodes = append(nodes, baseNode)
	nodeArray, linkArray := dependencyTree(nodes, links, baseNode)

	fmt.Println(nodeArray, linkArray)
}
