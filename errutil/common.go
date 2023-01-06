// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package errutil

import (
	"fmt"
	"strings"
	"text/scanner"

	"github.com/spacemonkeygo/errors"
)

var (
	Error = errors.NewClass("dbx")
)

var errorPosition = errors.GenSym()

func New(pos scanner.Position, format string, args ...interface{}) error {
	str := fmt.Sprintf("%s: %s", pos, fmt.Sprintf(format, args...))
	return Error.NewWith(str, SetErrorPosition(pos))
}

func SetErrorPosition(pos scanner.Position) errors.ErrorOption {
	return errors.SetData(errorPosition, pos)
}

func GetErrorPosition(err error) *scanner.Position {
	pos, ok := errors.GetData(err, errorPosition).(scanner.Position)
	if ok {
		return &pos
	}
	return nil
}

func GetContext(sources map[string][]byte, err error) string {
	if len(sources) == 0 {
		return ""
	}
	if pos := GetErrorPosition(err); pos != nil {
		return generateContext(sources, *pos)
	}
	return ""
}

func lineAround(data []byte, offset int) (start, end int) {
	// find the index of the '\n' before data[offset]
	start = 0
	for i := offset - 1; i >= 0; i-- {
		if data[i] == '\n' {
			start = i + 1
			break
		}
	}

	// find the index of the '\n' after data[offset]
	end = len(data)
	for i := offset; i < len(data); i++ {
		if data[i] == '\n' {
			end = i
			break
		}
	}

	return start, end
}

func generateContext(sources map[string][]byte, pos scanner.Position) (context string) {
	var context_bytes []byte

	source := sources[pos.Filename]
	if pos.Offset > len(source) {
		panic("internal error: underline on strange position")
	}

	line_start, line_end := lineAround(source, pos.Offset)
	line := string(source[line_start:line_end])

	var before_line string
	if line_start > 0 {
		before_start, before_end := lineAround(source, line_start-1)
		before_line = string(source[before_start:before_end])
		before_line = strings.Replace(before_line, "\t", "    ", -1)
		context_bytes = append(context_bytes,
			fmt.Sprintf("% 4d: ", pos.Line-1)...)
		context_bytes = append(context_bytes, before_line...)
		context_bytes = append(context_bytes, '\n')
	}

	tabs := strings.Count(line, "\t")
	line = strings.Replace(line, "\t", "    ", -1)
	context_bytes = append(context_bytes, fmt.Sprintf("% 4d: ", pos.Line)...)
	context_bytes = append(context_bytes, line...)
	context_bytes = append(context_bytes, '\n')

	offset := tabs*4 + pos.Column - 1 - tabs + 6
	underline := strings.Repeat(" ", offset) + "^"
	context_bytes = append(context_bytes, underline...)

	return string(context_bytes)
}
