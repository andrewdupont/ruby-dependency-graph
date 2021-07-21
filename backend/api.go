package main

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"sync"
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

type NodeSet struct {
	sync.Mutex
	nodes map[string]Node
}

// safely add new node to set of nodes and return whether it was added and the
// Node from the provided name.
func (m *NodeSet) add(id int, nodeName string) (bool, Node) {
	m.Lock()
	defer m.Unlock()
	_, ok := m.nodes[nodeName]
	if !ok {
			newNode := Node{id, nodeName}
			m.nodes[nodeName] = newNode
			return true, newNode
	}
	return false, m.nodes[nodeName]
}

func (nodesSet *NodeSet) getValues() []Node {
	nodes := []Node{}
	for _, node := range nodesSet.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func dependencyLayer(nodesSet NodeSet, links []Link, curNode Node) (NodeSet, []Link) {
	resp := requestGem(curNode.Name)
	deps := gjson.Get(resp, "dependencies.runtime").Array()
	for _, dependency := range deps {
		gemName := gjson.Get(dependency.String(), "name").String()
		// Guard, this node has already been found.
		// still add the link to the slice
		newID := len(nodesSet.nodes)
		added, node := nodesSet.add(newID, gemName)
		if !added {
			newLink := Link{curNode.Id, node.Id}
			links = append(links, newLink)
		}

		// Recursively find the next layer of dependencies
		go dependencyLayer(nodesSet, links, node)
	}
	return nodesSet, links
}

// https://stackoverflow.com/questions/27103161/recursive-goroutines-what-is-the-neatest-way-to-tell-go-to-stop-reading-from-ch
func DependencyGraph(gem string) (string, string) {
	var nodesSet = NodeSet{nodes: make(map[string]Node)}
	var links = []Link{}
	_, node := nodesSet.add(0, gem)
	nodesResults, linkResults := dependencyLayer(nodesSet, links, node)

	nodesJSON, err := json.Marshal(nodesResults.getValues())
	if err != nil {
		fmt.Println(err)
	}

	linksJSON, err := json.Marshal(linkResults)
	if err != nil {
		fmt.Println(err)
	}
	return string(nodesJSON), string(linksJSON)
}
