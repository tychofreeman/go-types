package typing

import (
    "fmt"
    "reflect"
    "go/ast"
    "go/token"
)

type Type interface {
    String() string
}

type NoneType struct{}

func (n NoneType) String() string {
    return "<UNKOWN TYPE>"
}

type SimpleType struct {
    name string
}

func (s SimpleType) String() string {
    return s.name
}

type FunctionType struct {
    receiver *Type
    params []Type
    returns []Type
}

func (f FunctionType) String() string {
    return "func"
}
    

type AliasedType struct {
    name string
    wrapped Type
}

func (a AliasedType) String() string {
    return "type " + a.name + " " + a.wrapped.String()   
}

func UnknownType(data string) Type {
    return NoneType{}
}

func StringType() Type {
    return SimpleType{"string"}
}

func RuneType() Type {
    return SimpleType{"rune"}
}

func IntType() Type {
    return SimpleType{"int"}
}

func FloatType() Type {
    return SimpleType{"float"}
}

func CharType() Type {
    return SimpleType{"char"}
}

func ComplexType() Type {
    return SimpleType{"complex"}
}

func AliasType(aliased Type) Type {
    return AliasedType{"", aliased}
}

func FuncType(recv *Type, paramList, results []Type) Type {
    return FunctionType{receiver:recv, params:paramList, returns:results}
}

// Create a FunctionType from the information found in the AST.
func funcType(recv *ast.FieldList, params, results []*ast.Field) Type {
    paramList := []Type{}
    for _, param := range params {
        for _, paramIdent := range param.Names {
            paramList = append(paramList, getTypes(paramIdent))
        }
    }

    rtnList := []Type{}
    for _, result := range results {
        if len(result.Names) == 0 {
            rtnList = append(rtnList, getTypes(result.Type))
        } else {
            for _, resultIdent := range result.Names {
                rtnList = append(rtnList, getTypes(resultIdent))
            }
        }
    }

    return FunctionType{receiver:nil, params:paramList, returns:rtnList}
}

func getTypes(n ast.Node) Type {
    switch t := n.(type) {
    case *ast.BasicLit:
        switch t.Kind {
        case token.STRING:
            return StringType()
        case token.INT:
            return IntType()
        case token.FLOAT:
            return FloatType()
        case token.CHAR:
            return CharType()
        case token.IMAG:
            return ComplexType()
        default:
            fmt.Printf("Unhandled BasicLit: %v\n", reflect.TypeOf(t.Kind))
        }
    case *ast.BinaryExpr:
        xType := getTypes(t.X)
        yType := getTypes(t.Y)
        if xType == yType {
            return xType
        } else {
            fmt.Printf("Unhandled BinaryExpr: %v vs %v\n", xType, yType)
        }
    case *ast.CallExpr:
        fnType := getTypes(t.Fun)
        switch fnType := fnType.(type) {
        case FunctionType:
            if len(fnType.returns) > 0 {
                return fnType.returns[0]
            } else {
                return NoneType{}
            }
        }
        return UnknownType("Wierd Call expression")
    case *ast.FuncDecl:
        return funcType(nil, nil, t.Type.Results.List)
    case *ast.FuncType:
        results := []*ast.Field{}
        params := []*ast.Field{}
        if t.Results != nil {
            results = t.Results.List
        }
        if t.Params != nil {
            params = t.Params.List
        }
        return funcType(nil, params, results)
    case *ast.Ident:
        if t.Obj == nil {
            switch t.Name {
            case "int":
                return IntType()
            case "string":
                return StringType()
            case "float":
                return FloatType()
            case "char":
                return CharType()
            case "rune":
                return RuneType()
            default:
                return UnknownType("TYPE IDENT " + t.Name)
            }
        }
        if t.Obj.Type != nil {
            switch it := t.Obj.Type.(type) {
            case Type:
                return it
            default:
                fmt.Printf("Unhandled Ident.Obj.Type: %v\n", reflect.TypeOf(t.Obj.Type))
            }
        }
        switch de := t.Obj.Decl.(type) {
        case ast.Node:
            return getTypes(de)
        default:
            fmt.Printf("Unhandled Ident.Obj.Decl: %v (data: %v)\n", reflect.TypeOf(t.Obj.Decl), reflect.TypeOf(t.Obj.Data))
        }
        return UnknownType("UNKNOWN!!!")
    case *ast.Field:
        switch tt := t.Type.(type) {
        case *ast.Ident:
            return getTypes(tt)
        default:
            fmt.Printf("Unhandled Field: %v\n", reflect.TypeOf(tt))
        }
        return UnknownType("???")
    case *ast.TypeSpec:
        //return AliasType(t.Name.Name, getTypes(t.Type))
        return AliasType(getTypes(t.Type))
    default:
        fmt.Printf("Unhandled Node: %v\n", reflect.TypeOf(n))
    }
    fmt.Printf("Returning 'UNKNOWN'\n")
    return UnknownType("UNKNOWN")
}

type TypeFillingVisitor struct {
}

func fillFieldList(fs *ast.FieldList) {
    if fs != nil {
        for _, l := range fs.List {
            for _, name := range l.Names {
                name.Obj.Type = getTypes(l.Type)
            }
        }
    }
}

// We'll have to hide the first 'case *ast.Ident' within a Return case.
// Then we'll have to create a new 'case *ast.AssignStmt' to handle the new test
func (v TypeFillingVisitor) Visit(n ast.Node) ast.Visitor {
    switch r := n.(type) {
    case *ast.ReturnStmt:
        for _, result := range r.Results {
            switch t := result.(type) {
            case *ast.Ident:
                if t.Obj.Kind == ast.Var {
                    switch f := t.Obj.Decl.(type) {
                    case *ast.Field:
                        switch i := f.Type.(type) {
                        case *ast.Ident:
                            if i.Name == "int" {
                                t.Obj.Type = IntType()
                            }
                        }
                    }
                }
            }
        }
    case *ast.AssignStmt:
        switch t := r.Lhs[0].(type) {
        case *ast.Ident:
            if t.Obj.Kind == ast.Var {
                t.Obj.Type = getTypes(r.Rhs[0])
            }
        }
    case *ast.FuncDecl:
        fillFieldList(r.Type.Params)
        fillFieldList(r.Type.Results)
        if r.Name != nil && r.Name.Obj != nil {
            r.Name.Obj.Type = getTypes(r.Type)
        }
    case *ast.FuncLit:
        fillFieldList(r.Type.Params)
        fillFieldList(r.Type.Results)
    default:
    }
    return v
}

func fillTypes(n ast.Node) {
    v := TypeFillingVisitor{}
    ast.Walk(v, n) 
}

