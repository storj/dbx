// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package tmplutil

import (
	"bytes"
	"io"
	"text/template"
)

func RenderString(tmpl *template.Template, name string,
	data interface{}) (string, error) {

	var buf bytes.Buffer
	if err := Render(tmpl, &buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func Render(tmpl *template.Template, w io.Writer, name string,
	data interface{}) error {

	if name == "" {
		return Error.Wrap(tmpl.Execute(w, data))
	}
	return Error.Wrap(tmpl.ExecuteTemplate(w, name, data))
}
