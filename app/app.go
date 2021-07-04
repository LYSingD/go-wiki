package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"runtime"

	"david-lys.dev/gowiki"
)

// Global variables
var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b) // /Users/lysing/tutorial/golang/write-web-applications/gowiki
	editPath   = filepath.Join("tmpl", "edit.html")
	viewPath   = filepath.Join("tmpl", "view.html")
	// To avoid *renderTemplate* calls ParseFiles every time a page is render

	// *Must* is a convenience wrapper that panics when passed a non-nil error value, and returns the *Template unaltered
	// A panic is appropriate here; if the templates can't be loaded the only sensible thing to do is exit the program

	// *ParseFiles* takes any number of string arguments that identify our template files, and parses those files into templates
	// that are named after the base file name
	templates = template.Must(template.ParseFiles(editPath, viewPath))
	// Prevent the user can supply an arbitrary path to be read/written on the server.
	// Add "regexp" to the import list
	// MustCompile will parse and compile the regular expression, and return a regexp.
	// is distinct from Compile in that it will panic if the expression compilation fails
	validPath = regexp.MustCompile(("^/$|^/(edit|save|view)/([a-zA-Z0-9-]+)$"))
)

// function handler is of the type `http.HandlerFunc`.
// It takes an http.ResponseWriter and an http.Request as its argument
// Params: http.ResponseWriter -> assembles the HTTP server's response; by writing to it, we send data to the HTTP client.
// Params: http.Request -> a data structure that represents the client HTTP request.
func handler(w http.ResponseWriter, r *http.Request) {
	// r.URL.Path is the path component of the request URL
	// the trailing [1:] means "create a sub-slice of Path from the 1st character to the end" -> drops the leading "/" from the path
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

// getTitle would validate the path and extract the page title
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid New Title")
	}
	return m[2], nil // The title is the second subexpression.
}

func renderTemplate(w http.ResponseWriter, p *gowiki.Page, action string) {
	// ParseFile will read the content of edit.html and return a *template.Template
	htmlFile := "tmpl/" + action + ".html"
	html, err := template.ParseFiles(htmlFile)
	if err != nil {
		// http.Error sends a specified HTTP response code (in this case "Internal Server Error") and error message
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Execute executes the template, writing the generated HTML to the http.ResponseWriter.
	err = html.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// newRenderTemplate could avoid the parseFile be called every time a page is rendered.
func newRenderTemplate(w http.ResponseWriter, p *gowiki.Page, action string) {
	htmlFile := action + ".html"
	err := templates.ExecuteTemplate(w, htmlFile, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Here we will extract the page title from the Request,
		// and call the provided handler 'fn'
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request, title string) {
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

// Allow users to view a wiki page. It will handle URLs prefixed with "/view/".
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	// Loads the page data, formats the page with a string of simple HTML, and writes it to w, the http.ResponseWriter
	p, err := gowiki.LoadPage(title)
	// If the requested Page doesn't exit, it should redirect the client to the edit Page so the content may be created
	if err != nil {
		updatedURL := "/edit/" + title
		// Redirect - adds an HTTP status code of http.StatusFound(302) and a Location header to the HTTP response
		http.Redirect(w, r, updatedURL, http.StatusFound)
		return
	}
	newRenderTemplate(w, p, "view")
	// fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body) // we could use html/template to make the html code better
}

// editHandler loads the page (or, if it doesn't exit, create an empty *Page* struct), and displays on HTML form.
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := gowiki.LoadPage(title)
	if err != nil {
		p = &gowiki.Page{Title: title}
	}
	newRenderTemplate(w, p, "edit")
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	// FormValue returned type of string
	body := r.FormValue("body")
	p := &gowiki.Page{Title: title, Body: []byte(body)}
	err := p.SavePage()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	redirectURL := "/view/" + title
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func oldMain() {
	// Begins with a call to http.HandleFun, which tells the http package to handle all request to the web root ("/") with handler.
	http.HandleFunc("/", handler)
	// Calls http.ListenAndServe, specifying that it should listen on port 8080 on any interface (":8080")
	// This function will block until the program is terminated.
	// ListenAndServe always returns an error, since it only returns when an expected error occurs. In order to log that error
	// we wrap the function call with `log.Fatal`
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	http.HandleFunc("/", makeHandler(rootHandler))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
