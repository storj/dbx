// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"sort"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"storj.io/dbx/code"
	"storj.io/dbx/ir"
	"storj.io/dbx/sql"
	"storj.io/dbx/sqlgen"
	"storj.io/dbx/sqlgen/sqlbundle"
	"storj.io/dbx/sqlgen/sqlembedgo"
	"storj.io/dbx/tmplutil"
)

type publicMethod struct {
	Signature string
	Invoke    string
}

type Options struct {
	Package         string
	SupportUserdata bool
}

type Renderer struct {
	loader          tmplutil.Loader
	header          *template.Template
	footer          *template.Template
	misc            *template.Template
	decl            *template.Template
	cre             *template.Template
	cre_raw         *template.Template
	get_all         *template.Template
	get_has         *template.Template
	get_count       *template.Template
	get_limitoffset *template.Template
	get_paged       *template.Template
	get_scalar      *template.Template
	get_scalar_all  *template.Template
	get_one         *template.Template
	get_one_all     *template.Template
	get_first       *template.Template
	upd             *template.Template
	del             *template.Template
	del_all         *template.Template
	del_world       *template.Template
	get_last        *template.Template
	methods         map[string]publicMethod
	options         Options
}

var _ code.Renderer = (*Renderer)(nil)

func New(loader tmplutil.Loader, options *Options) (
	r *Renderer, err error) {

	r = &Renderer{
		loader:  loader,
		options: *options,
		methods: map[string]publicMethod{},
	}

	r.header, err = loader.Load("golang.header.tmpl", nil)
	if err != nil {
		return nil, err
	}

	r.footer, err = loader.Load("golang.footer.tmpl", nil)
	if err != nil {
		return nil, err
	}

	r.misc, err = loader.Load("golang.misc.tmpl", nil)
	if err != nil {
		return nil, err
	}

	r.decl, err = loader.Load("golang.decl.tmpl", nil)
	if err != nil {
		return nil, err
	}

	funcs := template.FuncMap{
		"sliceof":           sliceofFn,
		"param":             paramFn,
		"arg":               argFn,
		"value":             valueFn,
		"zero":              zeroFn,
		"init":              initFn,
		"initnew":           initnewFn,
		"declare":           declareFn,
		"addrof":            addrofFn,
		"flatten":           flattenFn,
		"comma":             commaFn,
		"ctxparam":          ctxparamFn,
		"ctxarg":            ctxargFn,
		"embedsql":          embedsqlFn,
		"embedplaceholders": embedplaceholdersFn,
		"embedvalues":       embedvaluesFn,
		"rename":            renameFn,
		"double":            doubleFn,
		"slice":             sliceFn,
	}

	r.cre, err = loader.Load("golang.create.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.cre_raw, err = loader.Load("golang.create-raw.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_all, err = loader.Load("golang.get-all.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_has, err = loader.Load("golang.get-has.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_count, err = loader.Load("golang.get-count.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_paged, err = loader.Load("golang.get-paged.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_limitoffset, err = loader.Load("golang.get-limitoffset.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_scalar, err = loader.Load("golang.get-scalar.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_scalar_all, err = loader.Load("golang.get-scalar-all.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_one, err = loader.Load("golang.get-one.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_one_all, err = loader.Load("golang.get-one-all.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_first, err = loader.Load("golang.get-first.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.upd, err = loader.Load("golang.update.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.del, err = loader.Load("golang.delete.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.del_all, err = loader.Load("golang.delete-all.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.del_world, err = loader.Load("golang.delete-world.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_last, err = loader.Load("golang.get-last.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Renderer) RenderCode(root *ir.Root, dialects []sql.Dialect) (
	rendered []byte, err error) {
	var buf bytes.Buffer

	if err := r.renderHeader(&buf, root, dialects); err != nil {
		return nil, err
	}

	// Render any result structs for multi-field reads
	extra_structs := map[string]*Struct{}
	extra_struct_names := []string{}
	for _, read := range root.Reads {
		if read.View == ir.Count || read.View == ir.Has {
			continue
		}
		if model := read.SelectedModel(); model != nil {
			continue
		}
		s := ResultStructFromRead(read)
		if extra_structs[s.Name] != nil {
			continue
		}
		extra_structs[s.Name] = s
		extra_struct_names = append(extra_struct_names, s.Name)
	}
	for _, read := range root.Reads {
		if read.View != ir.Paged {
			continue
		}
		s := ContinuationStructFromRead(read)
		if extra_structs[s.Name] != nil {
			continue
		}
		extra_structs[s.Name] = s
		extra_struct_names = append(extra_struct_names, s.Name)
	}
	sort.Strings(extra_struct_names)
	for _, name := range extra_struct_names {
		if err := r.renderStruct(&buf, extra_structs[name]); err != nil {
			return nil, err
		}
	}

	for _, dialect := range dialects {
		var gets []*ir.Model
		for _, cre := range root.Creates {
			gets = append(gets, cre.Model)
			if err := r.renderCreate(&buf, cre, dialect); err != nil {
				return nil, err
			}
		}
		for _, read := range root.Reads {
			if err := r.renderRead(&buf, read, dialect); err != nil {
				return nil, err
			}
		}
		for _, upd := range root.Updates {
			gets = append(gets, upd.Model)
			if err := r.renderUpdate(&buf, upd, dialect); err != nil {
				return nil, err
			}
		}
		for _, del := range root.Deletes {
			if err := r.renderDelete(&buf, del, dialect); err != nil {
				return nil, err
			}
		}

		if len(gets) > 0 && !dialect.Features().Returning {
			// dialect does not support returning columns on insert and updates
			// so we need to generate a function to support getting by last
			// insert id.
			done := map[*ir.Model]bool{}
			for _, model := range gets {
				if done[model] {
					continue
				}
				done[model] = true
				if err := r.renderGetLast(&buf, model, dialect); err != nil {
					return nil, err
				}
			}
		}

		if err = r.renderDialectFuncs(&buf, dialect); err != nil {
			return nil, err
		}

		if err := r.renderDeleteWorld(&buf, root.Models, dialect); err != nil {
			return nil, err
		}
	}

	if err := r.renderFooter(&buf); err != nil {
		return nil, err
	}

	if err := r.renderDialectOpens(&buf, dialects); err != nil {
		return nil, err
	}

	rendered, err = format.Source(buf.Bytes())
	if err != nil {
		return nil, Error.Wrap(err)
	}

	return rendered, nil
}

func (r *Renderer) renderHeader(w io.Writer, root *ir.Root,
	dialects []sql.Dialect) error {

	type headerDialect struct {
		Name          string
		SchemaSQL     []string
		DropSchemaSQL []string
	}

	type headerParams struct {
		Package      string
		ExtraImports []string
		Dialects     []headerDialect
		Structs      []*ModelStruct
		Options      Options
		SQLSupport   string
	}

	params := headerParams{
		Package:    r.options.Package,
		Structs:    ModelStructsFromIR(root.Models),
		Options:    r.options,
		SQLSupport: sqlbundle.Source,
	}

	extra_imports := make(map[string]struct{})
	for _, dialect := range dialects {
		dialect_create_schema := sqlgen.RenderAll(dialect,
			sql.SchemaCreateSQL(root, dialect), sqlgen.NoTerminate, sqlgen.NoFlatten)

		dialect_drop_schema := sqlgen.RenderAll(dialect,
			sql.SchemaDropSQL(root, dialect), sqlgen.NoTerminate, sqlgen.NoFlatten)

		dialect_tmpl, err := r.loadDialect(dialect)
		if err != nil {
			return err
		}

		dialect_import, err := tmplutil.RenderString(dialect_tmpl, "import",
			nil)
		if err != nil {
			return err
		}
		for _, extra_import := range strings.Split(dialect_import, "\n") {
			extra_imports[strings.TrimSpace(extra_import)] = struct{}{}
		}

		params.Dialects = append(params.Dialects, headerDialect{
			Name:          dialect.Name(),
			SchemaSQL:     dialect_create_schema,
			DropSchemaSQL: dialect_drop_schema,
		})
	}

	for extra_import := range extra_imports {
		params.ExtraImports = append(params.ExtraImports, extra_import)
	}
	sort.Strings(params.ExtraImports)

	return tmplutil.Render(r.header, w, "", params)
}

func (r *Renderer) renderCreate(w io.Writer, ir_cre *ir.Create,
	dialect sql.Dialect) (err error) {

	if ir_cre.Raw {
		cre := RawCreateFromIR(ir_cre, dialect)
		return r.renderFunc(r.cre_raw, w, cre, dialect)
	} else {
		cre := CreateFromIR(ir_cre, dialect)
		return r.renderFunc(r.cre, w, cre, dialect)
	}
}

func (r *Renderer) renderRead(w io.Writer, ir_read *ir.Read,
	dialect sql.Dialect) error {

	get := GetFromIR(ir_read, dialect)

	var tmpl *template.Template
	switch ir_read.View {
	case ir.All:
		tmpl = r.get_all
	case ir.LimitOffset:
		tmpl = r.get_limitoffset
	case ir.Paged:
		tmpl = r.get_paged
	case ir.Count:
		tmpl = r.get_count
	case ir.Has:
		tmpl = r.get_has
	case ir.Scalar:
		if ir_read.Distinct() {
			tmpl = r.get_scalar
		} else {
			tmpl = r.get_scalar_all
		}
	case ir.One:
		if ir_read.Distinct() {
			tmpl = r.get_one
		} else {
			tmpl = r.get_one_all
		}
	case ir.First:
		tmpl = r.get_first
	default:
		panic(fmt.Sprintf("unhandled read view %s", ir_read.View))
	}

	return r.renderFunc(tmpl, w, get, dialect)
}

func (r *Renderer) renderUpdate(w io.Writer, ir_upd *ir.Update,
	dialect sql.Dialect) error {

	upd := UpdateFromIR(ir_upd, dialect)
	return r.renderFunc(r.upd, w, upd, dialect)
}

func (r *Renderer) renderDelete(w io.Writer, ir_del *ir.Delete,
	dialect sql.Dialect) error {

	del := DeleteFromIR(ir_del, dialect)
	if ir_del.Distinct() {
		return r.renderFunc(r.del, w, del, dialect)
	} else {
		return r.renderFunc(r.del_all, w, del, dialect)
	}
}

func (r *Renderer) renderDeleteWorld(w io.Writer, ir_models []*ir.Model,
	dialect sql.Dialect) error {

	type deleteWorld struct {
		Dialect string
		SQLs    []string
	}

	del := deleteWorld{
		Dialect: dialect.Name(),
	}
	for i := len(ir_models) - 1; i >= 0; i-- {
		sql := sqlgen.Render(dialect, sql.DeleteSQL(&ir.Delete{
			Model: ir_models[i],
		}, dialect))
		del.SQLs = append(del.SQLs, sql)
	}

	return r.renderFunc(r.del_world, w, del, dialect)
}

func (r *Renderer) renderFunc(tmpl *template.Template, w io.Writer,
	data interface{}, dialect sql.Dialect) (err error) {

	var signature bytes.Buffer
	err = tmplutil.Render(tmpl, &signature, "signature", data)
	if err != nil {
		return err
	}

	method := publicMethod{
		Signature: signature.String(),
	}

	if isExported(method.Signature) {
		var invoke bytes.Buffer
		err = tmplutil.Render(tmpl, &invoke, "invoke", data)
		if err != nil {
			return err
		}
		method.Invoke = invoke.String()
		r.methods[method.Signature] = method
	}

	var body bytes.Buffer
	err = tmplutil.Render(tmpl, &body, "body", data)
	if err != nil {
		return err
	}

	type funcDecl struct {
		ReceiverBase string
		Signature    string
		Body         string
	}

	decl := funcDecl{
		ReceiverBase: dialect.Name(),
		Signature:    method.Signature,
		Body:         body.String(),
	}

	err = tmplutil.Render(r.decl, w, "decl", decl)
	if err != nil {
		return err
	}

	return nil
}

func (r *Renderer) renderStruct(w io.Writer, s *Struct) (err error) {
	return tmplutil.Render(r.misc, w, "struct", s)
}

func isExported(signature string) bool {
	r, _ := utf8.DecodeRuneInString(signature)
	return unicode.IsUpper(r)
}

func (r *Renderer) renderGetLast(w io.Writer, model *ir.Model,
	dialect sql.Dialect) error {

	type getLast struct {
		Info   sqlembedgo.Info
		Return *Var
	}

	get_last_sql := sql.GetLastSQL(model, dialect)
	get_last := getLast{
		Info:   sqlembedgo.Embed("__", get_last_sql),
		Return: VarFromModel(model),
	}

	return r.renderFunc(r.get_last, w, get_last, dialect)
}

func (r *Renderer) renderFooter(w io.Writer) error {
	var keys []string
	for key := range r.methods {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	type footerData struct {
		Methods []publicMethod
	}

	data := footerData{}

	for _, key := range keys {
		data.Methods = append(data.Methods, r.methods[key])
	}

	return tmplutil.Render(r.footer, w, "", data)
}

func (r *Renderer) renderDialectFuncs(w io.Writer, dialect sql.Dialect) (
	err error) {

	type dialectFunc struct {
		Receiver string
	}

	dialect_func := dialectFunc{
		Receiver: fmt.Sprintf("%sImpl", dialect.Name()),
	}

	tmpl, err := r.loadDialect(dialect)
	if err != nil {
		return err
	}

	return tmplutil.Render(tmpl, w, "is-constraint-error", dialect_func)
}

func (r *Renderer) renderDialectOpens(w io.Writer, dialects []sql.Dialect) (
	err error) {

	for _, dialect := range dialects {
		tmpl, err := r.loadDialect(dialect)
		if err != nil {
			return err
		}
		if err := tmplutil.Render(tmpl, w, "open", nil); err != nil {
			return err
		}
	}

	return nil
}

func (r *Renderer) loadDialect(dialect sql.Dialect) (
	*template.Template, error) {

	return r.loader.Load(
		fmt.Sprintf("golang.dialect-%s.tmpl", dialect.Name()), nil)
}
