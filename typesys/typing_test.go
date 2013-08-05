package typing

import (
    "testing"
    . "github.com/tychofreeman/go-matchers"
    "go/ast"
    "go/parser"
    "go/token"
)

func getTypeOfExpr(expr string) Type {
    i, _ := parser.ParseExpr(expr)
    return TypeFillingVisitor{nil}.getTypes(i)
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
                fillTypes(i, nil)
                AssertThat(t, i3.Obj.Type, Equals(IntType()))
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

func ParseFile(filename, contents string) *ast.File {
    f, err := parser.ParseFile(token.NewFileSet(), filename, contents, parser.ParseComments)
    if err != nil {
        panic(err)
    }
    return f
}

func getTypesForIds(f *ast.File, ids ...string) (interface{}) {
    types := []interface{}{}
    v := FuncVisitor{
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                for _, id := range ids {
                    if ident.Name == id {
                        types = append(types, ident.Obj.Type)
                    }
                }
            }
        },
    }
    ast.Walk(v,f)
    return types
}

func TestFillsTypeOfVarInAssignment(t *testing.T) {
    f := ParseFile("TestFillsTypeOfVarInAssignment", "package main\nvar z func(int)int = func (a int) int { b := a + 2; return b }")
    visitor := FuncVisitor{
        func(n ast.Node) {
            switch i2 := n.(type) {
            case *ast.FuncLit:
                t.Log("Found Function Literal...\n")
                rtn := i2.Body.List[0]
                switch rtn2 := rtn.(type) {
                case *ast.AssignStmt:
                    switch i3 := rtn2.Lhs[0].(type) {
                    case *ast.Ident:
                        AssertThat(t, i3.Obj.Type, Equals(nil))
                        fillTypes(n, nil)
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

func _TestFillsTypeOfVarAssignedToCast(t *testing.T) {
    f := ParseFile("TestFillsTypeOfVarAssignedToCast", "package main\nfunc t(a int) float { b := float(a); return b }")
    v := FuncVisitor{
        func(n ast.Node) { 
            switch assign := n.(type) {
            case *ast.AssignStmt:
                switch lhs := assign.Lhs[0].(type) {
                    case *ast.Ident:
                        AssertThat(t, lhs.Obj.Type, Equals(nil))
                        fillTypes(n, nil)
                        AssertThat(t, lhs.Obj.Type, Equals(FloatType()))
                }
            }
        },
    }
    ast.Walk(v, f)
}

func TestFillsTypeOfFuncName(t *testing.T) {
    f := ParseFile("TestFillsTypeOfFuncName", "package main\nfunc f(a int) float { return float(a) }")
    fillTypes(f, nil)
    v := FuncVisitor{
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                if ident.Name == "f" {
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
    f := ParseFile("TestFillsVarAssignedToCallOfFunc", "package main\nfunc f(a int) int { return a * 2 }\nfunc g() {b := f(3)}")
    fillTypes(f, nil)
    types := getTypesForIds(f, "b")
    AssertThat(t, types, HasExactly(IntType()))
}

func TestFillsVarWithTypeAlias(t *testing.T) {
    f := ParseFile("TestFillsvarWithTypeAlias", "package main\ntype A int\nfunc f(a A) { b := A(a) }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "a")
    AssertThat(t, types, HasExactly(AliasType(IntType()),AliasType(IntType())))
}

func TestFillsMultipleParamsInFuncDef(t *testing.T) {
    f := ParseFile("TestFillsMultipleParamsInFuncDef", "package main\nfunc f(a, b int, c float) { }")
    fillTypes(f, nil)
    var types []interface{}
    v := FuncVisitor{
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                if ident.Name == "a" || ident.Name == "b" || ident.Name == "c" {
                    types = append(types, ident.Obj.Type)
                }
            }
        },
    }
    ast.Walk(v, f)
    AssertThat(t, types, HasExactly(IntType(),IntType(),FloatType()))
}

func TestFillsAllReturnValuesInFuncDef(t *testing.T) {
    f := ParseFile("TestFillsAllReturnValuesInFuncDef", "package main\nfunc f() (a, b int, c float) { return 0,0,1.0 }")
    fillTypes(f, nil)
    var types []interface{}
    v := FuncVisitor{
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                if ident.Name == "a" || ident.Name == "b" || ident.Name == "c" {
                    types = append(types, ident.Obj.Type)
                }
            }
        },
    }
    ast.Walk(v, f)
    AssertThat(t, types, HasExactly(IntType(),IntType(),FloatType()))
}

func TestFillsAllReturnAndParamTypesInFuncLiteral(t *testing.T) {
    f := ParseFile("TestFillsAllReturnAndParamTypesInFuncLiteral", "package main\nfunc init() { fn := func(a,b int, c float) (d,e string, f rune) { return } }")
    fillTypes(f, nil)
    var types []interface{}
    v := FuncVisitor{
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                switch ident.Name {
                case "a", "b", "c", "d", "e", "f":
                    types = append(types, ident.Obj.Type)
                }
            }
        },
    }
    ast.Walk(v, f)
    AssertThat(t, types, HasExactly(IntType(),IntType(),FloatType(),StringType(),StringType(),RuneType()))
}

func TestFindsCorrectTypeForParamsInFunctionIdent(t *testing.T) {
    f := ParseFile("TestFillsAllReturnAndParamTypesInFuncLiteral", "package main\nfunc fn(a,b int, c float) (d,e string, f rune) { return }")
    fillTypes(f, nil)
    types := []FunctionType{}
    v := FuncVisitor{
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                switch ident.Name {
                case "fn":
                    switch funcType := ident.Obj.Type.(type) {
                    case FunctionType:
                        types = append(types, funcType)
                    }
                }
            }
        },
    }
    ast.Walk(v, f)

    // Copy to interface{} slices...
    params := []interface{}{}
    for _, p := range types[0].params {
        params = append(params, p)
    }
    returns := []interface{}{}
    for _, p := range types[0].returns {
        returns = append(returns, p)
    }

    AssertThat(t, types[0].receiver, Equals(nil))
    AssertThat(t, params, HasExactly(IntType(),IntType(),FloatType()))
    AssertThat(t, returns,HasExactly(StringType(),StringType(),RuneType()))
}

func TestFillsIdentForStructType(t *testing.T) {
    f := ParseFile("TestFillsTypeOfFieldWithinStruct", "package main\ntype A struct {\nB int\n}\n func f(a A) { c := a.B }")
    fillTypes(f, nil)
    types := []interface{}{}
    v := FuncVisitor{
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                switch ident.Name {
                case "a":
                    types = append(types, ident.Obj.Type)
                }
            }
        },
    }

    ast.Walk(v,f)
    aType := AliasType(StructType(map[string]Type{"B":IntType()}, nil))
    AssertThat(t, types[0], Equals(aType))
}

func TestFillsAllIdentsForStructType(t *testing.T) {
    f := ParseFile("TestFillsAllIdentsForStructType", "package main\ntype A struct {\nB,C int\nD float\n}\n func f(a A) { c := a.B }")
    fillTypes(f, nil)
    types := []interface{}{}
    v := FuncVisitor{
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                switch ident.Name {
                case "a":
                    types = append(types, ident.Obj.Type)
                }
            }
        },
    }

    ast.Walk(v,f)
    aType := AliasType(StructType(map[string]Type{"B":IntType(), "C":IntType(), "D":FloatType()}, nil))
    AssertThat(t, types[0], Equals(aType))
}

func TestFillsAllAnonFieldsForStructType(t *testing.T) {
    f := ParseFile("TestFillsAllIdentsForStructType", "package main\ntype A struct {\nB,C int\nD float\n}\ntype E struct {\nA\n}\n func f(e E) { c := e.B }")
    fillTypes(f, nil)
    types := []interface{}{}
    v := FuncVisitor{
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                switch ident.Name {
                case "e":
                    types = append(types, ident.Obj.Type)
                }
            }
        },
    }

    ast.Walk(v,f)
    aType := AliasType(StructType(map[string]Type{"B":IntType(), "C":IntType(), "D":FloatType()}, nil))
    eType := AliasType(StructType(map[string]Type{}, []Type{aType}))
    AssertThat(t, types[0], Equals(eType))
    AssertThat(t, types[1], Equals(eType))
}

func TestFillsTypeOfFieldWithinStruct(t *testing.T) {
    f := ParseFile("TestFillsTypeOfFieldWithinStruct", "package main\ntype A struct {\nB int\n}\n func f(a A) { c := a.B }")
    fillTypes(f, nil)
    types := []interface{}{}
    v := FuncVisitor{
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                switch ident.Name {
                case "c":
                    types = append(types, ident.Obj.Type)
                }
            }
        },
    }

    ast.Walk(v,f)
    AssertThat(t, types, HasExactly(IntType()))
}

func TestFillsTypeOfFieldOfAnonymousSubfieldWithinStruct(t *testing.T) {
    f := ParseFile("TestFillsTypeOfFieldOfAnonymousSubfieldWithinStruct", "package main\ntype A struct {\nB int\n}\ntype D struct {\nA\n}\n func f(a D) { c := a.B }")
    fillTypes(f, nil)
    types := []interface{}{}
    v := FuncVisitor{
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                switch ident.Name {
                case "c":
                    types = append(types, ident.Obj.Type)
                }
            }
        },
    }

    ast.Walk(v,f)
    AssertThat(t, types, HasExactly(IntType()))
}

// Next, add selection of methods...
func TestFillsTypeForMethod(t *testing.T) {
    f := ParseFile("TestFillsTypeForMethod", "package main\ntype A int\nfunc (a A) Inc() A { return A(a + 1) }\nfunc f(a A) { b := a.Inc }")
    fillTypes(f, nil)
    types := []interface{}{}
    v := FuncVisitor{
        func(n ast.Node) {
            switch ident := n.(type) {
            case *ast.Ident:
                switch ident.Name {
                case "b":
                    types = append(types, ident.Obj.Type)
                }
            }
        },
    }

    ast.Walk(v,f)
    intAlias := AliasedType{"", IntType(), map[string]FunctionType{}}
    var aliasPtr Type = intAlias
    expected := FunctionType{&aliasPtr, []Type{}, []Type{intAlias}}
    intAlias.AddMethod("Inc", expected)
    AssertThat(t, types, HasExactly(expected))
}

func TestFillsTypeForMethodInAnotherPackage(t *testing.T) {
    f := ParseFile("TestFillsTypeForMethod", "package main\nimport \"mypkg\"\nfunc init() { b := mypkg.ReturnsInt() }")
    pkg := map[string]PackageType{"mypkg":PackageType{map[string]FunctionType{"ReturnsInt":FunctionType{nil,[]Type{},[]Type{IntType()}}}}}
    fillTypes(f, pkg)
    types := getTypesForIds(f, "b")

    AssertThat(t, types, HasExactly(IntType()))
}

// Need to handle package aliases, and the corner case of ".". That'll be tricky.
// Also, currently I can't have a package name and a variable name conflict. Is that ok?
