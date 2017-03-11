/*
 * gomacro - A Go intepreter with Lisp-like macros
 *
 * Copyright (C) 2017 Massimiliano Ghilardi
 *
 *     This program is free software you can redistribute it and/or modify
 *     it under the terms of the GNU General Public License as published by
 *     the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *
 *     This program is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU General Public License for more details.
 *
 *     You should have received a copy of the GNU General Public License
 *     along with this program.  If not, see <http//www.gnu.org/licenses/>.
 *
 * ast_slice.go
 *
 *  Created on Feb 25, 2017
 *      Author Massimiliano Ghilardi
 */

package interpreter

import (
	"go/ast"
	"go/token"
)

// Ast wrappers for variable-length slices of ast.Nodes - they are not full-blown ast.Node

func (x NodeSlice) Interface() interface{}  { return x.p }
func (x ExprSlice) Interface() interface{}  { return x.p }
func (x FieldSlice) Interface() interface{} { return x.p }
func (x DeclSlice) Interface() interface{}  { return x.p }
func (x IdentSlice) Interface() interface{} { return x.p }
func (x SpecSlice) Interface() interface{}  { return x.p }
func (x StmtSlice) Interface() interface{}  { return x.p }

func (x NodeSlice) Op() token.Token  { return token.COMMA }     // FIXME
func (x ExprSlice) Op() token.Token  { return token.COMMA }     // FIXME
func (x FieldSlice) Op() token.Token { return token.SEMICOLON } // FIXME
func (x DeclSlice) Op() token.Token  { return token.SEMICOLON } // FIXME
func (x IdentSlice) Op() token.Token { return token.COMMA }     // FIXME
func (x SpecSlice) Op() token.Token  { return token.SEMICOLON } // FIXME
func (x StmtSlice) Op() token.Token  { return token.SEMICOLON } // FIXME

func (x NodeSlice) New() Ast  { return NodeSlice{p: []ast.Node{}} }
func (x ExprSlice) New() Ast  { return ExprSlice{p: []ast.Expr{}} }
func (x FieldSlice) New() Ast { return FieldSlice{p: []*ast.Field{}} }
func (x DeclSlice) New() Ast  { return DeclSlice{p: []ast.Decl{}} }
func (x IdentSlice) New() Ast { return IdentSlice{p: []*ast.Ident{}} }
func (x SpecSlice) New() Ast  { return SpecSlice{p: []ast.Spec{}} }
func (x StmtSlice) New() Ast  { return StmtSlice{p: []ast.Stmt{}} }

func (x NodeSlice) Size() int  { return len(x.p) }
func (x ExprSlice) Size() int  { return len(x.p) }
func (x FieldSlice) Size() int { return len(x.p) }
func (x DeclSlice) Size() int  { return len(x.p) }
func (x IdentSlice) Size() int { return len(x.p) }
func (x SpecSlice) Size() int  { return len(x.p) }
func (x StmtSlice) Size() int  { return len(x.p) }

func (x NodeSlice) Get(i int) Ast  { return ToAst(x.p[i]) }
func (x ExprSlice) Get(i int) Ast  { return ToAst(x.p[i]) }
func (x FieldSlice) Get(i int) Ast { return ToAst(x.p[i]) }
func (x DeclSlice) Get(i int) Ast  { return ToAst(x.p[i]) }
func (x IdentSlice) Get(i int) Ast { return ToAst(x.p[i]) }
func (x SpecSlice) Get(i int) Ast  { return ToAst(x.p[i]) }
func (x StmtSlice) Get(i int) Ast  { return ToAst(x.p[i]) }

func (x NodeSlice) Set(i int, child Ast)  { x.p[i] = ToNode(child) }
func (x ExprSlice) Set(i int, child Ast)  { x.p[i] = ToExpr(child) }
func (x FieldSlice) Set(i int, child Ast) { x.p[i] = ToField(child) }
func (x DeclSlice) Set(i int, child Ast)  { x.p[i] = ToDecl(child) }
func (x IdentSlice) Set(i int, child Ast) { x.p[i] = ToIdent(child) }
func (x SpecSlice) Set(i int, child Ast)  { x.p[i] = ToSpec(child) }
func (x StmtSlice) Set(i int, child Ast)  { x.p[i] = ToStmt(child) }

func (x NodeSlice) Slice(lo, hi int) AstWithSlice  { x.p = x.p[lo:hi]; return x }
func (x ExprSlice) Slice(lo, hi int) AstWithSlice  { x.p = x.p[lo:hi]; return x }
func (x FieldSlice) Slice(lo, hi int) AstWithSlice { x.p = x.p[lo:hi]; return x }
func (x DeclSlice) Slice(lo, hi int) AstWithSlice  { x.p = x.p[lo:hi]; return x }
func (x IdentSlice) Slice(lo, hi int) AstWithSlice { x.p = x.p[lo:hi]; return x }
func (x SpecSlice) Slice(lo, hi int) AstWithSlice  { x.p = x.p[lo:hi]; return x }
func (x StmtSlice) Slice(lo, hi int) AstWithSlice  { x.p = x.p[lo:hi]; return x }

func (x NodeSlice) Append(child Ast) AstWithSlice  { x.p = append(x.p, ToNode(child)); return x }
func (x ExprSlice) Append(child Ast) AstWithSlice  { x.p = append(x.p, ToExpr(child)); return x }
func (x FieldSlice) Append(child Ast) AstWithSlice { x.p = append(x.p, ToField(child)); return x }
func (x DeclSlice) Append(child Ast) AstWithSlice  { x.p = append(x.p, ToDecl(child)); return x }
func (x IdentSlice) Append(child Ast) AstWithSlice { x.p = append(x.p, ToIdent(child)); return x }
func (x SpecSlice) Append(child Ast) AstWithSlice  { x.p = append(x.p, ToSpec(child)); return x }
func (x StmtSlice) Append(child Ast) AstWithSlice  { x.p = append(x.p, ToStmt(child)); return x }

// variable-length ast.Nodes

func (x BlockStmt) Interface() interface{}  { return x.p }
func (x FieldList) Interface() interface{}  { return x.p }
func (x File) Interface() interface{}       { return x.p }
func (x GenDecl) Interface() interface{}    { return x.p }
func (x ReturnStmt) Interface() interface{} { return x.p }

func (x BlockStmt) Node() ast.Node  { return x.p }
func (x FieldList) Node() ast.Node  { return x.p }
func (x File) Node() ast.Node       { return x.p }
func (x GenDecl) Node() ast.Node    { return x.p }
func (x ReturnStmt) Node() ast.Node { return x.p }

func (x BlockStmt) Op() token.Token  { return token.LBRACE }
func (x FieldList) Op() token.Token  { return token.ELLIPSIS }
func (x File) Op() token.Token       { return token.EOF }
func (x GenDecl) Op() token.Token    { return x.p.Tok }
func (x ReturnStmt) Op() token.Token { return token.RETURN }

func (x BlockStmt) New() Ast { return BlockStmt{&ast.BlockStmt{Lbrace: x.p.Lbrace, Rbrace: x.p.Rbrace}} }
func (x FieldList) New() Ast { return FieldList{&ast.FieldList{}} }
func (x File) New() Ast {
	return File{&ast.File{Doc: x.p.Doc, Package: x.p.Package, Name: x.p.Name, Scope: x.p.Scope, Imports: x.p.Imports, Comments: x.p.Comments}}
}
func (x GenDecl) New() Ast {
	return GenDecl{&ast.GenDecl{Doc: x.p.Doc, TokPos: x.p.TokPos, Tok: x.p.Tok, Lparen: x.p.Lparen, Rparen: x.p.Rparen}}
}
func (x ReturnStmt) New() Ast { return ReturnStmt{&ast.ReturnStmt{Return: x.p.Return}} }

func (x BlockStmt) Size() int  { return len(x.p.List) }
func (x FieldList) Size() int  { return len(x.p.List) }
func (x File) Size() int       { return len(x.p.Decls) }
func (x GenDecl) Size() int    { return len(x.p.Specs) }
func (x ReturnStmt) Size() int { return len(x.p.Results) }

func (x BlockStmt) Get(i int) Ast  { return ToAst(x.p.List[i]) }
func (x FieldList) Get(i int) Ast  { return ToAst(x.p.List[i]) }
func (x File) Get(i int) Ast       { return ToAst(x.p.Decls[i]) }
func (x GenDecl) Get(i int) Ast    { return ToAst(x.p.Specs[i]) }
func (x ReturnStmt) Get(i int) Ast { return ToAst(x.p.Results[i]) }

func (x BlockStmt) Set(i int, child Ast)  { x.p.List[i] = ToStmt(child) }
func (x FieldList) Set(i int, child Ast)  { x.p.List[i] = ToField(child) }
func (x File) Set(i int, child Ast)       { x.p.Decls[i] = ToDecl(child) }
func (x GenDecl) Set(i int, child Ast)    { x.p.Specs[i] = ToSpec(child) }
func (x ReturnStmt) Set(i int, child Ast) { x.p.Results[i] = ToExpr(child) }

func (x BlockStmt) Slice(lo, hi int) AstWithSlice  { x.p.List = x.p.List[lo:hi]; return x }
func (x FieldList) Slice(lo, hi int) AstWithSlice  { x.p.List = x.p.List[lo:hi]; return x }
func (x File) Slice(lo, hi int) AstWithSlice       { x.p.Decls = x.p.Decls[lo:hi]; return x }
func (x GenDecl) Slice(lo, hi int) AstWithSlice    { x.p.Specs = x.p.Specs[lo:hi]; return x }
func (x ReturnStmt) Slice(lo, hi int) AstWithSlice { x.p.Results = x.p.Results[lo:hi]; return x }

func (x BlockStmt) Append(child Ast) AstWithSlice {
	x.p.List = append(x.p.List, ToStmt(child))
	return x
}
func (x FieldList) Append(child Ast) AstWithSlice {
	x.p.List = append(x.p.List, ToField(child))
	return x
}
func (x File) Append(child Ast) AstWithSlice {
	x.p.Decls = append(x.p.Decls, ToDecl(child))
	return x
}
func (x GenDecl) Append(child Ast) AstWithSlice {
	x.p.Specs = append(x.p.Specs, ToSpec(child))
	return x
}
func (x ReturnStmt) Append(child Ast) AstWithSlice {
	x.p.Results = append(x.p.Results, ToExpr(child))
	return x
}