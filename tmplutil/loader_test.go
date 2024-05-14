// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package tmplutil

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"

	"storj.io/dbx/templates"
)

func TestErrorCodes(t *testing.T) {
	loader := FSLoader(templates.Asset)
	_, err := loader.Load("qwe", template.FuncMap{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not exist")
}
