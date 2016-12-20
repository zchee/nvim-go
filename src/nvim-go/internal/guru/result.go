// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package guru

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/cmd/guru/serial"
	"golang.org/x/tools/go/ast/astutil"
)

// callees
func (r *calleesSSAResult) Result(fset *token.FileSet) interface{} {
	j := &serial.Callees{
		Pos:  fset.Position(r.site.Pos()).String(),
		Desc: r.site.Common().Description(),
	}
	for _, callee := range r.funcs {
		j.Callees = append(j.Callees, &serial.Callee{
			Name: callee.String(),
			Pos:  fset.Position(callee.Pos()).String(),
		})
	}
	return j
}

func (r *calleesTypesResult) Result(fset *token.FileSet) interface{} {
	j := &serial.Callees{
		Pos:  fset.Position(r.site.Pos()).String(),
		Desc: "static function call",
	}
	j.Callees = []*serial.Callee{
		&serial.Callee{
			Name: r.callee.FullName(),
			Pos:  fset.Position(r.callee.Pos()).String(),
		},
	}
	return j
}

// callers
func (r *callersResult) Result(fset *token.FileSet) interface{} {
	var callers []serial.Caller
	for _, edge := range r.edges {
		callers = append(callers, serial.Caller{
			Caller: edge.Caller.Func.String(),
			Pos:    fset.Position(edge.Pos()).String(),
			Desc:   edge.Description(),
		})
	}
	return callers
}

// callstack
func (r *callstackResult) Result(fset *token.FileSet) interface{} {
	var callers []serial.Caller
	for i := len(r.callpath) - 1; i >= 0; i-- { // (innermost first)
		edge := r.callpath[i]
		callers = append(callers, serial.Caller{
			Pos:    fset.Position(edge.Pos()).String(),
			Caller: edge.Caller.Func.String(),
			Desc:   edge.Description(),
		})
	}
	return &serial.CallStack{
		Pos:     fset.Position(r.target.Pos()).String(),
		Target:  r.target.String(),
		Callers: callers,
	}
}

// definition
func (r *definitionResult) Result(fset *token.FileSet) interface{} {
	return &serial.Definition{
		Desc:   r.descr,
		ObjPos: fset.Position(r.pos).String(),
	}
}

// describe
func (r *describeUnknownResult) Result(fset *token.FileSet) interface{} {
	return &serial.Describe{
		Desc: astutil.NodeDescription(r.node),
		Pos:  fset.Position(r.node.Pos()).String(),
	}
}

func (r *describeValueResult) Result(fset *token.FileSet) interface{} {
	var value, objpos string
	if r.constVal != nil {
		value = r.constVal.String()
	}
	if r.obj != nil {
		objpos = fset.Position(r.obj.Pos()).String()
	}

	return &serial.Describe{
		Desc:   astutil.NodeDescription(r.expr),
		Pos:    fset.Position(r.expr.Pos()).String(),
		Detail: "value",
		Value: &serial.DescribeValue{
			Type:   r.qpos.typeString(r.typ),
			Value:  value,
			ObjPos: objpos,
		},
	}
}

func (r *describeTypeResult) Result(fset *token.FileSet) interface{} {
	var namePos, nameDef string
	if nt, ok := r.typ.(*types.Named); ok {
		namePos = fset.Position(nt.Obj().Pos()).String()
		nameDef = nt.Underlying().String()
	}
	return &serial.Describe{
		Desc:   r.description,
		Pos:    fset.Position(r.node.Pos()).String(),
		Detail: "type",
		Type: &serial.DescribeType{
			Type:    r.qpos.typeString(r.typ),
			NamePos: namePos,
			NameDef: nameDef,
			Methods: methodsToSerial(r.qpos.info.Pkg, r.methods, fset),
		},
	}
}

func (r *describePackageResult) Result(fset *token.FileSet) interface{} {
	var members []*serial.DescribeMember
	for _, mem := range r.members {
		typ := mem.obj.Type()
		var val string
		switch mem := mem.obj.(type) {
		case *types.Const:
			val = mem.Val().String()
		case *types.TypeName:
			typ = typ.Underlying()
		}
		members = append(members, &serial.DescribeMember{
			Name:    mem.obj.Name(),
			Type:    typ.String(),
			Value:   val,
			Pos:     fset.Position(mem.obj.Pos()).String(),
			Kind:    tokenOf(mem.obj),
			Methods: methodsToSerial(r.pkg, mem.methods, fset),
		})
	}
	return &serial.Describe{
		Desc:   r.description,
		Pos:    fset.Position(r.node.Pos()).String(),
		Detail: "package",
		Package: &serial.DescribePackage{
			Path:    r.pkg.Path(),
			Members: members,
		},
	}
}

func (r *describeStmtResult) Result(fset *token.FileSet) interface{} {
	return &serial.Describe{
		Desc:   r.description,
		Pos:    fset.Position(r.node.Pos()).String(),
		Detail: "unknown",
	}
}

// freevars
func (r *freevarsResult) Result(fset *token.FileSet) interface{} {
	var out []serial.FreeVar
	for _, ref := range r.refs {
		out = append(out, serial.FreeVar{
			Pos:  fset.Position(ref.obj.Pos()).String(),
			Kind: ref.kind,
			Ref:  ref.ref,
			Type: ref.typ.String(),
		})
	}
	return out
}

// implements
func (r *implementsResult) Result(fset *token.FileSet) interface{} {
	var method *serial.DescribeMethod
	if r.method != nil {
		method = &serial.DescribeMethod{
			Name: r.qpos.objectString(r.method),
			Pos:  fset.Position(r.method.Pos()).String(),
		}
	}
	return &serial.Implements{
		T:                       makeImplementsType(r.t, fset),
		AssignableTo:            makeImplementsTypes(r.to, fset),
		AssignableFrom:          makeImplementsTypes(r.from, fset),
		AssignableFromPtr:       makeImplementsTypes(r.fromPtr, fset),
		AssignableToMethod:      methodsToSerial(r.qpos.info.Pkg, r.toMethod, fset),
		AssignableFromMethod:    methodsToSerial(r.qpos.info.Pkg, r.fromMethod, fset),
		AssignableFromPtrMethod: methodsToSerial(r.qpos.info.Pkg, r.fromPtrMethod, fset),
		Method:                  method,
	}

}

// peers
func (r *peersResult) Result(fset *token.FileSet) interface{} {
	peers := &serial.Peers{
		Pos:  fset.Position(r.queryPos).String(),
		Type: r.queryType.String(),
	}
	for _, alloc := range r.makes {
		peers.Allocs = append(peers.Allocs, fset.Position(alloc).String())
	}
	for _, send := range r.sends {
		peers.Sends = append(peers.Sends, fset.Position(send).String())
	}
	for _, receive := range r.receives {
		peers.Receives = append(peers.Receives, fset.Position(receive).String())
	}
	for _, clos := range r.closes {
		peers.Closes = append(peers.Closes, fset.Position(clos).String())
	}
	return peers
}

// pointsto
func (r *pointstoResult) Result(fset *token.FileSet) interface{} {
	var pts []serial.PointsTo
	for _, ptr := range r.ptrs {
		var namePos string
		if nt, ok := deref(ptr.typ).(*types.Named); ok {
			namePos = fset.Position(nt.Obj().Pos()).String()
		}
		var labels []serial.PointsToLabel
		for _, l := range ptr.labels {
			labels = append(labels, serial.PointsToLabel{
				Pos:  fset.Position(l.Pos()).String(),
				Desc: l.String(),
			})
		}
		pts = append(pts, serial.PointsTo{
			Type:    r.qpos.typeString(ptr.typ),
			NamePos: namePos,
			Labels:  labels,
		})
	}
	return pts
}

// referrers
func (r *referrersInitialResult) Result(fset *token.FileSet) interface{} {
	var objpos string
	if pos := r.obj.Pos(); pos.IsValid() {
		objpos = fset.Position(pos).String()
	}
	return &serial.ReferrersInitial{
		Desc:   r.obj.String(),
		ObjPos: objpos,
	}
}

func (r *referrersPackageResult) Result(fset *token.FileSet) interface{} {
	refs := serial.ReferrersPackage{Package: r.pkg.Path()}
	r.foreachRef(func(id *ast.Ident, text string) {
		refs.Refs = append(refs.Refs, serial.Ref{
			Pos:  fset.Position(id.NamePos).String(),
			Text: text,
		})
	})
	return refs
}

// what
func (r *whatResult) Result(fset *token.FileSet) interface{} {
	var enclosing []serial.SyntaxNode
	for _, n := range r.path {
		enclosing = append(enclosing, serial.SyntaxNode{
			Description: astutil.NodeDescription(n),
			Start:       fset.Position(n.Pos()).Offset,
			End:         fset.Position(n.End()).Offset,
		})
	}

	var sameids []string
	for _, pos := range r.sameids {
		sameids = append(sameids, fset.Position(pos).String())
	}

	return &serial.What{
		Modes:      r.modes,
		SrcDir:     r.srcdir,
		ImportPath: r.importPath,
		Enclosing:  enclosing,
		Object:     r.object,
		SameIDs:    sameids,
	}
}

// whicherrs
func (r *whicherrsResult) Result(fset *token.FileSet) interface{} {
	we := &serial.WhichErrs{}
	we.ErrPos = fset.Position(r.errpos).String()
	for _, g := range r.globals {
		we.Globals = append(we.Globals, fset.Position(g.Pos()).String())
	}
	for _, c := range r.consts {
		we.Constants = append(we.Constants, fset.Position(c.Pos()).String())
	}
	for _, t := range r.types {
		var et serial.WhichErrsType
		et.Type = r.qpos.typeString(t.typ)
		et.Position = fset.Position(t.obj.Pos()).String()
		we.Types = append(we.Types, et)
	}
	return we
}
