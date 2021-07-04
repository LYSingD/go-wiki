// https://golang.org/doc/articles/wiki/
package gowiki

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
)

// basepath represents the path of the current folder
var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b) // /Users/lysing/tutorial/golang/write-web-applications/gowiki
)

// A wiki consists of a series of interconnected pages, each of which has a title and a body (the page content)
// Define Page as struct with two fields representing the title and body

type Page struct {
	Title string
	Body  []byte // a byte slice
	// Body element is a []byte rather than `string` because that is the type expected by the `io` libraries we will use
}

// Save the Page's body to a text file.
// Save takes as its receiver p, a pointer to a Page.
// The save method returns an error value because that is the return type of WriteFile (a standard library function that
// writes a byte slice to a file). The save method returns the error value, to let the application handle it should anything
// go wrong while writing the file. If all goes well, Page.save() will return nil (the zero-value for pointers, interface, and some other type)
func (p *Page) save() error {
	filename := p.Title + ".txt"
	path := filepath.Join(basepath, "doc", filename)
	return ioutil.WriteFile(path, p.Body, 0600) // chmod 0600, read-write permissions
}

func (p *Page) SavePage() error {
	filename := p.Title + ".txt"
	path := filepath.Join(basepath, "doc", filename)
	return ioutil.WriteFile(path, p.Body, 0600) // chmod 0600, read-write permissions
}

// The function loadPage constructs the file name from the title parameter,
// reads the file's contents into a new variable body,
// and returns a pointer to a Page literal constructed with the proper title and body values.
func LoadPage(title string) (*Page, error) {
	filename := title + ".txt"
	path := filepath.Join(basepath, "doc", filename) // /Users/lysing/tutorial/golang/write-web-applications/gowiki/doc/test.txt
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}
