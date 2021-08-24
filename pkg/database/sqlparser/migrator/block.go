// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package migrator

import "github.com/pingcap/parser/ast"

type StmtType string

const (
	DDL = "DDL"
	DML = "DML"
)

type Block struct {
	typ   StmtType
	nodes []ast.Node
}

func (block Block) Type() StmtType {
	return block.typ
}

func (block Block) Nodes() []ast.Node {
	return block.nodes
}

func AppendBlock(blocks []Block, node ast.Node, typ StmtType) []Block {
	if len(blocks) == 0 {
		return []Block{{typ: typ, nodes: []ast.Node{node}}}
	}
	last := len(blocks) - 1
	if blocks[last].Type() == typ {
		blocks[last].nodes = append(blocks[last].nodes, node)
		return blocks
	}
	return append(blocks, Block{
		typ:   typ,
		nodes: []ast.Node{node},
	})
}
