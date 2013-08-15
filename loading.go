package types

import (
    "strings"
    "go/ast"
    "go/parser"
    "go/token"
)

type PkgVisitor struct {
    pkgs []string
}

func (p *PkgVisitor) Visit(n ast.Node) ast.Visitor {
    switch n := n.(type) {
    case *ast.GenDecl:
        for _, spec := range n.Specs {
            switch spec := spec.(type) {
                case *ast.ImportSpec:
                    path := strings.Trim(spec.Path.Value, "\"")
                    p.pkgs = append(p.pkgs, path)
            }
        }
    }
    return p
}

type Parser struct {
    dirName string
    pkgs map[string]PackageType
}

func CreateParser(dirName string) *Parser {
    return &Parser{dirName:dirName, pkgs:map[string]PackageType{}}
}

func (p *Parser)Parse(fname string) (map[string]*ast.Package,error) {
    fset := token.NewFileSet()
    stuff := []string{fname}
    stuff2 := map[string]*ast.Package{}
    
    for _, src := range stuff {
        tmpPkgs, _ := parser.ParseDir(fset, src, nil, parser.ImportsOnly)
        v := PkgVisitor{[]string{}}
        for _, pkg := range tmpPkgs {
            ast.Walk(&v, pkg)
        }
        for _, pkgName := range v.pkgs {
            if _, ok := p.pkgs[pkgName]; !ok {
                p.Parse(p.dirName + "/" + pkgName)
            }
        }

        pkgs, _ := parser.ParseDir(fset, src, nil, parser.ParseComments)
        for pkgName, pkg := range pkgs {
            stuff2[pkgName] = pkg
            for _, file := range pkg.Files {
                constructed := FillTypes(file, p.pkgs)
                p.pkgs[pkgName] = constructed
            }
        }
    }
    
    return stuff2, nil
}
