// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package golang

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"storj.io/dbx/consts"
	"storj.io/dbx/internal/inflect"
	"storj.io/dbx/ir"
	"storj.io/dbx/sqlgen/sqlembedgo"
)

func funcMap(dialect string) template.FuncMap {
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

		"add": func(arg0 int, args ...int) int {
			total := arg0
			for _, v := range args {
				total += v
			}
			return total
		},
	}

	if dialect == "spanner" {
		funcs["initnew"] = spanner_initnewFn
		funcs["embedvalues"] = spanner_embedvaluesFn
		funcs["setupdatablefields"] = spanner_setupdatablefieldsFn
		funcs["setoptionalfields"] = spanner_setoptionalfieldsFn
		funcs["addrof"] = spanner_addrofFn
	}

	return funcs
}

func renameFn(name string, v *Var) *Var {
	vc := v.Copy()
	vc.Name = name
	return vc
}

func sliceofFn(intf any) (string, error) {
	vs, err := forVars(intf, (*Var).SliceOf)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func paramFn(intf any) (string, error) {
	vs, err := forVars(intf, (*Var).Param)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ",\n"), nil
}

func argFn(intf any) (string, error) {
	vs, err := forVars(intf, (*Var).Arg)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func valueFn(intf any) (string, error) {
	vs, err := forVars(intf, (*Var).Value)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func addrofFn(intf any) (string, error) {
	vs, err := forVars(intf, (*Var).AddrOf)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func initFn(intf any) (string, error) {
	vs, err := forVars(intf, (*Var).Init)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, "\n"), nil
}

func initnewFn(intf any) (string, error) {
	vs, err := forVars(intf, (*Var).InitNew)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, "\n"), nil
}

func declareFn(intf any) (string, error) {
	vs, err := forVars(intf, (*Var).Declare)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, "\n"), nil
}

func zeroFn(intf any) (string, error) {
	vs, err := forVars(intf, (*Var).Zero)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func flattenFn(intf any) (flattened []*Var, err error) {
	switch obj := intf.(type) {
	case *Var:
		flattened = obj.Flatten()
	case []*Var:
		for _, v := range obj {
			flattened = append(flattened, v.Flatten()...)
		}
	default:
		return nil, Error.New("unsupported type %T", obj)
	}
	return flattened, nil
}

func ctxparamFn(intf any) (string, error) {
	param, err := paramFn(intf)
	if err != nil {
		return "", err
	}
	if param == "" {
		return "ctx context.Context", nil
	}
	return "ctx context.Context,\n" + param, nil
}

func ctxargFn(intf any) (string, error) {
	arg, err := argFn(intf)
	if err != nil {
		return "", err
	}
	if arg == "" {
		return "ctx", nil
	}
	return "ctx, " + arg, nil
}

func commaFn(in string) string {
	if in == "" {
		return ""
	}
	return in + ", "
}

func doubleFn(vs []*Var) (out []*Var) {
	for _, v := range vs {
		out = append(out, v, v)
	}
	return out
}

func sliceFn(start, end int, intf any) any {
	rv := reflect.ValueOf(intf)
	if start < 0 {
		start += rv.Len()
	}
	if end < 0 {
		end += rv.Len()
	}
	return rv.Slice(start, end).Interface()
}

func forVars(intf any, fn func(v *Var) string) ([]string, error) {
	var elems []string
	switch obj := intf.(type) {
	case ConditionArg:
		elems = append(elems, fn(obj.Var))
	case []ConditionArg:
		for _, arg := range obj {
			elems = append(elems, fn(arg.Var))
		}
	case *Var:
		elems = append(elems, fn(obj))
	case []*Var:
		for _, v := range obj {
			elems = append(elems, fn(v))
		}
	default:
		return nil, Error.New("unsupported type %T", obj)
	}
	return elems, nil
}

func structName(m *ir.Model) string {
	return inflect.Camelize(m.Name)
}

func fieldName(f *ir.Field) string {
	return inflect.Camelize(f.Name)
}

func convertSuffix(suffix []string) string {
	parts := make([]string, 0, len(suffix))
	for _, part := range suffix {
		parts = append(parts, inflect.Camelize(part))
	}
	return strings.Join(parts, "_")
}

func embedplaceholdersFn(info sqlembedgo.Info) string {
	var out bytes.Buffer

	for _, hole := range info.Holes {
		_, _ = fmt.Fprintf(&out, "var %s = %s\n", hole.Name, hole.Expression)
	}

	for _, cond := range info.Conditions {
		_, _ = fmt.Fprintf(&out, "var %s = %s\n", cond.Name, cond.Expression)
	}

	return out.String()
}

func embedsqlFn(info sqlembedgo.Info, name string) string {
	return fmt.Sprintf("var %s = %s\n", name, info.Expression)
}

func embedvaluesFn(args []ConditionArg, name string) string {
	var out bytes.Buffer
	var run []string

	for _, arg := range args {
		if arg.IsCondition {
			if len(run) > 0 {
				_, _ = fmt.Fprintf(&out, "%s = append(%s, %s)\n", name, name, strings.Join(run, ", "))
				run = run[:0]
			}
			_, _ = fmt.Fprintf(&out, "if !%s.isnull() {\n", arg.Var.Name)
			_, _ = fmt.Fprintf(&out, "\t__cond_%d.Null = false\n", arg.Condition)
			_, _ = fmt.Fprintf(&out, "\t%s = append(%s, %s.value())\n", name, name, arg.Var.Name)
			_, _ = fmt.Fprintf(&out, "}\n")
		} else {
			run = append(run, fmt.Sprintf("%s.value()", arg.Var.Name))
		}
	}
	if len(run) > 0 {
		_, _ = fmt.Fprintf(&out, "%s = append(%s, %s)\n", name, name, strings.Join(run, ", "))
	}

	return out.String()
}

func spanner_embedvaluesFn(args []ConditionArg, name string) string {
	var out bytes.Buffer
	var run []string

	for _, arg := range args {
		if arg.IsCondition {
			if len(run) > 0 {
				_, _ = fmt.Fprintf(&out, "%s = append(%s, %s)\n", name, name, strings.Join(run, ", "))
				run = run[:0]
			}

			_, _ = fmt.Fprintf(&out, "if !%s.isnull() {\n", arg.Var.Name)
			_, _ = fmt.Fprintf(&out, "\t__cond_%d.Null = false\n", arg.Condition)

			if wrap := spannerWrapFunc(arg.Var.Underlying); wrap != "" {
				_, _ = fmt.Fprintf(&out, "\t%s = append(%s, %v(%s.value()))\n", name, name, wrap, arg.Var.Name)
			} else {
				_, _ = fmt.Fprintf(&out, "\t%s = append(%s, %s.value())\n", name, name, arg.Var.Name)
			}

			_, _ = fmt.Fprintf(&out, "}\n")
		} else {
			if wrap := spannerWrapFunc(arg.Var.Underlying); wrap != "" {
				run = append(run, fmt.Sprintf("%v(%s.value())", wrap, arg.Var.Name))
			} else {
				run = append(run, fmt.Sprintf("%s.value()", arg.Var.Name))
			}
		}
	}
	if len(run) > 0 {
		_, _ = fmt.Fprintf(&out, "%s = append(%s, %s)\n", name, name, strings.Join(run, ", "))
	}

	return out.String()
}

func spanner_setupdatablefieldsFn(modelFields []*ModelField) string {
	var out bytes.Buffer

	for _, field := range modelFields {
		if field == nil {
			continue
		}

		_, _ = fmt.Fprintf(&out, "if update.%s._set {\n", field.Name)
		if wrap := spannerWrapFunc(field.Underlying); wrap != "" {
			_, _ = fmt.Fprintf(&out, "\t__values = append(__values, %v(update.%s.value()))\n", wrap, field.Name)
		} else {
			_, _ = fmt.Fprintf(&out, "\t__values = append(__values, update.%s.value())\n", field.Name)
		}
		_, _ = fmt.Fprintf(&out, "\t__sets_sql.SQLs = append(__sets_sql.SQLs, __sqlbundle_Literal(\"%s = ?\"))\n", field.Column)
		_, _ = fmt.Fprintf(&out, "}\n")
	}
	return out.String()
}

func spanner_setoptionalfieldsFn(modelFields []*ModelField) string {
	var out bytes.Buffer

	for _, field := range modelFields {
		if field == nil {
			continue
		}

		_, _ = fmt.Fprintf(&out, "if optional.%s._set {\n", field.Name)
		if wrap := spannerWrapFunc(field.Underlying); wrap != "" {
			_, _ = fmt.Fprintf(&out, "\t__values = append(__values, %v(optional.%s.value()))\n", wrap, field.Name)
		} else {
			_, _ = fmt.Fprintf(&out, "\t__values = append(__values, optional.%s.value())\n", field.Name)
		}
		_, _ = fmt.Fprintf(&out, "\t__optional_columns.SQLs = append(__optional_columns.SQLs, __sqlbundle_Literal(\"%s\"))\n", field.Column)
		_, _ = fmt.Fprintf(&out, "\t__optional_placeholders.SQLs = append(__optional_placeholders.SQLs, __sqlbundle_Literal(\"?\"))\n")
		_, _ = fmt.Fprintf(&out, "}\n")
	}
	return out.String()
}

func spanner_initnewFn(intf any) (string, error) {
	vs, err := forVars(intf, spanner_Var_InitNew)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, "\n"), nil
}

func spanner_Var_InitNew(v *Var) string {
	if wrap := spannerWrapFunc(v.Underlying); wrap != "" {
		return fmt.Sprintf("%s := %v(%s)", v.Name, wrap, v.InitVal)
	} else {
		return fmt.Sprintf("%s := %s", v.Name, v.InitVal)
	}
}

func spanner_addrofFn(intf any) (string, error) {
	vs, err := forVars(intf, spanner_Var_AddrOf)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func spanner_Var_AddrOf(v *Var) string {
	if v.Underlying.Type == consts.JsonField {
		wrap := spannerWrapFunc(v.Underlying)
		return fmt.Sprintf("%v(&%s)", wrap, v.Name)
	} else {
		return fmt.Sprintf("&%s", v.Name)
	}
}

func spannerWrapFunc(v UnderlyingType) string {
	switch v.Type {
	case consts.JsonField:
		return "spannerConvertJSON"
	default:
		return ""
	}
}
