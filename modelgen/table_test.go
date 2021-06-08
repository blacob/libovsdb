package modelgen

import (
	"encoding/json"
	"fmt"
	"testing"
	"text/template"

	"github.com/ovn-org/libovsdb/ovsdb"
	"github.com/stretchr/testify/assert"
)

func TestNewTableTemplate(t *testing.T) {
	rawSchema := []byte(`
	{
		"name": "AtomicDB",
		"version": "0.0.0",
		"tables": {
			"atomicTable": {
				"columns": {
					"str": {
						"type": "string"
					},
					"int": {
						"type": "integer"
					},
					"float": {
						"type": "real"
					}
				}
			}
		}
	}`)

	test := []struct {
		name      string
		extend    func(tmpl *template.Template, data map[string]interface{})
		expected  string
		err       bool
		formatErr bool
	}{
		{
			name: "normal",
			expected: `// Code generated by "libovsdb.modelgen"
// DO NOT EDIT.

package test

// test defines an object in test table
type test struct {
	UUID  string  ` + "`" + `ovs:"_uuid"` + "`" + `
	Float float64 ` + "`" + `ovs:"float"` + "`" + `
	Int   int     ` + "`" + `ovs:"int"` + "`" + `
	Str   string  ` + "`" + `ovs:"str"` + "`" + `
}
`,
		},
		{
			name: "add fields using same data",
			extend: func(tmpl *template.Template, data map[string]interface{}) {
				extra := `{{ define "extraFields" }} {{ range index . "Fields" }}	Other{{ .Name }}  {{ .Type }}
{{end}} {{end}}`
				_, err := tmpl.Parse(extra)
				if err != nil {
					panic(err)
				}
			},
			expected: `// Code generated by "libovsdb.modelgen"
// DO NOT EDIT.

package test

// test defines an object in test table
type test struct {
	UUID  string  ` + "`" + `ovs:"_uuid"` + "`" + `
	Float float64 ` + "`" + `ovs:"float"` + "`" + `
	Int   int     ` + "`" + `ovs:"int"` + "`" + `
	Str   string  ` + "`" + `ovs:"str"` + "`" + `

	OtherUUID  string
	OtherFloat float64
	OtherInt   int
	OtherStr   string
}
`,
		},
		{
			name: "add extra functions using extra data",
			extend: func(tmpl *template.Template, data map[string]interface{}) {
				extra := `{{ define "postStructDefinitions" }}
func {{ index . "TestName" }} () string {
    return "{{ index . "StructName" }}"
} {{ end }}
`
				_, err := tmpl.Parse(extra)
				if err != nil {
					panic(err)
				}
				data["TestName"] = "TestFunc"
			},
			expected: `// Code generated by "libovsdb.modelgen"
// DO NOT EDIT.

package test

// test defines an object in test table
type test struct {
	UUID  string  ` + "`" + `ovs:"_uuid"` + "`" + `
	Float float64 ` + "`" + `ovs:"float"` + "`" + `
	Int   int     ` + "`" + `ovs:"int"` + "`" + `
	Str   string  ` + "`" + `ovs:"str"` + "`" + `
}

func TestFunc() string {
	return "test"
}
`,
		},
		{
			name:      "add bad code",
			formatErr: true,
			extend: func(tmpl *template.Template, data map[string]interface{}) {
				extra := `{{ define "preStructDefinitions" }}
WRONG FORMAT
{{ end }}
`
				_, err := tmpl.Parse(extra)
				if err != nil {
					panic(err)
				}
			},
		},
	}

	var schema ovsdb.DatabaseSchema
	err := json.Unmarshal(rawSchema, &schema)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range test {
		t.Run(fmt.Sprintf("Table Test: %s", tt.name), func(t *testing.T) {
			table := schema.Tables["atomicTable"]
			templ, data := NewTableTemplate(
				"test",
				"test",
				&table,
			)
			if tt.err {
				assert.NotNil(t, err)
			} else {
				if tt.extend != nil {
					tt.extend(templ, data)
				}

				for i := 0; i < 3; i++ {
					g := NewGenerator(false)
					b, err := g.Format(templ, data)
					if tt.formatErr {
						assert.NotNil(t, err)
					} else {
						assert.Nil(t, err)
						assert.Equal(t, tt.expected, string(b))
					}
				}
			}
		})
	}
}

func TestFieldName(t *testing.T) {
	cases := []struct {
		in       string
		expected string
	}{
		{"foo", "Foo"},
	}
	for _, tt := range cases {
		if s := FieldName(tt.in); s != tt.expected {
			t.Fatalf("got %s, wanted %s", s, tt.expected)
		}
	}

}

func TestStructName(t *testing.T) {
	if s := StructName("Foo_Bar"); s != "FooBar" {
		t.Fatalf("got %s, wanted FooBar", s)
	}
}

/*
TODO: Write Test
func TestFieldType(t *testing.T) {
	tests := []struct {
		name string
		in   *ovsdb.ColumnSchema
		out  string
	}
	if got := FieldType(tt.args.column); got != tt.want {
		t.Errorf("FieldType() = %v, want %v", got, tt.want)
	}
}
*/

func TestAtomicType(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"IntegerToInt", ovsdb.TypeInteger, "int"},
		{"RealToFloat", ovsdb.TypeReal, "float64"},
		{"BooleanToBool", ovsdb.TypeBoolean, "bool"},
		{"StringToString", ovsdb.TypeString, "string"},
		{"UUIDToString", ovsdb.TypeUUID, "string"},
		{"Invalid", "notAType", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AtomicType(tt.in); got != tt.out {
				t.Errorf("got %s, wanted %s", got, tt.out)
			}
		})
	}
}

func TestTag(t *testing.T) {
	if s := Tag("Foo_Bar"); s != "ovs:\"Foo_Bar\"" {
		t.Fatalf("got %s, wanted ovs:\"Foo_Bar\"", s)
	}
}

func TestFileName(t *testing.T) {
	if s := FileName("foo"); s != "foo.go" {
		t.Fatalf("got %s, wanted foo.go", s)
	}
}

func TestCamelCase(t *testing.T) {
	cases := []struct {
		in       string
		expected string
	}{
		{"foo_bar_baz", "FooBarBaz"},
		{"foo-bar-baz", "FooBarBaz"},
		{"foos-bars-bazs", "FoosBarsBazs"},
		{"ip_port_mappings", "IPPortMappings"},
		{"external_ids", "ExternalIDs"},
		{"ip_prefix", "IPPrefix"},
		{"dns_records", "DNSRecords"},
		{"logical_ip", "LogicalIP"},
		{"ip", "IP"},
	}
	for _, tt := range cases {
		if s := camelCase(tt.in); s != tt.expected {
			t.Fatalf("got %s, wanted %s", s, tt.expected)
		}
	}
}

func ExampleNewTableTemplate() {
	schemaString := []byte(`
	{
		"name": "MyDB",
		"version": "0.0.0",
		"tables": {
			"table1": {
				"columns": {
					"string_column": {
						"type": "string"
					},
					"some_integer": {
						"type": "integer"
					},
				}
			}
		}
	}`)
	var schema ovsdb.DatabaseSchema
	_ = json.Unmarshal(schemaString, &schema)

	base, data := NewTableTemplate("mypackage", "table1", schema.Table("table1"))

	// Add a function at after the struct definition
	// It can access the default data values plus any extra field that is added to data
	_, _ = base.Parse(`{{define "postStructDefinitions"}}
func (t {{ index . "StructName" }} {{ index . FuncName}}() string {
    return "bar"
}{{end}}`)
	data["FuncName"] = "TestFunc"

	gen := NewGenerator(false)
	_ = gen.Generate("generated.go", base, data)
}
