// govanityredirects creates static HTML that redirects go packages hosted
// on a vanity domain to GitHub.
package main // import "lazyhackergo.com/govanityredirect"

import (
	"flag"
	"fmt"
	"go/build"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const html = `<html>
<head>
<meta name="go-import" content="{{.Domain}}/{{.Repo}} git https://github.com/{{.Github}}/{{.Repo}}">
</head>
</html>
`

// templateData holds the data to be rendered by the template
type templateData struct {
	Domain string // vanity domain
	Repo   string // github repository name
	Github string // github user name
}

var (
	vanity = flag.String("domain", "", "vanity domain to forward to")
	user   = flag.String("github", "", "github user name")
	out    = flag.String("outdir", "", "output directory for the redirect html files")
	tmpl   = template.Must(template.New("index").Parse(html))
)

// Generate walks through the the vanity domain in your $GOPATH/src and creates
// the mapping files to your remote import path.
func Generate() error {

	// local path to the vanity domain repos
	root := filepath.Join(filepath.Join(build.Default.GOPATH, "src"), *vanity)

	abs, err := filepath.Abs(*out)

	if err != nil {
		return fmt.Errorf("unable to get absolute path of output directory. %v", err)
	}

	if strings.HasPrefix(abs, root) {
		return fmt.Errorf("Cannot set a output directory in GOPATH/src.")
	}

	// Traverse through the packages
	err = filepath.Walk(root,
		func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("received none nil err. %v", err)
			}

			// Skip hidden directories
			if f.IsDir() && f.Name()[0:1] == "." && f.Name() != root {
				return filepath.SkipDir
			}

			s, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("error getting path stat. %v", err)
			}

			// Skip files and if the directory is the vanity domain root
			if s.Mode().IsDir() && f.Name() != *vanity {
				var repo string
				dirs := strings.Split(path, string(os.PathSeparator))
				for i := 0; i < len(dirs); i++ {
					if dirs[i] == *vanity {
						repo = strings.Join(dirs[i+1:], string(os.PathSeparator))
						break
					}
				}
				data := templateData{
					Domain: *vanity,
					Repo:   repo,
					Github: *user,
				}
				writeIndexHTML(path, data)
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func writeIndexHTML(path string, data templateData) error {

	path = filepath.Join(*out, data.Domain, data.Repo)

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("Failed to mkdir %q: %s", path, err)
	}

	outfile := filepath.Join(path, "index.html")
	f, err := os.Create(outfile)
	if err != nil {
		return fmt.Errorf("failed to create %v: %v", outfile, err)
	}
	log.Printf("Writing %v\n", outfile)
	return tmpl.Execute(f, data)
}

func main() {

	flag.Parse()

	if *vanity == "" || *out == "" || *user == "" {
		flag.Usage()
		os.Exit(1)
	}

	err := Generate()
	if err != nil {
		log.Println(err)
	}
}
