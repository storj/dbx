// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package tmplutil

import (
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"storj.io/dbx/internal/inflect"
)

type Loader interface {
	Load(name string, funcs template.FuncMap) (*template.Template, error)
}

type LoaderFunc func(name string, funcs template.FuncMap) (
	*template.Template, error)

func (fn LoaderFunc) Load(name string, funcs template.FuncMap) (
	*template.Template, error) {

	return fn(name, funcs)
}

type dirLoader struct {
	dir      string
	fallback Loader
}

func DirLoader(dir string, fallback Loader) Loader {
	return dirLoader{
		dir:      dir,
		fallback: fallback,
	}
}

func (d dirLoader) Load(name string, funcs template.FuncMap) (
	*template.Template, error) {

	data, err := os.ReadFile(filepath.Join(d.dir, name))
	if err != nil {
		if os.IsNotExist(err) {
			return d.fallback.Load(name, funcs)
		}
		return nil, Error.Wrap(err)
	}
	return loadTemplate(name, data, funcs)
}

func FSLoader(fs fs.FS) Loader {
	return fsLoader{FS: fs}
}

type fsLoader struct{ fs.FS }

func (b fsLoader) Load(name string, funcs template.FuncMap) (
	*template.Template, error) {

	data, err := fs.ReadFile(b.FS, name)
	if err != nil {
		return nil, Error.Wrap(err)
	}
	return loadTemplate(name, data, funcs)
}

func loadTemplate(name string, data []byte, funcs template.FuncMap) (
	*template.Template, error) {

	if funcs == nil {
		funcs = make(template.FuncMap)
	}

	safeset := func(name string, fn interface{}) {
		if funcs[name] == nil {
			funcs[name] = fn
		}
	}

	safeset("pluralize", inflect.Pluralize)
	safeset("singularize", inflect.Singularize)
	safeset("camelize", inflect.Camelize)
	safeset("cameldown", inflect.CamelizeDownFirst)
	safeset("underscore", inflect.Underscore)

	tmpl, err := template.New(name).Funcs(funcs).Parse(string(data))
	return tmpl, Error.Wrap(err)
}
