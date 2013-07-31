package typing

import (
    "fmt"
    "reflect"
    "testing"
    . "github.com/tychofreeman/go-matchers"
    "go/ast"
    "go/parser"
    "go/token"
)

func getTypeOfExpr(expr string) Type {
    i, _ := parser.ParseExpr(expr)
    return getTypes(i)
}

func TestFindsTypeOfStringExpression(t *testing.T) {
    exprType := getTypeOfExpr("\"my str\"")
    AssertThat(t, exprType, Equals(StringType()))
}

func TestFindsTypeOfIntExpression(t *testing.T) {
    exprType := getTypeOfExpr("1")
    AssertThat(t, exprType, Equals(IntType()))
}

func TestFindTypeOfFloatExpression(t *testing.T) {
    exprType := getTypeOfExpr("1.0")
    AssertThat(t, exprType, Equals(FloatType()))
}

func TestFindTypeOfCharExpression(t *testing.T) {
    exprType := getTypeOfExpr("'a'")
    AssertThat(t, exprType, Equals(CharType()))
}

func TestFindTypeOfImaginaryExpression(t *testing.T) {
    exprType := getTypeOfExpr("2.4i")
    AssertThat(t, exprType, Equals(ComplexType()))
}

func TestFindsTypeOfAdditionExpr(t *testing.T) {
    exprType := getTypeOfExpr("1 + 2")
    AssertThat(t, exprType, Equals(IntType()))
}

func TestFillsTypeOfExpression(t *testing.T) {
    i, _ := parser.ParseExpr("func (a int) int { return a }")
    switch i2 := i.(type) {
    case *ast.FuncLit:
        rtn := i2.Body.List[0]
        switch rtn2 := rtn.(type) {
        case *ast.ReturnStmt:
            switch i3 := rtn2.Results[0].(type) {
            case *ast.Ident:
                AssertThat(t, i3.Obj.Type, Equals(nil))
                fillTypes(i)
                AssertThat(t, i3.Obj.Type, Equals(Type{"int"}))
            default:
                t.Error("Expected to find an identifier at Expr.FuncLit.Body.List[0].ReturnStmt.Results[0]")
            }
        default:
            t.Error("Expected to find a return statment at Expr.FuncLit.Body.List[0]")
        }
    default:
        t.Error()
    }
}

type FuncVisitor struct {
    f func(n ast.Node)
}
func (fv FuncVisitor) Visit(n ast.Node) ast.Visitor {
    if n != nil {
        fv.f(n)
    }
    return fv
}

func ParseFile(contents string) *ast.File {
    f, err := parser.ParseFile(token.NewFileSet(), "tmp.go", contents, parser.ParseComments)
    if err != nil {
        panic(err)
    }
    return f
}

func TestFillsTypeOfVarInAssignment(t *testing.T) {
    f := ParseFile("package main\nfunc (a int) int { b := a + 2; return b }")
    visitor := FuncVisitor{
        func(n ast.Node) {
            switch i2 := n.(type) {
            case *ast.FuncLit:
        fmt.Printf("Node [%v]\n", reflect.TypeOf(n))
                t.Log("Found Function Literal...\n")
                rtn := i2.Body.List[0]
                switch rtn2 := rtn.(type) {
                case *ast.AssignStmt:
                    switch i3 := rtn2.Lhs[0].(type) {
                    case *ast.Ident:
                        AssertThat(t, i3.Obj.Type, Equals(nil))
                        fillTypes(n)
                        AssertThat(t, i3.Obj.Type, Equals(IntType()))
                    default:
                        t.Error("Expected to find an identifier at Expr.FuncLit.Body.List[0].ReturnStmt.Results[0]")
                    }
                default:
                    t.Error("Expected to find a return statment at Expr.FuncLit.Body.List[0]")
                }
            }
        },
    }
    ast.Walk(visitor, f)
}

func TestFillsTypeOfVarAssignedToCast(t *testing.T) {
    f := ParseFile("func (a int) float { b := float(a); return b }")
    v := FuncVisitor{
        func(n ast.Node) { 
            switch assign := n.(type) {
            case *ast.AssignStmt:
                switch lhs := assign.Lhs[0].(type) {
                    case *ast.Ident:
                        AssertThat(t, lhs.Obj.Type, Equals(nil))
                        fillTypes(n)
                        AssertThat(t, lhs.Obj.Type, Equals(FloatType()))
                }
            }
        },
    }
    ast.Walk(v, f)
}

func TestFillsTypeOfFuncName(t *testing.T) {
    f := ParseFile("func f(a int) float { return float(a) }")
    fillTypes(f)
    v := FuncVisitor{
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                if ident.Name == "a" {
                    if ident.Obj == nil {
                        t.Errorf("indent.Obj should not be nil")
                    } else if ident.Obj.Type == nil {
                        t.Errorf("indent.Obj.Type should not be nil")
                    }
                }
            }
        },
    }
    ast.Walk(v, f)
}

func TestFillsVarAssignedToCallOfFunc(t *testing.T) {
    f := ParseFile("func f(a int) int { return a * 2 }\nfunc g() {b := f(3)}")
    fillTypes(f)
    bWasFound := false
    v := FuncVisitor {
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                if ident.Name == "b" {
                    bWasFound = true
                    AssertThat(t, ident.Obj.Type, Equals(IntType()))
                }
            }
        },
    }
    ast.Walk(v, f)
    AssertThat(t, bWasFound, IsTrue)
}
