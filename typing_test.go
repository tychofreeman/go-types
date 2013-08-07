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
    return TypeFillingVisitor{nil,nil,true,PackageType{}}.getTypes(i)
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
    types := getTypesForIds(f, "a","b","c")
    AssertThat(t, types, HasExactly(IntType(),IntType(),FloatType()))
}

func TestFillsAllReturnValuesInFuncDef(t *testing.T) {
    f := ParseFile("TestFillsAllReturnValuesInFuncDef", "package main\nfunc f() (a, b int, c float) { return 0,0,1.0 }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "a","b","c")
    AssertThat(t, types, HasExactly(IntType(),IntType(),FloatType()))
}

func TestFillsAllReturnAndParamTypesInFuncLiteral(t *testing.T) {
    f := ParseFile("TestFillsAllReturnAndParamTypesInFuncLiteral", "package main\nfunc init() { fn := func(a,b int, c float) (d,e string, f rune) { return } }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "a", "b", "c", "d", "e", "f")
    AssertThat(t, types, HasExactly(IntType(),IntType(),FloatType(),StringType(),StringType(),RuneType()))
}

func TestFindsCorrectTypeForParamsInFunctionIdent(t *testing.T) {
    f := ParseFile("TestFillsAllReturnAndParamTypesInFuncLiteral", "package main\nfunc fn(a,b int, c float) (d,e string, f rune) { return }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "fn")

    AssertThat(t, types, HasExactly(FunctionType{nil, []Type{IntType(),IntType(),FloatType()}, []Type{StringType(),StringType(),RuneType()}}))
}

func TestFillsIdentForStructType(t *testing.T) {
    f := ParseFile("TestFillsTypeOfFieldWithinStruct", "package main\ntype A struct {\nB int\n}\n func f(a A) { c := a.B }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "a")

    aType := AliasType(StructType(map[string]Type{"B":IntType()}, nil))
    AssertThat(t, types, HasExactly(aType, aType))
}

func TestFillsAllIdentsForStructType(t *testing.T) {
    f := ParseFile("TestFillsAllIdentsForStructType", "package main\ntype A struct {\nB,C int\nD float\n}\n func f(a A) { c := a.B }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "a")

    aType := AliasType(StructType(map[string]Type{"B":IntType(), "C":IntType(), "D":FloatType()}, nil))
    AssertThat(t, types, HasExactly(aType, aType))
}

func TestFillsAllAnonFieldsForStructType(t *testing.T) {
    f := ParseFile("TestFillsAllIdentsForStructType", "package main\ntype A struct {\nB,C int\nD float\n}\ntype E struct {\nA\n}\n func f(e E) { c := e.B }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "e")

    aType := AliasType(StructType(map[string]Type{"B":IntType(), "C":IntType(), "D":FloatType()}, nil))
    eType := AliasType(StructType(map[string]Type{}, []Type{aType}))
    AssertThat(t, types, HasExactly(eType, eType))
}

func TestFillsTypeOfFieldWithinStruct(t *testing.T) {
    f := ParseFile("TestFillsTypeOfFieldWithinStruct", "package main\ntype A struct {\nB int\n}\n func f(a A) { c := a.B }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "c")

    AssertThat(t, types, HasExactly(IntType()))
}

func TestFillsTypeOfFieldOfAnonymousSubfieldWithinStruct(t *testing.T) {
    f := ParseFile("TestFillsTypeOfFieldOfAnonymousSubfieldWithinStruct", "package main\ntype A struct {\nB int\n}\ntype D struct {\nA\n}\n func f(a D) { c := a.B }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "c")

    AssertThat(t, types, HasExactly(IntType()))
}

// Next, add selection of methods...
func TestFillsTypeForMethod(t *testing.T) {
    f := ParseFile("TestFillsTypeForMethod", "package main\ntype A int\nfunc (a A) Inc() A { return A(a + 1) }\nfunc f(a A) { b := a.Inc }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "b")

    intAlias := AliasedType{"", IntType(), map[string]FunctionType{}}
    var aliasPtr Type = intAlias
    expected := FunctionType{&aliasPtr, []Type{}, []Type{intAlias}}
    intAlias.AddMethod("Inc", expected)
    AssertThat(t, types, HasExactly(expected))
}

func TestFillsTypeForMethodInAnotherPackage(t *testing.T) {
    f := ParseFile("TestFillsTypeForMethod", "package main\nimport \"mypkg\"\nfunc init() { b := mypkg.ReturnsInt() }")
    pkg := map[string]PackageType{"mypkg":PackageType{map[string]Type{"ReturnsInt":FunctionType{nil,[]Type{},[]Type{IntType()}}}}}
    fillTypes(f, pkg)
    types := getTypesForIds(f, "b")

    AssertThat(t, types, HasExactly(IntType()))
}

func TestFillsTypeForMethodInAliasedPackage(t *testing.T) {
    f := ParseFile("TestFillsTypeForMethodInAliasedPackage", "package main\nimport mp \"mypkg\"\nfunc init() { b := mp.ReturnsInt() }")
    pkg := map[string]PackageType{"mypkg":PackageType{map[string]Type{"ReturnsInt":FunctionType{nil,[]Type{},[]Type{IntType()}}}}}
    fillTypes(f, pkg)
    types := getTypesForIds(f, "b")

    AssertThat(t, types, HasExactly(IntType()))
}

func TestFillsTypeForTypeInDotAliasedPackage(t *testing.T) {
    f := ParseFile("TestFillsTypeForMethodInAliasedPackage", "package main\nimport . \"mypkg\"\nfunc init() { b := ExtTypeA }")
    aType := AliasedType{"ExtTypeA",IntType(),map[string]FunctionType{}}
    pkg := map[string]PackageType{"mypkg":PackageType{map[string]Type{"ExtTypeA":aType}}}
    fillTypes(f, pkg)
    types := getTypesForIds(f, "b")

    AssertThat(t, types, HasExactly(aType))
}

func TestFillsTypeForCompositeLitWithAndWithoutFields(t *testing.T) {
    // Interestingly, I don't think we care about the values of the fields within the CompositeLit.
    f := ParseFile("TestFillsTypeForMethodInAliasedPackage", "package main\nimport . \"mypkg\"\nfunc init() { b := ExtTypeA{}; c := ExtTypeA{0} }")
    aType := AliasedType{"ExtTypeA",StructType(map[string]Type{"A":IntType()}, []Type{}), map[string]FunctionType{}}
    pkg := map[string]PackageType{"mypkg":PackageType{map[string]Type{"ExtTypeA":aType}}}
    fillTypes(f, pkg)
    types := getTypesForIds(f, "b", "c")

    AssertThat(t, types, HasExactly(aType, aType))
}

func TestFillsTypeForTypeConversion(t *testing.T) {
    f := ParseFile("TestFillsTypeForMethodInAliasedPackage", "package main\nfunc init() { a := int(1.0) }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "a")

    AssertThat(t, types, HasExactly(IntType()))
}

func TestFillsTypeForArrayIndex(t *testing.T) {
    f := ParseFile("TestFillsTypeForArrayIndex", "package main\nfunc f(ints []int, floats [2]float, runes [...]rune) { a := ints[0]; b := floats; c := runes[0] }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "a", "b", "c")

    AssertThat(t, types, HasExactly(IntType(),MakeArray(2,FloatType()), RuneType()))
}

func TestFillsTypeForPointer(t *testing.T) {
    f := ParseFile("TestFillsTypeForPointer", "package main\nfunc (a *int) f(i *int) { b := *i; c := i; d := &i }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "a", "b", "c", "d")

    AssertThat(t, types, HasExactly(PointerType{IntType()}, IntType(), PointerType{IntType()}, PointerType{PointerType{IntType()}}))
}

func TestFillsTypeForMap(t *testing.T) {
    f := ParseFile("TestFillsTypeForMap", "package main\nfunc f(m map[string]float) { a := m[\"hi\"] }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "a")

    AssertThat(t, types, HasExactly(FloatType()))
}

func TestFillsTypeForNumbersWithUnaryOps(t *testing.T) {
    f := ParseFile("TestFillsTypeForMap", "package main\nfunc init() { a := -100; b:= +100; c := ^1 }")
    fillTypes(f, nil)
    types := getTypesForIds(f, "a", "b", "c")

    AssertThat(t, types, HasExactly(IntType(),IntType(),IntType()))
}

func TestReturnsPackageWithExportedSymbols(t *testing.T) {
    f := ParseFile("TestReturnsPackageWithExportedSymbols", "package main\ntype A int\nfunc B() {}")
    pkg := fillTypes(f, nil)
   
    AssertThat(t, pkg.types["A"], Equals(AliasedType{"",IntType(),map[string]FunctionType{}}))
    AssertThat(t, pkg.types["B"], Equals(FuncType(NoneType{},[]Type{},[]Type{})))
}

// Need to handle:
//  channels
//  func literals?

// Also, currently I can't have a package name and a variable name conflict. Is that ok?
