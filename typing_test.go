package typing

import (
    "testing"
    . "github.com/tychofreeman/go-matchers"
    "go/ast"
    "go/parser"
    "go/token"
    "reflect"
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
    t *testing.T
}
func (fv FuncVisitor) Visit(n ast.Node) ast.Visitor {
    if n == nil {
        return fv
    }
    fv.t.Log("walking ast.Node=", reflect.TypeOf(n))
    switch i2 := n.(type) {
    case *ast.FuncLit:
        fv.t.Log("Found Function Literal...\n")
        rtn := i2.Body.List[0]
        switch rtn2 := rtn.(type) {
        case *ast.AssignStmt:
            switch i3 := rtn2.Lhs[0].(type) {
            case *ast.Ident:
                AssertThat(fv.t, i3.Obj.Type, Equals(nil))
                fillTypes(n)
                AssertThat(fv.t, i3.Obj.Type, Equals(Type{"int"}))
            default:
                fv.t.Error("Expected to find an identifier at Expr.FuncLit.Body.List[0].ReturnStmt.Results[0]")
            }
        default:
            fv.t.Error("Expected to find a return statment at Expr.FuncLit.Body.List[0]")
        }
    }
    return fv
}

// Need to figure out exactly what we're trying to do here...
// Well EFFF - ParseExpr doesn't seem to populate variable info. I need 'a' to have Ident.Obj data populated, but it's not. It might be an oversight: solitary expressions don't usually define variables.
func TestFillsTypeOfVarInAssignment(t *testing.T) {
    f, _ := parser.ParseFile(token.NewFileSet(), "tmp.go", "func (a int) int { b := a + 2; return b }", parser.ParseComments)
    visitor := FuncVisitor{t}
    ast.Walk(visitor, f)
}

