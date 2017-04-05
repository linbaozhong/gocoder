package gocoder

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type GoFile struct {
	filename string
	gopath   string

	goFuncs []*GoFunc

	importPackages []*GoPackage

	options *Options

	*GoExpr
}

func NewGoFile(filename string, options ...Option) (goFile *GoFile, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return
	}

	gf := &GoFile{
		filename: filename,
		GoExpr: &GoExpr{
			astFile:    f,
			astFileSet: fset,
		},
		options: &Options{},
	}

	for i := 0; i < len(options); i++ {
		if err = options[i](gf.options); err != nil {
			return
		}
	}

	if len(gf.options.GoPath) == 0 {
		gf.options.GoPath = os.Getenv("GOPATH")
	}

	if err = gf.load(); err != nil {
		return
	}

	if err = gf.loadImportPackages(); err != nil {
		return
	}

	goFile = gf

	return
}

func (p *GoFile) Print() error {
	return ast.Print(p.astFileSet, p.astFile)
}

func (p *GoFile) Funcs() []*GoFunc {
	return p.goFuncs
}

func (p *GoFile) Package() *GoPackage {
	return p.options.GoPackage
}

func (p *GoFile) Filename() string {
	return p.filename
}

func (p *GoFile) ShortFilename() string {
	return strings.TrimPrefix(p.filename, p.options.GoPath+"/src/")
}

func (p *GoFile) Imports() []*GoPackage {
	return p.importPackages
}

func (p *GoFile) FindImport(importPath string) (*GoPackage, bool) {
	for i := 0; i < len(p.importPackages); i++ {
		if importPath == p.importPackages[i].Path() {
			return p.importPackages[i], true
		}
	}
	return nil, false
}

func (p *GoFile) load() (err error) {
	if err = p.loadDecls(); err != nil {
		return
	}

	return
}

func (p *GoFile) loadImportPackages() (err error) {
	for _, impt := range p.astFile.Imports {
		var pkg *GoPackage
		imptPath := strings.Trim(impt.Path.Value, "\"")

		if p.options.IgnoreSystemPackages && !strings.HasPrefix(imptPath, p.options.GoPath) {
			continue
		}

		pkg, err = NewGoPackage(imptPath, OptionImportBy(p.Package()))
		if err != nil {
			return
		}
		p.importPackages = append(p.importPackages, pkg)
	}
	return nil
}

func (p *GoFile) loadDecls() error {
	for _, decl := range p.astFile.Decls {
		ast.Inspect(decl, func(n ast.Node) bool {

			switch d := n.(type) {
			case *ast.FuncDecl:
				{
					p.parseDeclFunc(d)
				}
			}

			return true
		})
	}
	return nil
}

func (p *GoFile) parseDeclFunc(decl *ast.FuncDecl) {
	p.goFuncs = append(p.goFuncs, newGoFunc(p.GoExpr, decl))
}