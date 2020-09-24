// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package golang

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"storj.io/dbx/internal/inflect"
	"storj.io/dbx/ir"
	"storj.io/dbx/sqlgen/sqlembedgo"
)

func renameFn(name string, v *Var) *Var {
	vc := v.Copy()
	vc.Name = name
	return vc
}

func sliceofFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).SliceOf)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func paramFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).Param)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ",\n"), nil
}

func argFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).Arg)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func valueFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).Value)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func addrofFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).AddrOf)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func initFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).Init)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, "\n"), nil
}

func initnewFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).InitNew)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, "\n"), nil
}

func declareFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).Declare)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, "\n"), nil
}

func zeroFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).Zero)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func flattenFn(intf interface{}) (flattened []*Var, err error) {
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

func ctxparamFn(intf interface{}) (string, error) {
	param, err := paramFn(intf)
	if err != nil {
		return "", err
	}
	if param == "" {
		return "ctx context.Context", nil
	}
	return "ctx context.Context,\n" + param, nil
}

func ctxargFn(intf interface{}) (string, error) {
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

func sliceFn(start, end int, intf interface{}) interface{} {
	rv := reflect.ValueOf(intf)
	if start < 0 {
		start += rv.Len()
	}
	if end < 0 {
		end += rv.Len()
	}
	return rv.Slice(start, end).Interface()
}

func forVars(intf interface{}, fn func(v *Var) string) ([]string, error) {
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
		fmt.Fprintf(&out, "var %s = %s\n", hole.Name, hole.Expression)
	}

	for _, cond := range info.Conditions {
		fmt.Fprintf(&out, "var %s = %s\n", cond.Name, cond.Expression)
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
				fmt.Fprintf(&out, "%s = append(%s, %s)\n", name, name, strings.Join(run, ", "))
				run = run[:0]
			}
			fmt.Fprintf(&out, "if !%s.isnull() {\n", arg.Var.Name)
			fmt.Fprintf(&out, "\t__cond_%d.Null = false\n", arg.Condition)
			fmt.Fprintf(&out, "\t%s = append(%s, %s.value())\n", name, name, arg.Var.Name)
			fmt.Fprintf(&out, "}\n")
		} else {
			run = append(run, fmt.Sprintf("%s.value()", arg.Var.Name))
		}
	}
	if len(run) > 0 {
		fmt.Fprintf(&out, "%s = append(%s, %s)\n", name, name, strings.Join(run, ", "))
	}

	return out.String()
}
