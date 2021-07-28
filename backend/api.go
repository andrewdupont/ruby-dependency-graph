package main

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"time"
	"sync"
)

func requestGem(gem string, retry int) string {
	url := fmt.Sprintf("https://rubygems.org/api/v1/gems/%v", gem)
	// fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	// sleep and retry on rate limited
	if resp.StatusCode == 429 {
		time.Sleep(1 * time.Second)
		retry++
		fmt.Printf("retry to request gem: %s %v times\n", gem, retry)
		if retry > 5 {
			panic(resp.Status)
		}

		requestGem(gem, retry)
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
		fmt.Printf("added: %v\n", nodeName)
		newNode := Node{id, nodeName}
		m.nodes[nodeName] = newNode
		return true, newNode
	}
	return false, m.nodes[nodeName]
}

func (nodeSet *NodeSet) getValues() []Node {
	nodes := []Node{}
	for _, node := range nodeSet.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

type LinkSet struct {
	links map[Link]bool
}

func (m *LinkSet) add(link Link) {
	m.links[link] = true
}

func (linkSet *LinkSet) getValues() []Link {
	links := []Link{}
	for link := range linkSet.links {
		links = append(links, link)
	}
	return links
}

func dependencyLayer(wg *sync.WaitGroup, nodeSet NodeSet, linkSet LinkSet, curNode Node) {
	defer wg.Done()
	resp := requestGem(curNode.Name, 0)
	deps := gjson.Get(resp, "dependencies.runtime").Array()
	for _, dependency := range deps {

		dependencyName := gjson.Get(dependency.String(), "name").String()
		newID := len(nodeSet.nodes)
		added, newNode := nodeSet.add(newID, dependencyName)
		newLink := Link{curNode.Id, newNode.Id}
		linkSet.add(newLink)

		if !added {
			continue
		}
		wg.Add(1)
		// Recursively find the next layer of dependencies
		go dependencyLayer(wg, nodeSet, linkSet, newNode)
	}
}

// https://stackoverflow.com/questions/27103161/recursive-goroutines-what-is-the-neatest-way-to-tell-go-to-stop-reading-from-ch
func DependencyGraph(gem string) (string, string) {
	var nodeSet = NodeSet{nodes: make(map[string]Node)}
	_, node := nodeSet.add(0, gem)
	var linkSet = LinkSet{links: make(map[Link]bool)}
	var wg sync.WaitGroup
	wg.Add(1)
	dependencyLayer(&wg, nodeSet, linkSet, node)
	wg.Wait()

	nodesJSON, err := json.Marshal(nodeSet.getValues())
	if err != nil {
		fmt.Println(err)
	}
	linksJSON, err := json.Marshal(linkSet.getValues())
	if err != nil {
		fmt.Println(err)
	}
	return string(nodesJSON), string(linksJSON)
}
