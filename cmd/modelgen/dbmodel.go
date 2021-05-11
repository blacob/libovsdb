package main

import (
	"text/template"

	"github.com/ovn-org/libovsdb/ovsdb"
)

const MODEL_TEMPLATE = `
// Code generated by "libovsdb.modelgen"
// DO NOT EDIT.

package {{ .PackageName }}

import (
	goovn "github.com/ovn-org/libovsdb"
)

// FullDatabaseModel() returns the DatabaseModel object to be used in libovsdb
func FullDatabaseModel() (*goovn.DBModel, error) {
	return goovn.NewDBModel("{{ .DatabaseName }}", map[string]goovn.Model{
    {{ range $tableName, $structName := .Tables }} "{{ $tableName }}" : &{{ $structName }}{}, 
    {{ end }}
	})
}
`

//DBModelTemplateData is the data needed for template processing
type DBModelTemplateData struct {
	PackageName  string
	DatabaseName string
	Tables       map[string]string
}

//NewDBModelGenerator returns a new DBModel generator
func NewDBModelGenerator(pkg string, schema *ovsdb.DatabaseSchema) Generator {
	templateData := DBModelTemplateData{
		PackageName:  pkg,
		DatabaseName: schema.Name,
		Tables:       map[string]string{},
	}
	for tableName := range schema.Tables {
		templateData.Tables[tableName] = StructName(tableName)
	}
	modelTemplate := template.Must(template.New("DBModel").Parse(MODEL_TEMPLATE))
	return newGenerator("model.go", modelTemplate, templateData)
}
