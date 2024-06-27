// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package golang

import (
	"bytes"
	"errors"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"

	"storj.io/dbx/tmplutil"
)

func TestLoadTemplate(t *testing.T) {
	loader := testLoader{
		"golang.foo.tmpl":            "foo",
		"golang.foo.d1.tmpl":         "d1 foo",
		"golang.bar.tmpl":            "bar",
		"golang.onlydialect.d1.tmpl": "only",
		"golang.header.tmpl":         "header",
		"golang.misc.tmpl":           "misc",
		"golang.footer.tmpl":         "footer",
	}

	renderer, err := New(loader, &Options{})
	require.NoError(t, err)

	assertRender := func(expected string, tmpl *template.Template) {
		out := bytes.NewBuffer([]byte{})
		err = tmpl.ExecuteTemplate(out, "test", nil)
		require.NoError(t, err)
		require.Equal(t, expected, out.String())
	}

	_, err = renderer.LoadTemplate("qwe", "d1")
	require.Error(t, err)

	tmpl, err := renderer.LoadTemplate("foo", "d1")
	require.NoError(t, err)
	assertRender("d1 foo", tmpl)

	tmpl, err = renderer.LoadTemplate("foo", "d2")
	require.NoError(t, err)
	assertRender("foo", tmpl)

	tmpl, err = renderer.LoadTemplate("onlydialect", "d1")
	require.NoError(t, err)
	assertRender("only", tmpl)

	_, err = renderer.LoadTemplate("onlydialect", "d2")
	require.Error(t, err)
}

type testLoader map[string]string

func (t testLoader) Load(name string, funcs template.FuncMap) (*template.Template, error) {
	templateDefinition, found := t[name]
	if !found {
		return nil, errors.New("does not exist " + name)
	}
	return template.New("test").Parse(templateDefinition)
}

var _ tmplutil.Loader = testLoader{}
