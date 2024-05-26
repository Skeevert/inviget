package main

import (
    "fmt"
    "golang.org/x/net/html"
    "net/http"
    "os"
    "strings"
    "time"
)

func errExit(err error) {
    fmt.Errorf("Error: %w\n", err)
    os.Exit(1)
}

func findNode(node *html.Node, target string) *html.Node {
    for child := node.FirstChild; child != nil; child = child.NextSibling {
        if child.Type == html.ElementNode && child.Data == target {
            return child
        }
    }

    return nil
}

// Called upon li node
func parseLink(node *html.Node) (bool, string) {
    pNode := findNode(node, "p")
    if pNode == nil {
        return false, ""
    }

    aNode := findNode(pNode, "a")
    if aNode == nil {
        return false, ""
    }

    for _, attr := range aNode.Attr {
        if attr.Key == "href" {
            return true, attr.Val
        }
    }

    return false, ""
}

func searchLinks(doc *html.Node) []string {
    var links []string
    var link func(*html.Node)
    link = func(n *html.Node) {
        if n.Type == html.ElementNode && n.Data == "li" {
            found, link := parseLink(n)
            if found {
                links = append(links, link)
            }
        }

        // traverses the HTML of the webpage from the first child node
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            link(c)
        }
    }
    link(doc)

    return links
}

func probeLink(ch chan string, link string) {
    client := &http.Client {
        Timeout: time.Second * 5,
    }

    resp, err := client.Get(link)
    if err != nil {
        return
    }

    if resp.StatusCode == 200 {
        ch <- link
    }
}

func main() {
    client := &http.Client {
        Timeout: time.Second * 5,
    }

    resp, err := client.Get("https://docs.invidious.io/instances/#list-of-public-invidious-instances-sorted-from-oldest-to-newest")
    if err != nil {
        errExit(err)
    }

    doc, err := html.Parse(resp.Body)
    if err != nil {
        errExit(err)
    }

    links := searchLinks(doc)

    linkChan := make(chan string)
    for _, link := range links {
        // Do not probe i2p and onion links
        if !strings.HasSuffix(link, ".i2p") && !strings.HasSuffix(link, ".onion") {
            go probeLink(linkChan, link)
        }
    }

    available := <-linkChan
    fmt.Println(available)
    close(linkChan)
}
