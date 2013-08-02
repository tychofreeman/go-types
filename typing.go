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

func FuncType(recv *ast.FieldList, params, results []*ast.Field) Type {
    
    paramList := []Type{}
    rtnList := []Type{}
    var firstRtnType Type = nil
    if len(results) == 0 {
        return NoneType{}
    }
    if len(results[0].Names) > 0 {
        firstRtnType = getTypes(results[0].Names[0])
    }
    switch firstRtnType.(type) {
    case NoneType:
        rtnList = append(rtnList, getTypes(results[0].Type))
    case nil:
        rtnList = append(rtnList, getTypes(results[0].Type))
    default:
        rtnList = append(rtnList, firstRtnType)
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
        return FuncType(nil, nil, t.Type.Results.List)
    case *ast.FuncType:
        if t.Results != nil {
            return FuncType(nil, nil, t.Results.List)
        } else {
            return FuncType(nil, nil, []*ast.Field{})
        }
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
        return AliasedType{t.Name.Name, getTypes(t.Type)}
    default:
        fmt.Printf("Unhandled Node: %v\n", reflect.TypeOf(n))
    }
    fmt.Printf("Returning 'UNKNOWN'\n")
    return UnknownType("UNKNOWN")
}

type TypeFillingVisitor struct {
}

// We'll have to hide the first 'case *ast.Ident' within a Return case.
// Then we'll have to create a new 'case *ast.AssignStmt' to handle the new test
func (v TypeFillingVisitor) Visit(n ast.Node) ast.Visitor {
    switch r := n.(type) {
    case *ast.ReturnStmt:
        switch t := r.Results[0].(type) {
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
    case *ast.AssignStmt:
        switch t := r.Lhs[0].(type) {
        case *ast.Ident:
            if t.Obj.Kind == ast.Var {
                t.Obj.Type = getTypes(r.Rhs[0])
            }
        }
    case *ast.FuncDecl:
        r.Name.Obj.Type = getTypes(r.Type)
        // Add handling of parameter and results here...
    default:
    }
    return v
}

func fillTypes(n ast.Node) {
    v := TypeFillingVisitor{}
    ast.Walk(v, n) 
}

