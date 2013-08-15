// This is kind of annoying... It doesn't seem to be easy to fake the OS interaction here. Grrr...
package types

import (
    "testing"
    . "github.com/tychofreeman/go-matchers"
)

func TestLoadsDependenciesInFile(t *testing.T) {
    n := CreateParser("test-source/system/")
    pkgs,_ := n.Parse("test-source/local-src/")
    f := pkgs["a"].Files["test-source/local-src/a.go"]
    FillTypes(f, n.pkgs)

    types := getTypesForIds(f, "c")
    
    AssertThat(t, types, HasExactly(AliasedType{"", IntType(), map[string]FunctionType{}}))
}
