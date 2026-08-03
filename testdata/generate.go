package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

var sources = [...]string{
	// "https://google.com",
	"https://wikipedia.org",
	// "https://youtube.com",
	"https://microsoft.com",
	// "https://apple.com",
	"https://amazon.com",
	// "https://yahoo.com",
	"https://dw.com",
	"https://bbc.com",
	"https://cnn.com",
	// "https://nytimes.com",
	"https://www.theguardian.com/europe",
}

func fileName(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "www.")
	s, _, _ = strings.Cut(s, ".")
	return fmt.Sprintf("./testdata/popular/%s.html", s)
}

func download(source, destination string) (err error) {
	r, err := http.Get(source)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, r.Body.Close())
	}()

	tree, err := html.Parse(bytes.NewReader(r.Body))
	if err != nil {
		return err
	}
	if tree == nil {
		return errors.New("parsed HTML is nil")
	}

	b := bytes.NewBuffer(nil)
	writeNodesWithLoremIpsum(b, tree, 0)
	if len(b.Bytes()) == 0 {
		return errors.New("parsed HTML empty")
	}

	f, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, b)
	return err
}

func main() {
	var err error
	for _, source := range sources {
		// p := fileName(source)
		// contents, err := os.ReadFile(p)
		// if err != nil || contents == nil {
		// 	panic(err)
		// }
		// for child := range tree.Descendants() {
		// 	fmt.Fprintln(os.Stdout, "found", child.Type, child.Data)
		// }
		// b := bytes.NewBuffer(nil)
		// writeNodesWithLoremIpsum(b, tree, 0)
		// if err = os.WriteFile(p, b.Bytes(), 0644); err != nil {
		// 	panic(err)
		// }
		// os.Exit(1)
		if err = download(source, fileName(source)); err != nil {
			slog.Default().Error(err.Error(), slog.String("URL", source))
		}
	}
}

func writeNodesWithLoremIpsum(w io.Writer, n *html.Node, depth int) {
	indent := strings.Repeat("\t", depth)
	switch n.Type {
	case html.DocumentNode:
		// Just parse children for the root document node
		for c := range n.ChildNodes() {
			writeNodesWithLoremIpsum(w, c, depth)
		}

	case html.ElementNode:
		// Build attributes string
		var attrs string
		for _, a := range n.Attr {
			attrs += fmt.Sprintf(" %s=\"%s\"", a.Key, a.Val)
		}

		// Handle self-closing tags or tags with no children
		if n.FirstChild == nil {
			fmt.Fprintf(w, "%s<%s%s />\n", indent, n.Data, attrs)
			return
		}

		// Open tag
		fmt.Fprintf(w, "%s<%s%s>\n", indent, n.Data, attrs)

		// Process children with increased depth
		for c := range n.ChildNodes() {
			writeNodesWithLoremIpsum(w, c, depth+1)
		}

		// Close tag
		fmt.Fprintf(w, "%s</%s>\n", indent, n.Data)

	case html.TextNode:
		if n.Parent != nil {
			if n.Parent.Type == html.ElementNode {
				if n.Parent.Data == "script" || n.Parent.Data == "style" {
					return // never print text nodes inside script or style tags
				}
			}
		}
		text := strings.TrimSpace(n.Data)
		length := len(text)
		if length == 0 {
			return
		}
		if length > len(loremIpsum) {
			length = len(loremIpsum)
		}
		fmt.Fprintf(w, "%s%s\n", indent, loremIpsum[:length])
	case html.CommentNode:
		// skip comments
		// fmt.Fprintf(w, "%s<!-- %s -->\n", indent, n.Data)
	}
}

const loremIpsum = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi at velit ut velit tincidunt iaculis. Praesent dignissim magna erat, sit amet sagittis est pellentesque quis. Mauris sapien lectus, eleifend congue odio luctus, sodales sollicitudin urna. Sed lobortis lorem nec ex iaculis convallis. Mauris non mauris nisl. Cras luctus et dui a vulputate. Cras quis urna libero. Nunc velit purus, posuere non nisi sit amet, elementum tempor ligula. Vivamus rhoncus tincidunt leo quis hendrerit. Fusce nec vulputate libero, eu efficitur nisi.

Praesent felis neque, feugiat non turpis et, venenatis ultricies eros. Etiam nunc tellus, molestie ut malesuada ac, rhoncus sit amet mi. Proin massa sem, tempor quis euismod a, euismod at ligula. Aenean lobortis, massa vel fermentum porta, magna felis semper magna, ac efficitur eros ligula molestie elit. Maecenas vel malesuada nisl, id lacinia nulla. Phasellus at dictum enim. Duis consequat neque sem, at congue velit porttitor at.

Sed tincidunt a mi vel lobortis. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Curabitur posuere sed purus nec porta. Integer iaculis bibendum massa, nec maximus urna feugiat sed. Vivamus massa urna, condimentum rhoncus massa sit amet, scelerisque volutpat orci. In in elit et ante commodo lacinia. Aenean nibh lectus, commodo nec dolor at, tristique scelerisque odio. Etiam laoreet enim ac maximus accumsan. Maecenas ultricies nunc quam, eu feugiat nisi commodo non. Fusce risus urna, varius id molestie ut, viverra ut dolor. Phasellus nibh quam, luctus sed porta ac, consectetur non lectus. Pellentesque ac viverra justo. Aenean sit amet egestas diam, eu suscipit mi.

Ut sagittis arcu leo, in aliquet dui ultricies in. Vestibulum elementum condimentum nunc, eu varius augue condimentum euismod. Morbi congue, est semper gravida porttitor, ligula neque suscipit lorem, vitae cursus lorem lorem eu lacus. Nunc eget viverra orci. Cras fringilla aliquet arcu, quis vulputate enim vehicula quis. Suspendisse eget semper lorem. Praesent sollicitudin porta pellentesque. Nunc dictum sit amet mi a suscipit. Fusce ac accumsan ipsum, id posuere ex. Phasellus convallis nunc id semper dapibus. Nulla maximus nisi eget lectus interdum, eu laoreet lorem pretium. Donec nec libero justo.

Vestibulum viverra pulvinar velit, quis vulputate velit congue vel. In molestie at purus eget vestibulum. Proin lobortis orci in commodo posuere. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Sed non sollicitudin lorem, nec malesuada mi. Nulla aliquam elit sem, eget semper mauris pulvinar accumsan. Integer pharetra placerat efficitur. Vivamus sodales ante vel est elementum, pulvinar dapibus velit tincidunt. Suspendisse id nisl ut lorem efficitur ultrices. Pellentesque vitae ultrices massa, et commodo mauris. Etiam et nibh sed ante sagittis sodales. Morbi faucibus accumsan bibendum. Maecenas sit amet elit ante. Donec porta a turpis vel varius. Morbi ut turpis id felis posuere finibus.`
