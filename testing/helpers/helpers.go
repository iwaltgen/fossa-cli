package helpers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fossas/fossa-cli/module"
	"github.com/fossas/fossa-cli/pkg"
)

// AssertDosFile parses the file and ensures that every line ending is formatted
// for DOS operating systems with a carriage return and line feed ("\r\n").
func AssertDosFile(t *testing.T, file []byte) {
	fixture := string(file)
	for i := range fixture {
		if i == 0 {
			continue
		}
		if fixture[i] == '\n' {
			assert.Equal(t, uint8('\r'), fixture[i-1])
		}
	}
}

// AssertUnixFile parses the file and ensures that every line ending is formatted
// for Unix/Linux operating systems with only line feed ("\n").
func AssertUnixFile(t *testing.T, file []byte) {
	fixture := string(file)
	for i := range fixture {
		if i == 0 {
			continue
		}
		if fixture[i] == '\n' {
			assert.NotEqual(t, uint8('\r'), fixture[i-1])
		}
	}
}

// PackageInTransitiveGraph searches a map (typically from Graph.Deps.Transitive)
// for a package and returns it if it exists.
func PackageInTransitiveGraph(packages map[pkg.ID]pkg.Package, name, revision string) pkg.Package {
	for id := range packages {
		if id.Name == name && id.Revision == revision {
			return packages[id]
		}
	}
	return pkg.Package{}
}

// AssertPackageImport searches a list of imports (typically from pkg.Package.Imports)
// for a package and asserts on its existence.
func AssertPackageImport(t *testing.T, imports pkg.Imports, name, revision string) {
	for _, importedProj := range imports {
		if importedProj.Resolved.Name == name {
			if importedProj.Resolved.Revision == revision {
				return
			}
			assert.Fail(t, "found "+name+"@"+importedProj.Resolved.Revision+" instead of "+revision)
		}
	}
	assert.Fail(t, "missing "+name+"@"+revision)
}

// AssertModuleExists searches a list of modules and determines if the specified module exists.
// This is most often used for testing Discover methods.
func AssertModuleExists(t *testing.T, modules []module.Module, modType pkg.Type, name, directory, target string) {
	for _, module := range modules {
		if module.Type == modType {
			if module.Name == name {
				if module.BuildTarget == target {
					if module.Dir == directory {
						return
					}
				}
			}
		}
	}
	assert.Fail(t, fmt.Sprintf("missing module `name: %s, type: %s, dir: %s, target: %s, in modules: %v", name, modType, directory, target, modules))
}
