package typing

import (
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
        }
    case *ast.BinaryExpr:
        xType := getTypes(t.X)
        yType := getTypes(t.Y)
        if xType == yType {
            return xType
        }
    }
    return Type{"UNKNOWN"}
}

type TypeFillingVisitor struct {
}

// We'll have to hide the first 'case *ast.Ident' within a Return case.
// Then we'll have to create a new 'case *ast.AssignStmt' to handle the new test
func (v TypeFillingVisitor) Visit(n ast.Node) ast.Visitor {
    switch t := n.(type) {
    // UGH!!
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
    return v
}

func fillTypes(n ast.Node) {
    v := TypeFillingVisitor{}
    ast.Walk(v, n) 
}

