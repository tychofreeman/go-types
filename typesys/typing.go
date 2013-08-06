package typing

import (
    "fmt"
    "reflect"
    "go/ast"
    "go/token"
    "strings"
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

type PackageType struct {
    types map[string]Type
}

func (p PackageType) String() string {
    return "pkg"
}

func (p PackageType) Equals(other interface{}) (bool,string) {
    match := true
    msg := ""
    switch other := other.(type) {
    case PackageType:
        for k,f := range p.types {
            if otherF, ok := other.types[k]; !ok {
                match = false
                msg = fmt.Sprintf("Function %s is not found", k)
            } else if ok2, _ := f.Equals(otherF); !ok2 {
                match = false
                msg = fmt.Sprintf("Function %s does not match: %v vs %v", k, f, otherF)
            }
        }
    default:
        match = false
        msg = fmt.Sprintf("PakageType != %T", other)
    }
    return match, msg
}

type FunctionType struct {
    receiver *Type
    params []Type
    returns []Type
}

func (f FunctionType) String() string {
    if f.receiver != nil  {
        return fmt.Sprintf("func (%v) (%v) (%v)", *f.receiver, f.params, f.returns)
    }
    return fmt.Sprintf("func  (%v) (%v)", f.params, f.returns)
}

func (f FunctionType) Equals(other interface{}) (bool,string) {
    switch other := other.(type) {
    case FunctionType:
        match := true
        if f.receiver != nil && other.receiver != nil {
            match, _ = ((*f.receiver).Equals(*other.receiver))
        } else {
            match = (f.receiver == other.receiver)
        }
        if len(f.params) != len(other.params) {
            match = false
        } else {
            for i, _ := range f.params {
                if eq, _ := f.params[i].Equals(other.params[i]); !eq{
                    match = false
                }
            }
        }
        if len(f.returns) != len(other.returns) {
            match = false
        } else {
            for i, _ := range f.returns {
                if eq, _ := f.returns[i].Equals(other.returns[i]); !eq {
                    match = false
                }
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
    methods map[string]FunctionType
}

func (a AliasedType) AddMethod(name string, t FunctionType) {
    a.methods[name] = t
}

func (a AliasedType) String() string {
    out := "type " + a.name + " " + a.wrapped.String() + "\n"
    for k, _ := range a.methods {
        out += "\t func " + k + "\n"
    }
    return out
}

func (a AliasedType) Equals(other interface{}) (bool,string) {
    switch other := other.(type) {
    case *AliasedType:
        return a.Equals(*other)
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
    return &AliasedType{"", aliased, map[string]FunctionType{}}
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

func (v TypeFillingVisitor) getStructInternals(fields *ast.FieldList) (map[string]Type, []Type) {
    internals := map[string]Type{}
    anons := []Type{}
    if fields != nil {
        for _, f := range fields.List {
            if len(f.Names) > 0 {
                for _, name := range f.Names {
                    internals[name.Name] = v.getTypes(f.Type)
                }
            } else {
                anons = append(anons, v.getTypes(f.Type))
            }
        }
    }
    return internals, anons
}

func (v TypeFillingVisitor) getSelectedType(baseType Type, name string) (Type,bool) {
    switch baseType := baseType.(type) {
    case *AliasedType:
        if m, ok := baseType.methods[name]; ok {
            return m, ok
        }
        return v.getSelectedType(baseType.wrapped, name)
    case StructureType:
        selected, ok := baseType.internals[name]
        if ok {
            return selected, ok
        }
        for _, anon := range baseType.anons {
            if sub, ok := v.getSelectedType(anon, name); ok {
                return sub, ok
            }
        }
        return NoneType{}, false
    case PackageType:
        item, ok := baseType.types[name]
        return item, ok
    default:
        fmt.Printf("could not get selected type of %v - %T\n", baseType, baseType)
        return NoneType{}, false
    }
}

func FuncType(recv Type, paramList, results []Type) Type {
    recvPtr := &recv
    switch recv.(type) {
    case NoneType:
        recvPtr = nil
    }
    return FunctionType{receiver:recvPtr, params:paramList, returns:results}
}


// Create a FunctionType from the information found in the AST.
func (v TypeFillingVisitor) funcType(recv *ast.FieldList, params, results []*ast.Field) Type {
    paramList := []Type{}
    for _, param := range params {
        for _, paramIdent := range param.Names {
            paramList = append(paramList, v.getTypes(paramIdent))
        }
    }

    rtnList := []Type{}
    for _, result := range results {
        if len(result.Names) == 0 {
            rtnList = append(rtnList, v.getTypes(result.Type))
        } else {
            for _, resultIdent := range result.Names {
                rtnList = append(rtnList, v.getTypes(resultIdent))
            }
        }
    }

    var recvType *Type
    if recv != nil && len(recv.List) > 0 {
        rc := v.getTypes(recv.List[0].Names[0])
        recvType = &rc
    }


    return FunctionType{receiver:recvType, params:paramList, returns:rtnList}
}

func (v TypeFillingVisitor) getTypes(n ast.Node) Type {
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
        xType := v.getTypes(t.X)
        yType := v.getTypes(t.Y)
        if xType == yType {
            return xType
        } else {
            fmt.Printf("Unhandled BinaryExpr: %v vs %v\n", xType, yType)
        }
    case *ast.CallExpr:
        switch fnName := t.Fun.(type) {
        case *ast.Ident:
            lits := map[string]bool{"int":true, "float":true,"char":true,"complex":true,"rune":true}
            if _, ok := lits[fnName.Name]; ok {
                return SimpleType{fnName.Name}
            }
        }
        fnType := v.getTypes(t.Fun)
        fmt.Printf("CallExpr: %T - %T\n", t.Fun, fnType)
        switch fnType := fnType.(type) {
        case FunctionType:
            if len(fnType.returns) > 0 {
                return fnType.returns[0]
            } else {
                return NoneType{}
            }
        }
        return UnknownType("Wierd Call expression")
    case *ast.CompositeLit:
        return v.getTypes(t.Type)
    case *ast.FuncDecl:
        results := []*ast.Field{}
        params := []*ast.Field{}
        if t.Type.Results != nil {
            results = t.Type.Results.List
        }
        if t.Type.Params != nil {
            params = t.Type.Params.List
        }
        ft := v.funcType(t.Recv, params, results)
        return ft
    case *ast.FuncType:
        results := []*ast.Field{}
        params := []*ast.Field{}
        if t.Results != nil {
            results = t.Results.List
        }
        if t.Params != nil {
            params = t.Params.List
        }
        return v.funcType(nil, params, results)
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
                pkgName := t.Name
                if alias, ok := v.aliases[pkgName]; ok {
                    pkgName = alias
                }
                if pkg, ok := v.pkg[pkgName]; ok {
                    return pkg
                }
                for alias, pkgName := range v.aliases {
                    if alias == "." {
                        pkg := v.pkg[pkgName].types
                        if identType, ok := pkg[t.Name]; ok {
                            return identType
                        }
                    }
                }
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
            return v.getTypes(de)
        default:
            fmt.Printf("Unhandled Ident.Obj.Decl: %v (data: %v)\n", reflect.TypeOf(t.Obj.Decl), reflect.TypeOf(t.Obj.Data))
        }
        return UnknownType("UNKNOWN!!!")
    case *ast.Field:
        switch tt := t.Type.(type) {
        case *ast.Ident:
            return v.getTypes(tt)
        default:
            fmt.Printf("Unhandled Field: %v\n", reflect.TypeOf(tt))
        }
        return UnknownType("???")
    case *ast.TypeSpec:
        //return AliasType(t.Name.Name, v.getTypes(t.Type))
        return AliasType(v.getTypes(t.Type))
    case *ast.StructType:
        return StructType(v.getStructInternals(t.Fields))
    case *ast.SelectorExpr:
        if rtn, ok := v.getSelectedType(v.getTypes(t.X), t.Sel.Name); ok {
            return rtn
        }
    default:
        fmt.Printf("Unhandled Node: %v\n", reflect.TypeOf(n))
    }
    fmt.Printf("Returning 'UNKNOWN' for type %T\n", n)
    return UnknownType("UNKNOWN")
}

type TypeFillingVisitor struct {
    pkg map[string]PackageType
    aliases map[string]string
}

func (v TypeFillingVisitor) fillFieldList(fs *ast.FieldList) {
    if fs != nil {
        for _, l := range fs.List {
            for _, name := range l.Names {
                name.Obj.Type = v.getTypes(l.Type)
            }
        }
    }
}

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
                t.Obj.Type = v.getTypes(r.Rhs[0])
            }
        }
    case *ast.TypeSpec:
        if r.Name != nil {
            r.Name.Obj.Type = v.getTypes(r)
        }
    case *ast.ImportSpec:
        if v.aliases != nil && r.Name != nil {
            v.aliases[r.Name.Name] = strings.Trim(r.Path.Value, "\"")
        }
    case *ast.FuncDecl:
        v.fillFieldList(r.Type.Params)
        v.fillFieldList(r.Type.Results)
        var fnType FunctionType
        switch t := v.getTypes(r).(type) {
        case FunctionType:
            fnType = t
        }
        if r.Name != nil && r.Name.Obj != nil {
            r.Name.Obj.Type = fnType
        }
        if r.Recv != nil {
            switch recvDecl := r.Recv.List[0].Type.(type) {
            case *ast.Ident:
                switch recv := recvDecl.Obj.Type.(type) {
                case *AliasedType:
                    recv.AddMethod(r.Name.Name, fnType)
                }
            }
        }
    case *ast.FuncLit:
        v.fillFieldList(r.Type.Params)
        v.fillFieldList(r.Type.Results)
    default:
    }
    return v
}

func fillTypes(n ast.Node, pkg map[string]PackageType) {
    v := TypeFillingVisitor{pkg:pkg, aliases:map[string]string{}}
    ast.Walk(v, n) 
}

