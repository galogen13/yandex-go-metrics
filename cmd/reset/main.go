package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/tools/go/packages"
)

const resetGenerateComment = "// generate:reset"

const templateStr = `
// Code generated automaticaly by reset utility; DO NOT EDIT.

package {{.Name}}

{{range .StructsInfo}}
func (v *{{.Name}}) Reset() {
	{{range .Fields}}
    	{{.ResetCode}}
	{{end}}
} 
{{end}}
`

type FieldInfo struct {
	Name      string
	ResetCode string
}

type StructInfo struct {
	Name          string
	Package       string
	OutputPath    string
	HasUsersReset bool // Флаг, есть ли уже пользовательский метод Reset
	Fields        []FieldInfo
}

type packageInfo struct {
	Name        string
	StructsInfo []StructInfo
}

func main() {

	if err := logger.Initialize("debug"); err != nil {
		log.Fatalf("Cannot Initialize logger: %v\n", err)
	}
	defer logger.Log.Sync()

	rootDir, err := findProjectRoot()
	if err != nil {
		log.Fatalf("Cannot found project root directory: %v\n", err)
	}

	// Сканируем все пакеты и находим структуры с комментарием // generate:reset
	structs, err := scanPackages(rootDir)
	if err != nil {
		log.Fatalf("Cannot scan packages: %v\n", err)
	}

	if len(structs) == 0 {
		logger.Log.Info("No structs with comment", zap.String("comment", resetGenerateComment))
		return
	}

	structsByPackage := groupByPackage(structs)

	for pkgPath, pkgStructs := range structsByPackage {
		if err := generateResetFile(pkgPath, pkgStructs); err != nil {
			logger.Log.Error("Cannot generate resetter for package", zap.String("package path", pkgPath), zap.Error(err))
		}
	}

	logger.Log.Info("Generation completed")
}

// findProjectRoot ищет корневую директорию проекта (где находится go.mod)
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// scanPackages сканирует все пакеты и находит структуры с комментарием // generate:reset
func scanPackages(rootDir string) ([]StructInfo, error) {
	var structs []StructInfo

	cfg := &packages.Config{Dir: rootDir, Mode: packages.LoadFiles}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("error load packages: %w", err)
	}

	for _, pkg := range pkgs {

		for _, goFileName := range pkg.GoFiles {

			if !strings.HasSuffix(goFileName, ".go") ||
				strings.HasSuffix(goFileName, "_test.go") ||
				strings.HasSuffix(goFileName, ".gen.go") ||
				strings.HasSuffix(goFileName, ".mock.go") {
				continue
			}

			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, goFileName, nil, parser.ParseComments)
			if err != nil {
				logger.Log.Error("parse file error", zap.String("fileName", goFileName), zap.Error(err))
				continue
			}

			ast.Inspect(node, func(n ast.Node) bool {
				genDecl, ok := n.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					return true
				}

				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if s, isStruct := typeSpec.Type.(*ast.StructType); isStruct {
							if hasGenerateResetComment(genDecl) {
								fields := []FieldInfo{}
								for _, field := range s.Fields.List {
									if len(field.Names) == 0 {
										continue
									}
									resetCode, err := getResetCode(field.Names[0].Name, field.Type)
									if err != nil {
										logger.Log.Info("no reset code for field", zap.String("field", field.Names[0].Name), zap.Error(err))
										continue
									}
									fields = append(fields,
										FieldInfo{
											Name:      field.Names[0].Name,
											ResetCode: resetCode,
										})
								}

								structInfo := StructInfo{
									Name:       typeSpec.Name.Name,
									Package:    pkg.Name,
									OutputPath: pkg.Dir,
									Fields:     fields,
								}
								structInfo.HasUsersReset = hasResetMethod(node, typeSpec.Name.Name)

								structs = append(structs, structInfo)
							}
						}
					}
				}
				return true
			})

		}

	}

	return structs, nil
}

// hasGenerateResetComment проверяет наличие комментария // generate:reset над структурой
func hasGenerateResetComment(genDecl *ast.GenDecl) bool {
	if genDecl.Doc == nil {
		return false
	}

	for _, comment := range genDecl.Doc.List {
		if strings.Contains(comment.Text, resetGenerateComment) {
			return true
		}
	}
	return false
}

// hasResetMethod проверяет, существует ли уже метод Reset для структуры
func hasResetMethod(file *ast.File, structName string) bool {
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if funcDecl.Name.Name != "Reset" {
			continue
		}

		if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			continue
		}

		recvType := funcDecl.Recv.List[0].Type

		var typeName string
		switch t := recvType.(type) {
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				typeName = ident.Name
			}
		case *ast.Ident:
			typeName = t.Name
		}

		if typeName == structName {
			return true
		}
	}
	return false
}

// groupByPackage группирует структуры по директории вывода
func groupByPackage(structs []StructInfo) map[string][]StructInfo {
	result := make(map[string][]StructInfo)
	for _, s := range structs {
		result[s.OutputPath] = append(result[s.OutputPath], s)
	}
	return result
}

var tmpl = template.Must(template.New("reset").Parse(templateStr))

// generateResetFile создает файл reset.gen.go для пакета
func generateResetFile(pkgPath string, structs []StructInfo) error {
	outputFile := filepath.Join(pkgPath, "reset.gen.go")

	packageName := structs[0].Package

	logger.Log.Info("Generating reset methods",
		zap.String("package", packageName),
		zap.String("file dest", outputFile),
		zap.Int("count", len(structs)))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, packageInfo{Name: packageName, StructsInfo: structs})
	if err != nil {
		return fmt.Errorf("template execution error: %w", err)
	}

	bufFmt, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting go-file error: %w", err)
	}

	err = os.WriteFile(outputFile, bufFmt, 0644)
	if err != nil {
		return fmt.Errorf("writing file error: %w", err)
	}

	return nil
}

func getResetCode(name string, expr ast.Expr) (string, error) {

	switch t := expr.(type) {
	case *ast.Ident:
		return getSimpleResetCode(name, t.Name), nil
	case *ast.StarExpr:
		rCode, err := getResetCode(name, t.X)
		if err != nil {
			return "", fmt.Errorf("cannot parse star expr: %w", err)
		}
		var buf bytes.Buffer
		buf.WriteString("if v." + name + " != nil {\n")
		buf.WriteString("\t*" + rCode + "\n")
		buf.WriteString("}\n")
		return buf.String(), nil
	case *ast.ArrayType:
		return "v." + name + " = v." + name + "[:0]", nil
	case *ast.MapType:
		return "clear(v." + name + ")", nil
	case *ast.StructType:
		return getResetCodeStruct(name), nil
	default:
		return "", fmt.Errorf("no reset code for expr")
	}
}

// getZeroValue возвращает zero value для типа
func getZeroValue(typeName string) string {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return "0"
	case "float32", "float64":
		return "0.0"
	case "bool":
		return "false"
	case "string":
		return "\"\""
	default:
		return "nil"
	}
}

func getResetCodeStruct(name string) string {
	var buf bytes.Buffer
	buf.WriteString("if resetter, ok := v." + name + ".(interface{ Reset() }); ok && v." + name + " != nil {\n")
	buf.WriteString("\tresetter.Reset()\n")
	buf.WriteString("}")
	return buf.String()
}

func getSimpleResetCode(name string, typeName string) string {
	return "v." + name + " = " + getZeroValue(typeName)
}
