package types

import (
    "fmt"
    "reflect"
    "go/ast"
    "go/token"
    "strings"
    "strconv"
)

type Type interface {
    String() string
    Equals(interface{}) (bool,string)
}

type MapType struct {
    key, value Type
}

func (m MapType) String() string {
    return "map[" + m.key.String() + "]" + m.value.String()
}

func (m MapType) Equals(other interface{}) (bool,string) {
    switch other := other.(type) {
    case MapType:
        keyEq, keyMsg := m.key.Equals(other.key)
        valueEq, valueMsg := m.value.Equals(other.value)
        msg := "KEYS: " + keyMsg + " AND VALUES: " + valueMsg
        return (keyEq && valueEq), msg
    }
    return false, fmt.Sprintf("MapType(*,*) != %T\n", other)
}

type PointerType struct {
    wrapped Type
}

func (p PointerType) String() string {
    return "*" + p.wrapped.String()
}

func (p PointerType) Equals(other interface{}) (bool,string) {
    switch other := other.(type) {
    case PointerType:
        eq, msg := p.wrapped.Equals(other.wrapped)
        return eq, fmt.Sprintf("*%s != *%s (%s)", p.wrapped, other.wrapped, msg)
    }
    return false, fmt.Sprintf("PointerType(*) != %T", other)
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

const (
    Slice = iota
    UnknownLen
    KnownLen
)

type Len struct {
    lenType int
    len int
}

func (l Len) String() string {
    switch l.lenType {
    case Slice:
        return ""
    case UnknownLen:
        return "..."
    case KnownLen:
        return fmt.Sprintf("%d", l.len)
    default:
        panic(fmt.Sprintf("malformed Len: %v\n", l))
    }
}

type SliceType struct {
    len Len
    subtype Type
}

func (s SliceType)String() string {
    return fmt.Sprintf("[%v]%v", s.len, s.subtype)
}

func (s SliceType) Equals(other interface{}) (bool, string) {
    switch other := other.(type) {
    case SliceType:
        if s.len.lenType != other.len.lenType {
            return false, fmt.Sprintf("Indexability %v != %v", s.len.lenType, other.len.lenType)
        }
        if s.len.lenType == KnownLen && s.len.len != other.len.len {
            return false, fmt.Sprintf("Array length %v != %v\n", s.len.len, other.len.len)
        }
        return s.subtype.Equals(other.subtype)
    }
    return false, fmt.Sprintf("SliceType(*) != %T", other)
}

func MakeSlice(sub Type) Type {
    return SliceType{Len{Slice,0}, sub}
}

func MakeArray(len int, sub Type) Type {
    return SliceType{Len{KnownLen, len}, sub}
}

func MakeUnlenArray(sub Type) Type {
    return SliceType{Len{UnknownLen, 0}, sub}
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
    v.starIsDeref = false
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


    v.starIsDeref = true
    return FunctionType{receiver:recvType, params:paramList, returns:rtnList}
}

func (v TypeFillingVisitor) getTypes(n ast.Node) Type {
    switch t := n.(type) {
    case *ast.ArrayType:
        subtype := v.getTypes(t.Elt)
        switch len := t.Len.(type) {
        case *ast.BasicLit:
            switch len.Kind {
            case token.INT:
                l, _ := strconv.Atoi(len.Value)
                return MakeArray(l, subtype)
            }
        case *ast.Ellipsis:
            return MakeUnlenArray(subtype)
        case nil:
            return MakeSlice(subtype)
        }
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
            panic(fmt.Sprintf("Unhandled BinaryExpr: %v vs %v\n", xType, yType))
        }
    case *ast.CallExpr:
        fnType := v.getTypes(t.Fun)
        switch fnType := fnType.(type) {
        case FunctionType:
            if len(fnType.returns) > 0 {
                return fnType.returns[0]
            } else {
                return NoneType{}
            }
        case Type:
            return fnType
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
        fmt.Printf("FUNC TYPE!!!\n")
        results := []*ast.Field{}
        params := []*ast.Field{}
        if t.Results != nil {
            results = t.Results.List
        }
        if t.Params != nil {
            params = t.Params.List
        }
        t2 := v.funcType(nil, params, results)
        fmt.Printf("// END FUN TYPE!!!\n")
        return t2
    case *ast.Ident:
        if t.Obj == nil {
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
        case *ast.StarExpr:
            return PointerType{v.getTypes(tt.X)}
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
        fmt.Printf("Could not find types for %s.%s\n", t.X, t.Sel.Name)
    case *ast.IndexExpr:
        switch t := v.getTypes(t.X).(type) {
        case SliceType:
            return t.subtype
        case MapType:
            return t.value
        }
    case *ast.UnaryExpr:
        switch t.Op {
        case token.AND:
            return PointerType{v.getTypes(t.X)}
        case token.ADD, token.SUB, token.XOR:
            return v.getTypes(t.X)
        }
    case *ast.StarExpr:
        if v.starIsDeref {
            switch t := v.getTypes(t.X).(type) {
            case PointerType:
                return t.wrapped
            default:
                fmt.Printf("Star Expr: Type %T - %v\n", t, t)
            }
        } else {
            return PointerType{v.getTypes(t.X)}
        }
    case *ast.MapType:
        return MapType{v.getTypes(t.Key), v.getTypes(t.Value)}
    default:
        fmt.Printf("Unhandled Node: %v\n", reflect.TypeOf(n))
    }
    fmt.Printf("Returning 'UNKNOWN' for type %T\n", n)
    return UnknownType("UNKNOWN")
}

type TypeFillingVisitor struct {
    pkg map[string]PackageType
    aliases map[string]string
    starIsDeref bool
    builtPkg PackageType
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
            rType := v.getTypes(r)
            r.Name.Obj.Type = rType
            v.builtPkg.types[r.Name.Name] = rType
        }
    case *ast.ImportSpec:
        if v.aliases != nil && r.Name != nil {
            v.aliases[r.Name.Name] = strings.Trim(r.Path.Value, "\"")
        }
    case *ast.FuncDecl:
        v.starIsDeref = false
        v.fillFieldList(r.Type.Params)
        v.fillFieldList(r.Type.Results)
        v.fillFieldList(r.Recv)
        v.starIsDeref = true
        var fnType FunctionType
        switch t := v.getTypes(r).(type) {
        case FunctionType:
            fnType = t
            v.builtPkg.types[r.Name.Name] = fnType
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
        v.starIsDeref = false
        v.fillFieldList(r.Type.Params)
        v.fillFieldList(r.Type.Results)
        v.starIsDeref = true
    default:
    }
    return v
}

var builtIns PackageType = PackageType{
    map[string]Type{
        //"slice" - ?
        "int":IntType(),
        "int8":SimpleType{"int8"},
        "int16":SimpleType{"int16"},
        "int32":SimpleType{"int32"},
        "int64":SimpleType{"int64"},
        "complex64":SimpleType{"complex64"},
        "complex128":SimpleType{"complex128"},
        //"real" and "imag" destructuring functions, and "complex" construction function
        //"recover" and "panic"
        //"make" and "new"
        //"delete" and "close"
        //"true" and "false"
        "float":FloatType(),
        "float32":SimpleType{"float32"},
        "float64":SimpleType{"float64"},
        "bool":SimpleType{"bool"},
        "byte":SimpleType{"byte"},
        //"cap", "copy" and "len":FunctionType{},
        "string":StringType(),
        "uint":SimpleType{"uint"},
        "uint8":SimpleType{"uint8"},
        "uint16":SimpleType{"uint16"},
        "uint32":SimpleType{"uint32"},
        "uint64":SimpleType{"uint64"},
        "uintptr":SimpleType{"uintptr"},
        "rune":RuneType(),
    },
}

func union(a,b map[string]PackageType) map[string]PackageType {
    out := map[string]PackageType{}
    f := func(a map[string]PackageType) {
        for k,v := range a {
            out[k] = v
        }
    }
    f(a)
    f(b)
    return out
}

func MakeTFVisitor(pkg map[string]PackageType) TypeFillingVisitor {
    return TypeFillingVisitor{pkg:union(pkg, map[string]PackageType{"builtins":builtIns}), aliases:map[string]string{".":"builtins"},starIsDeref:true,builtPkg:PackageType{types:map[string]Type{}}}
}

func FillTypes(n ast.Node, pkg map[string]PackageType) PackageType {
    v := MakeTFVisitor(pkg)
    ast.Walk(v, n) 
    return v.builtPkg
}
