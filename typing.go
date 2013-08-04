package typing

import (
    "fmt"
    "reflect"
    "go/ast"
    "go/token"
)

type Type interface {
    String() string
    Equals(interface{}) (bool,string)
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

func (s SimpleType) Equals(other interface{}) (bool,string) {
    switch other := other.(type) {
    case SimpleType:
        return s.name == other.name, fmt.Sprintf("%s != %s", s, other)
    default:
        return false, fmt.Sprintf("SimpleType(*) != %T", other)
    }
}

type FunctionType struct {
    receiver *Type
    params []Type
    returns []Type
}

func (f FunctionType) String() string {
    return "func"
}

func (f FunctionType) Equals(other interface{}) (bool,string) {
    switch other := other.(type) {
    case FunctionType:
        match := true
        match = (f.receiver == other.receiver)
        for i, _ := range f.params {
            if eq, _ := f.params[i].Equals(other.params[i]); !eq{
                match = false
            }
        }
        for i, _ := range f.returns {
            if eq, _ := f.returns[i].Equals(other.returns[i]); !eq {
                match = false
            }
        }
        return match, fmt.Sprintf("%s != %s", f, other)
    default:
        return false, fmt.Sprintf("FunctionType != %T", other)
    }
}
    

type AliasedType struct {
    name string
    wrapped Type
}

func (a AliasedType) String() string {
    return "type " + a.name + " " + a.wrapped.String()   
}

func (a AliasedType) Equals(other interface{}) (bool,string) {
    switch other := other.(type) {
    case AliasedType:
        eq, _ := a.wrapped.Equals(other.wrapped)
        return a.name == other.name && eq, fmt.Sprintf("%s != %s", a, other)
    default:
        return false, fmt.Sprintf("AliasedType != %T", other)
    }
}

func UnknownType(data string) Type {
    return NoneType{}
}

func (n NoneType) Equals(other interface{}) (bool,string) {
    return false, fmt.Sprintf("NoneType != %T", other)
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

type StructureType struct {
    internals map[string]Type
    anons []Type
}

func (s StructureType)String() string {
    str := "struct { "
    for _, anon := range s.anons {
        str += " " + anon.String() + ";"
    }
    for name, internal := range s.internals {
        str += " " + name + " " + internal.String() + ";"
    }
    str += "}"
    return str
}

func (s StructureType) Equals(other interface{}) (bool, string) {
    switch other := other.(type) {
    case StructureType:
        for k,_ := range s.internals {
            if v, ok := other.internals[k]; !ok {
                return false, fmt.Sprintf("missing internal type named %s with type %s", k, v)
            }
        }
        for k,_ := range other.internals {
            if v, ok := s.internals[k]; !ok {
                return false, fmt.Sprintf("missing internal type named %s with type %s", k, v)
            }
        }
        return true, ""
    default:
        return false, fmt.Sprintf("Structure Type != %T", other)
    }
}

func StructType(fields map[string]Type, anons []Type) Type {
    return StructureType{fields, anons}
}

func getStructInternals(fields *ast.FieldList) (map[string]Type, []Type) {
    internals := map[string]Type{}
    anons := []Type{}
    if fields != nil {
        for _, f := range fields.List {
            if len(f.Names) > 0 {
                for _, name := range f.Names {
                    internals[name.Name] = getTypes(f.Type)
                }
            } else {
                anons = append(anons, getTypes(f.Type))
            }
        }
    }
    return internals, anons
}

func getSelectedType(baseType Type, name string) (Type,bool) {
    switch baseType := baseType.(type) {
    case AliasedType:
        return getSelectedType(baseType.wrapped, name)
    case StructureType:
        selected, ok := baseType.internals[name]
        if ok {
            return selected, ok
        }
        for _, anon := range baseType.anons {
            if sub, ok := getSelectedType(anon, name); ok {
                return sub, ok
            }
        }
        return NoneType{}, false
    default:
        return NoneType{}, false
    }
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
    case *ast.StructType:
        return StructType(getStructInternals(t.Fields))
    case *ast.SelectorExpr:
        if rtn, ok := getSelectedType(getTypes(t.X), t.Sel.Name); ok {
            return rtn
        }
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

