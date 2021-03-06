// govanityredirects creates static HTMLs files that redirects go packages
// vanity import paths to Github.
//
// govanityredirect looks under $GOPATH/src for the custom domain directory
// and traverse through the packages and create a go-import index.html file to
// the output directory.
package main // import "lazyhacker.dev/govanityredirect"

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
{{ range .Domain }}
<meta name="go-import" content="{{ . }}/{{$.Repo}} git https://github.com/{{$.Github}}/{{$.Repo}}">
{{ end }}
<meta http-equiv="refresh" content="2; url=https://godoc.org/{{index .Domain 0}}/{{.Repo}}">
</head>
<body>
Redirecting to <a href="https://godoc.org/{{index .Domain 0}}/{{.Repo}}">https://godoc.org/{{index .Domain 0}}/{{.Repo}}</a>.
</html>
`

// templateData holds the data to be rendered by the template
type templateData struct {
	Domain []string // vanity domains
	Repo   string   // github repository name
	Github string   // github user name
}

var (
	rootdir = flag.String("repo", "", "The root dir of all the depos in $GOPATH/src/...")
	vanity  = flag.String("vanity", "", "vanity domain(s) to forward to (comma separated)")
	user    = flag.String("github", "", "github user name")
	out     = flag.String("outdir", "", "output directory for the redirect html files")
	alt     = flag.String("alt", "vanity", "if index.html exists write to index.html.<alt> instead (default: vanity)")
	tmpl    = template.Must(template.New("index").Parse(html))
)

// Generate walks through the the vanity domain in your $GOPATH/src and creates
// the mapping files to your remote import path.
func Generate() error {

	// local path to the vanity domain repos
	root := filepath.Join(filepath.Join(build.Default.GOPATH, "src"), *rootdir)

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
			if s.Mode().IsDir() && f.Name() != *rootdir {
				var repo string
				dirs := strings.Split(path, string(os.PathSeparator))
				for i := 0; i < len(dirs); i++ {
					if dirs[i] == *rootdir {
						repo = strings.Join(dirs[i+1:], string(os.PathSeparator))
						break
					}
				}

				vd := strings.Split(*vanity, ",")
				data := templateData{
					Domain: vd,
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

	path = filepath.Join(*out, data.Repo)

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("Failed to mkdir %q: %s", path, err)
	}

	outfile := filepath.Join(path, "index.html")

	if _, err := os.Stat(outfile); err == nil {
		fmt.Printf("%v exists!.", outfile)
		outfile = fmt.Sprintf("%v.%v", outfile, *alt)
		fmt.Printf("  Writing to %v instead.\n", outfile)
	}
	f, err := os.Create(outfile)
	if err != nil {
		return fmt.Errorf("failed to create %v: %v", outfile, err)
	}
	fmt.Printf("Writing %v\n", outfile)
	return tmpl.Execute(f, data)

}

func main() {

	flag.Parse()

	if *vanity == "" || *out == "" || *user == "" || *rootdir == "" {
		flag.Usage()
		os.Exit(1)
	}

	err := Generate()
	if err != nil {
		log.Println(err)
	}
}
