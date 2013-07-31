package typing

import (
    "fmt"
    "reflect"
    "go/ast"
    "go/token"
)

type Type struct {
    name string
}

func StringType() Type {
    return Type{"string"}
}

func IntType() Type {
    return Type{"int"}
}

func FloatType() Type {
    return Type{"float"}
}

func CharType() Type {
    return Type{"char"}
}

func ComplexType() Type {
    return Type{"complex"}
}

func FuncType() Type {
    return Type{"func"}
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
                return Type{"TYPE IDENT " + t.Name}
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
        return Type{"UNKNOWN!!!"}
    case *ast.Field:
        switch tt := t.Type.(type) {
        case *ast.Ident:
            return getTypes(tt)
        default:
            fmt.Printf("Unhandled Field: %v\n", reflect.TypeOf(tt))
        }
        return Type{"???"}
    default:
        fmt.Printf("Unhandled Node: %v\n", reflect.TypeOf(n))
    }
    fmt.Printf("Returning 'UNKNOWN'\n")
    return Type{"UNKNOWN"}
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
                            t.Obj.Type = Type{"int"}
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
    }
    return v
}

func fillTypes(n ast.Node) {
    v := TypeFillingVisitor{}
    ast.Walk(v, n) 
}

