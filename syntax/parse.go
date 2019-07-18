// Copyright (C) 2019 Storj Labs, Inc.
// Copyright (C) 2017 Space Monkey, Inc.
// See LICENSE for copying information.

package syntax

import "storj.io/dbx/ast"

func Parse(path string, data []byte) (root *ast.Root, err error) {
	scanner, err := NewScanner(path, data)
	if err != nil {
		return nil, err
	}

	return parseRoot(scanner)
}

func parseRoot(scanner *Scanner) (*ast.Root, error) {
	list, err := scanRoot(scanner)
	if err != nil {
		return nil, err
	}

	root := new(ast.Root)

	err = list.consumeAnyTuples(tupleCases{
		"model": func(node *tupleNode) error {
			model, err := parseModel(node)
			if err != nil {
				return err
			}
			root.Models = append(root.Models, model)

			return nil
		},
		"create": func(node *tupleNode) error {
			cre, err := parseCreate(node)
			if err != nil {
				return err
			}
			root.Creates = append(root.Creates, cre)

			return nil
		},
		"read": func(node *tupleNode) error {
			read, err := parseRead(node)
			if err != nil {
				return err
			}
			root.Reads = append(root.Reads, read)

			return nil
		},
		"update": func(node *tupleNode) error {
			upd, err := parseUpdate(node)
			if err != nil {
				return err
			}
			root.Updates = append(root.Updates, upd)

			return nil
		},
		"delete": func(node *tupleNode) error {
			del, err := parseDelete(node)
			if err != nil {
				return err
			}
			root.Deletes = append(root.Deletes, del)

			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return root, nil
}
