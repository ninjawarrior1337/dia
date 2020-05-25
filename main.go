package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Dia struct {
	BasePath     string
	Port         int
	tmplExecutor TemplateExecutor
	templateGlob string
	Debug        bool
}

type Context struct {
	w    http.ResponseWriter
	r    *http.Request
	path string
}

func main() {
	dia := NewDia()
	log.Fatal(dia.Start())
}

func NewDia() *Dia {
	debug := true
	glob := "templates/*.tmpl"
	var executor TemplateExecutor
	if debug {
		executor = DebugTemplateExecutor{Glob: glob}
	} else {
		executor = ReleaseTemplateExecutor{
			template.Must(template.ParseGlob(glob)),
		}
	}
	return &Dia{
		BasePath:     "filetester",
		Port:         1234,
		tmplExecutor: executor,
	}
}

func (d *Dia) FullPath(r *http.Request) string {
	return d.BasePath + r.URL.Path
}

func (d *Dia) Start() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{
			w:    w,
			r:    r,
			path: d.FullPath(r),
		}
		var f interface{}
		f, err := ioutil.ReadDir(ctx.path)
		if err != nil {
			f, _ = os.Open(ctx.path)
		}
		fmt.Println(f)
		switch f := f.(type) {
		case []os.FileInfo:
			hasIndex := d.handleIndex(ctx, f)
			if hasIndex {
				d.ProcessFile(ctx)
				return
			}
			d.tmplExecutor.ExecuteTemplate(ctx.w, "folder.tmpl", f)
		case *os.File:
			d.ProcessFile(ctx)
		}
	})
	return http.ListenAndServe(fmt.Sprintf(":%v", d.Port), nil)
}

func (d *Dia) ProcessFile(ctx *Context) {
	if filepath.Ext(ctx.path) == ".lambda" {
		err := runFileAsLambda(ctx)
		if err != nil {
			ctx.w.Write([]byte(err.Error()))
		}
		return
	}
	http.ServeFile(ctx.w, ctx.r, ctx.path)
}

func (d *Dia) handleIndex(ctx *Context, fileInfoArr []os.FileInfo) bool {
	for _, f := range fileInfoArr {
		if strings.Contains(f.Name(), "index") {
			indexPath := path.Join(ctx.path, f.Name())
			log.Println(indexPath)
			ctx.path = indexPath
			return true
		}
	}
	return false
}
