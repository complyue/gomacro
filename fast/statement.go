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
 * statement.go
 *
 *  Created on Apr 01, 2017
 *      Author Massimiliano Ghilardi
 */

package fast

import (
	"go/ast"
	"go/token"
	r "reflect"

	. "github.com/cosmos72/gomacro/base"
)

func stmtNop(env *Env) (Stmt, *Env) {
	env.IP++
	return env.Code[env.IP], env
}

// declare a var instead of function: code.go needs the address of Interrupt
var Interrupt Stmt = func(env *Env) (Stmt, *Env) {
	return env.ThreadGlobals.Interrupt, env
}

func popEnv(env *Env) (Stmt, *Env) {
	outer := env.Outer
	outer.IP = env.IP + 1
	env.FreeEnv()
	// Debugf("PopEnv, IP = %d of %d", outer.IP, len(outer.Code))
	return outer.Code[outer.IP], outer
}

func (c *Comp) Stmt(node ast.Stmt) {
	label := ""
	for {
		if node != nil {
			c.Pos = node.Pos()
		}
		switch node := node.(type) {
		case nil:
		case *ast.AssignStmt:
			c.Assign(node)
		case *ast.BlockStmt:
			c.Block(node)
		case *ast.BranchStmt:
			c.Branch(node)
		case *ast.CaseClause:
			c.misplacedCase(node, node.List == nil)
		case *ast.CommClause:
			c.misplacedCase(node, node.Comm == nil)
		case *ast.DeclStmt:
			c.Decl(node.Decl)
		// case *ast.DeferStmt:
		//   c.DeferStmt(node.Call)
		case *ast.EmptyStmt:
			// nothing to do
		case *ast.ExprStmt:
			expr := c.Expr(node.X)
			if !expr.Const() {
				c.Code.Append(expr.AsStmt())
			}
		case *ast.ForStmt:
			c.For(node, label)
		case *ast.GoStmt:
			c.Go(node)
		case *ast.IfStmt:
			c.If(node)
		case *ast.IncDecStmt:
			c.IncDec(node)
		case *ast.LabeledStmt:
			// c.Label(node)
			label = node.Label.Name
			continue
		// case *ast.RangeStmt:
		//   c.Range(node)
		case *ast.ReturnStmt:
			c.Return(node)
		// case *ast.SelectStmt:
		//   c.Select(node, label)
		// case *ast.SendStmt:
		//   c.Send(node)
		// case *ast.SwitchStmt:
		//   c.Switch(node, label)
		// case *ast.TypeSwitchStmt:
		//   c.TypeSwitch(node, label)
		default:
			c.Errorf("unimplemented statement: %v <%v>", node, r.TypeOf(node))
		}
		return
	}
}

func (c *Comp) misplacedCase(node ast.Node, isdefault bool) {
	label := "case"
	if isdefault {
		label = "default"
	}
	c.Errorf("misplaced %s: not inside switch or select: %v <%v>", label, node, r.TypeOf(node))
}

// Block compiles a block statement, i.e. { ... }
func (c *Comp) Block(block *ast.BlockStmt) {
	if block == nil || len(block.List) == 0 {
		return
	}
	c.List(block.List)
}

// Block0 compiles a block statement, i.e. { ... }
func (c *Comp) List(list []ast.Stmt) {
	if len(list) == 0 {
		c.Errorf("Block0 invoked on empty statement list")
	}
	var nbinds [2]int // # of binds in the block

	c2, locals := c.pushEnvIfLocalBinds(&nbinds, list...)

	for _, node := range list {
		c2.Stmt(node)
	}

	c2.popEnvIfLocalBinds(locals, &nbinds, list...)

	// c.Debugf("Block compiled. inner *Comp = %#v", c2)
}

// Branch compiles a break, continue, fallthrough, goto or return statement
func (c *Comp) Branch(node *ast.BranchStmt) {
	switch node.Tok {
	case token.BREAK:
		c.Break(node)
	case token.CONTINUE:
		c.Continue(node)
	/*
		case token.FALLTHROUGH:
			c.FallThrough(node)
		case token.GOTO:
			c.Goto(node)
	*/
	default:
		c.Errorf("unimplemented branch statement: %v <%v>", node, r.TypeOf(node))
	}
}

// Break compiles a "break" statement
func (c *Comp) Break(node *ast.BranchStmt) {
	label := ""
	if node.Label != nil {
		label = node.Label.Name
	}
	upn := 0
	// do not cross function boundaries
	for o := c; o != nil && o.Func == nil; o = o.Outer {
		if o.Loop != nil && o.Loop.Break != nil {
			if len(label) == 0 || o.Loop.ThisLabel == label {
				// only keep a reference to the jump target, NOT TO THE WHOLE *Comp!
				c.compileJumpOut(upn, o.Loop.Break)
				return
			}
		}
		upn += o.UpCost // count how many Env:s we must exit at runtime
	}
	if len(label) != 0 {
		c.Errorf("break label not defined: %v", label)
	} else {
		c.Errorf("break outside for/switch")
	}
}

// Continue compiles a "continue" statement
func (c *Comp) Continue(node *ast.BranchStmt) {
	label := ""
	if node.Label != nil {
		label = node.Label.Name
	}
	upn := 0
	// do not cross function boundaries
	for o := c; o != nil && o.Func == nil; o = o.Outer {
		if o.Loop != nil && o.Loop.Continue != nil {
			if len(label) == 0 || o.Loop.ThisLabel == label {
				// only keep a reference to the jump target, NOT TO THE WHOLE *Comp!
				c.compileJumpOut(upn, o.Loop.Continue)
				return
			}
		}
		upn += o.UpCost // count how many Env:s we must exit at runtime
	}
	if len(label) != 0 {
		c.Errorf("continue label not defined: %v", label)
	} else {
		c.Errorf("continue outside for")
	}
}

// compileJumpOut compiles a break or continue statement
// ip is a pointer because the jump target may not be known yet... it will be filled later
func (c *Comp) compileJumpOut(upn int, ip *int) {
	var stmt Stmt
	switch upn {
	case 0:
		stmt = func(env *Env) (Stmt, *Env) {
			env.IP = *ip
			return env.Code[env.IP], env
		}
	case 1:
		stmt = func(env *Env) (Stmt, *Env) {
			env = env.Outer
			env.IP = *ip
			return env.Code[env.IP], env
		}
	case 2:
		stmt = func(env *Env) (Stmt, *Env) {
			env = env.Outer.Outer
			env.IP = *ip
			return env.Code[env.IP], env
		}
	default:
		stmt = func(env *Env) (Stmt, *Env) {
			env = env.Outer.Outer.Outer
			for i := 3; i < upn; i++ {
				env = env.Outer
			}
			env.IP = *ip
			return env.Code[env.IP], env
		}
	}
	c.Code.Append(stmt)
}

// For compiles a "for" statement
func (c *Comp) For(node *ast.ForStmt, label string) {
	initLocals := false
	var initBinds [2]int
	if node.Init != nil {
		c, initLocals = c.pushEnvIfLocalBinds(&initBinds, node.Init)
		c.Stmt(node.Init)
	}
	flag, fun, err := true, (func(*Env) bool)(nil), false // "for { }" without a condition means "for true { }"
	if node.Cond != nil {
		pred := c.Expr(node.Cond)
		flag, fun, err = pred.TryAsPred()
		if err {
			c.invalidPred(node.Cond, pred)
			return
		}
	}
	var jump struct{ Cond, Post, Break int }
	c.Loop = &LoopInfo{
		Continue:  &jump.Post,
		Break:     &jump.Break,
		ThisLabel: label,
	}

	// compile the condition, if not a constant
	jump.Cond = c.Code.Len()
	if fun != nil {
		c.Code.Append(func(env *Env) (Stmt, *Env) {
			var ip int
			if fun(env) {
				ip = env.IP + 1
				// Debugf("for: condition = true, iterating. IntBinds = %v", env.IntBinds)
			} else {
				// Debugf("for: condition = false, exiting. IntBinds = %v", env.IntBinds)
				ip = jump.Break
			}
			env.IP = ip
			return env.Code[ip], env
		})
	}
	// compile the body
	c.Block(node.Body)
	// compile the post
	if node.Post == nil {
		jump.Post = jump.Cond // no post statement. "continue" jumps to the condition
	} else {
		jump.Post = c.Code.Len()
		if containLocalBinds(node.Post) {
			c.Errorf("invalid for: cannot declare new variables in post statement: %v", node.Post)
		}
		c.Stmt(node.Post)
	}
	c.Code.Append(func(env *Env) (Stmt, *Env) {
		// jump back to the condition
		// Debugf("for: body executed, jumping back to condition. IntBinds = %v", env.IntBinds)
		// time.Sleep(time.Second / 10)
		ip := jump.Cond
		env.IP = ip
		return env.Code[ip], env
	})
	if fun == nil && !flag {
		// "for false { }" means that body, post and jump back to condition are never executed...
		// still compiled above (to check for errors) but drop the generated code
		c.Code.List = c.Code.List[0:jump.Cond]
	}
	jump.Break = c.Code.Len()
	if node.Init != nil {
		c = c.popEnvIfLocalBinds(initLocals, &initBinds, node.Init)
	}
}

// Go compiles a "go" statement i.e. a goroutine
func (c *Comp) Go(node *ast.GoStmt) {
	// we must create a new ThreadGlobals with a new Pool.
	// Ideally, the new ThreadGlobals could be created inside the call,
	// but that requires modifying the function being executed.
	// Instead, we create the new ThreadGlobals here and wrap it into an "unnecessary" Env
	// Thus we must create a corresponding "unnecessary" Comp and use it to compile the call
	c2 := NewComp(c)

	call := c2.callExpr(node.Call, nil)
	exprfun := call.Fun.AsX1()
	argfuns := call.MakeArgfuns()

	c2.Code.Append(func(env *Env) (Stmt, *Env) {
		// create a new Env to hold the new ThreadGlobals and (initially empty) Pool
		env2 := NewEnv4Func(env, 0, 0)
		tg := env.ThreadGlobals
		env2.ThreadGlobals = &ThreadGlobals{
			FileEnv: tg.FileEnv,
			TopEnv:  tg.TopEnv,
			// Interrupt, Signal, PoolSize and Pool are zero-initialized, fine with that
			Globals: tg.Globals,
		}
		// env2.MarkUsedByClosure() // redundant, done by exprfun(env2) below

		// function and arguments are evaluated in the caller's goroutine
		// using the new Env: we compiled them with c2 => execute them with env2
		funv := exprfun(env2)
		argv := make([]r.Value, len(argfuns))
		for i, argfun := range argfuns {
			argv[i] = argfun(env2)
		}
		// the call is executed in a new goroutine
		go funv.Call(argv)

		env.IP++
		return env.Code[env.IP], env
	})

	// propagate back the compiled code
	c.Code = c2.Code
}

// If compiles an "if" statement
func (c *Comp) If(node *ast.IfStmt) {
	var jump struct{ Then, Else, End int }

	initLocals := false
	var initBinds [2]int
	if node.Init != nil {
		c, initLocals = c.pushEnvIfLocalBinds(&initBinds, node.Init)
		c.Stmt(node.Init)
	}
	pred := c.Expr(node.Cond)
	flag, fun, err := pred.TryAsPred()
	if err {
		c.invalidPred(node.Cond, pred)
		return
	}
	if fun != nil {
		c.Code.Append(func(env *Env) (Stmt, *Env) {
			var ip int
			if fun(env) {
				ip = jump.Then
			} else {
				ip = jump.Else
			}
			env.IP = ip
			return env.Code[ip], env
		})
	}
	// compile 'then' branch
	jump.Then = c.Code.Len()
	c.Block(node.Body)
	if fun == nil && !flag {
		// 'then' branch is never executed...
		// still compiled above (to check for errors) but drop the generated code
		c.Code.List = c.Code.List[0:jump.Then]
	}
	// compile a 'goto' between 'then' and 'else' branches
	if fun != nil && node.Else != nil {
		c.Code.Append(func(env *Env) (Stmt, *Env) {
			// after executing 'then' branch, we must skip 'else' branch
			env.IP = jump.End
			return env.Code[jump.End], env
		})
	}
	// compile 'else' branch
	jump.Else = c.Code.Len()
	if node.Else != nil {
		// parser should guarantee Else to be a block or another "if"
		// but macroexpansion can optimize away the block if it contains no declarations.
		// still, better be safe and wrap the Else again in a block because:
		// 1) catches improper macroexpander optimizations
		// 2) there is no runtime performance penalty
		xelse := node.Else
		_, ok1 := xelse.(*ast.BlockStmt)
		_, ok2 := xelse.(*ast.IfStmt)
		if ok1 || ok2 {
			c.Stmt(xelse)
		} else {
			c.Block(&ast.BlockStmt{List: []ast.Stmt{xelse}})
		}
		if fun == nil && flag {
			// 'else' branch is never executed...
			// still compiled above (to check for errors) but drop the generated code
			c.Code.List = c.Code.List[0:jump.Else]
		}
	}
	jump.End = c.Code.Len()

	if node.Init != nil {
		c = c.popEnvIfLocalBinds(initLocals, &initBinds, node.Init)
	}
}

// IncDec compiles a "place++" or "place--" statement
func (c *Comp) IncDec(node *ast.IncDecStmt) {
	place := c.Place(node.X)
	op := node.Tok
	if op == token.DEC {
		op = token.SUB
	} else {
		op = token.ADD
	}
	one := exprValue(untypedOne)
	if place.IsVar() {
		c.SetVar(&place.Var, op, one)
	} else {
		c.SetPlace(place, op, one)
	}
}

// Return compiles a "return" statement
func (c *Comp) Return(node *ast.ReturnStmt) {
	var cinfo *FuncInfo
	var upn int
	var cf *Comp
	for cf = c; cf != nil; cf = cf.Outer {
		if cf.Func != nil {
			cinfo = cf.Func
			break
		}
		upn += cf.UpCost // count how many Env:s we must exit at runtime
	}
	if cinfo == nil {
		c.Errorf("return outside function")
		return
	}

	resultBinds := cinfo.Results
	resultExprs := node.Results
	n := len(resultBinds)
	switch len(resultExprs) {
	case n:
		// ok
	case 1:
		c.Errorf("unimplemented: return of multi-valued expression: %v", node)
	case 0:
		if !cinfo.NamedResults {
			// naked return requires results to have names
			c.Errorf("return: expecting %d expressions, found %d: %v", n, len(resultExprs), node)
			return
		}
		n = 0 // naked return. results are already set
	default:
		c.Errorf("return: expecting %d expressions, found %d: %v", n, len(resultExprs), node)
		return
	}

	exprs := c.Exprs(resultExprs)
	for i := 0; i < n; i++ {
		c.SetVar(resultBinds[i].AsVar(upn, PlaceSettable), token.ASSIGN, exprs[i])
	}
	stmt := func(env *Env) (Stmt, *Env) {
		common := env.ThreadGlobals
		common.Signal = SigReturn
		return common.Interrupt, env
	}
	c.Code.Append(stmt)
}

// containLocalBinds return true if one or more of the given statements (but not their contents:
// blocks are not examined) contain some function/variable declaration.
// ignores types, constants and anything named "_"
func containLocalBinds(list ...ast.Stmt) bool {
	if len(list) == 0 {
		Errorf("internal error: containLocalBinds() invoked on empty statement list")
	}
	for i, node := range list {
		switch node := node.(type) {
		case *ast.AssignStmt:
			if node.Tok == token.DEFINE {
				return true
			}
		case *ast.DeclStmt:
			switch decl := node.Decl.(type) {
			case *ast.FuncDecl:
				// Go compiler forbids local functions... we allow them
				if decl.Name != nil && decl.Name.Name != "_" {
					return true
				}
			case *ast.GenDecl:
				if decl.Tok != token.VAR {
					continue
				}
				// found local variables... bail out unless they are all named "_"
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.ValueSpec:
						for _, ident := range spec.Names {
							if ident.Name != "_" {
								return true
							}
						}
					}
				}
			}
		case nil:
			Errorf("internal error: containLocalBinds() statement[%d] is nil: %v", i, list)
		}
	}
	return false
}

// pushEnvIfLocalBinds compiles a PushEnv statement if list contains local binds
// returns the *Comp to use to compile statement list.
func (c *Comp) pushEnvIfLocalBinds(nbinds *[2]int, list ...ast.Stmt) (inner *Comp, locals bool) {
	if len(list) == 0 {
		inner.Errorf("internal error: pushEnvIfLocalBinds() invoked on empty statement list")
	}
	// optimization: examine statements. if none of them is a function/variable declaration,
	// no need to create a new *Env at runtime
	// note: we still create a new *Comp at compile time to handle constant/type declarations
	locals = containLocalBinds(list...)
	if locals {
		// push new *Env at runtime. we will know # of binds in the block only later, so use a closure on them
		c.Code.Append(func(env *Env) (Stmt, *Env) {
			inner := NewEnv(env, nbinds[0], nbinds[1])
			inner.IP++
			// Debugf("PushEnv, IP = %d of %d, pushed %d binds and %d intbinds", inner.IP, nbinds[0], nbinds[1])
			return inner.Code[inner.IP], inner
		})
	}
	inner = NewComp(c)
	if !locals {
		inner.UpCost = 0
		inner.Depth--
	}
	return inner, locals
}

// popEnvIfLocalBinds compiles a PopEnv statement if locals is true. also sets *nbinds and *nintbinds
func (inner *Comp) popEnvIfLocalBinds(locals bool, nbinds *[2]int, list ...ast.Stmt) *Comp {
	if len(list) == 0 {
		inner.Errorf("internal error: popEnvIfLocalBinds() invoked on empty statement list")
	}
	c := inner.Outer
	c.Code = inner.Code       // copy back accumulated code
	nbinds[0] = inner.BindNum // we finally know these
	nbinds[1] = inner.IntBindNum

	if locals != (inner.BindNum != 0 || inner.IntBindNum != 0) {
		c.Errorf(`internal error: containLocalBinds() returned %t, but block actually defined %d Binds and %d IntBinds:
	Binds = %v
	Block =
%v`, locals, inner.BindNum, inner.IntBindNum, inner.Binds, &ast.BlockStmt{List: list})
		return nil
	}

	if locals {
		// pop *Env at runtime
		c.Code.Append(popEnv)
	}
	return c
}
