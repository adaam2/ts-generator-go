package generator

import (
	"fmt"
	"strings"
)

type Construct interface {
	Generate() string
}

func Ternary[T any](condition bool, If, Else T) T {
	if condition {
		return If
	}
	return Else
}

type GeneratorOptions struct {
	Indent int
}

func NewGenerator(outDir string, opts *GeneratorOptions) *Generator {
	gen := &Generator{
		outDir: outDir,
		indent: opts.Indent,
	}

	return gen
}

type Generator struct {
	outDir      string
	sourceFiles []*SourceFile
	indent      int
}

func (api *Generator) Generate() error {
	return nil
}

func (api *Generator) String() (r string) {
	for _, sf := range api.sourceFiles {
		r += fmt.Sprintf("\n// %s\n", sf.Path)
		r += sf.Generate()
	}

	return r
}

func (api *Generator) AddSourceFile(path string) *SourceFile {
	sf := &SourceFile{
		Path:      path,
		generator: api,
	}

	api.sourceFiles = append(api.sourceFiles, sf)

	return sf
}

type SourceFile struct {
	Path string

	interfaces []*Interface
	classes    []*Class
	generator  *Generator
}

func (sf *SourceFile) Generate() (r string) {
	for _, i := range sf.interfaces {
		r += fmt.Sprintf("%s\n", i.Generate())
	}

	for _, c := range sf.classes {
		r += fmt.Sprintf("%s\n", c.Generate())
	}

	return r
}

type Property struct {
	name      string
	typeInfo  string
	generator *Generator
}

type Interface struct {
	name       string
	properties []*Property
	export     bool
	generator  *Generator
}

func (i *Interface) Generate() (r string) {
	r += fmt.Sprintf("%sinterface %s {\n", Ternary(i.export, "export ", ""), i.name)

	for _, prop := range i.properties {
		r += fmt.Sprintf("%s%s: %s;\n", strings.Repeat(" ", i.generator.indent), prop.name, prop.typeInfo)
	}

	r += "}\n"
	return r
}

type ClassMethod struct {
	name       string
	returnType string
	parameters *ParameterCollection
	generator  *Generator
	scope      string
}

func (m *ClassMethod) Generate() (r string) {
	r += fmt.Sprintf("%s%s = (%s) : %s => {\n", Ternary(m.scope != "", fmt.Sprintf("%s ", m.scope), ""), m.name, m.parameters.Generate(), m.returnType)

	r += fmt.Sprintf("%s}\n", strings.Repeat(" ", m.generator.indent))
	return r
}

func (m *ClassMethod) SetReturnType(typeInfo string) {
	m.returnType = typeInfo
}

func (m *ClassMethod) SetScope(scope string) {
	m.scope = scope
}

func (m *ClassMethod) AddParameter(name string, t string) *Parameter {
	p := &Parameter{
		name:      name,
		typeInfo:  t,
		generator: m.generator,
	}

	if m.parameters == nil {
		m.parameters = &ParameterCollection{}
	}

	m.parameters.Parameters = append(m.parameters.Parameters, p)

	return p
}

type Class struct {
	name        string
	export      bool
	extends     string
	constructor *Constructor
	properties  []*ClassProperty

	classMethods []*ClassMethod
	generator    *Generator
}

type ClassProperty struct {
	name      string
	typeInfo  string
	generator *Generator
	scope     string
}

func (p *ClassProperty) Generate() (r string) {
	r += fmt.Sprintf("%s%s%s : %s;\n", strings.Repeat(" ", p.generator.indent), Ternary(p.scope != "", fmt.Sprintf("%s ", p.scope), ""), p.name, p.typeInfo)
	return r
}

func (c *Class) Generate() (r string) {
	r += fmt.Sprintf("%sclass %s%s {\n", Ternary(c.export, "export ", ""), c.name, Ternary(c.extends != "", fmt.Sprintf(" %s ", c.extends), ""))

	for _, prop := range c.properties {
		r += prop.Generate()
	}

	r += c.constructor.Generate()

	for _, method := range c.classMethods {
		r += fmt.Sprintf("%s%s\n", strings.Repeat(" ", c.generator.indent), method.Generate())
	}

	r += "}\n"

	return r
}

type Assignment struct {
	lhs string
	rhs string

	generator *Generator
}

func (a *Assignment) Generate() string {
	return fmt.Sprintf("%s = %s;", a.lhs, a.rhs)
}

type ParameterCollection struct {
	Parameters []*Parameter
}

func (pc *ParameterCollection) Generate() (r string) {
	for i, p := range pc.Parameters {
		if len(pc.Parameters) == 1 || i == len(pc.Parameters)-1 {
			r += p.Generate()
		} else {
			r += fmt.Sprintf("%s, ", p.Generate())
		}
	}
	return r
}

type Constructor struct {
	assignments []*Assignment
	parameters  *ParameterCollection
	generator   *Generator
}

func (c *Constructor) Generate() (r string) {
	r += fmt.Sprintf("%sconstructor(%s) {\n", strings.Repeat(" ", c.generator.indent), c.parameters.Generate())
	for _, a := range c.assignments {
		r += fmt.Sprintf("%s%s\n", strings.Repeat(" ", c.generator.indent*2), a.Generate())
	}
	r += fmt.Sprintf("%s}\n", strings.Repeat(" ", c.generator.indent))
	return r
}

func (cb *Constructor) AddParameter(name string, typeInfo string) {
	p := &Parameter{
		name:     name,
		typeInfo: typeInfo,
	}

	if cb.parameters == nil {
		cb.parameters = &ParameterCollection{}
	}

	cb.parameters.Parameters = append(cb.parameters.Parameters, p)
}

func (cb *Constructor) AddAssignment(lhs string, rhs string) {
	assignment := &Assignment{
		lhs: lhs,
		rhs: rhs,
	}

	cb.assignments = append(cb.assignments, assignment)
}

func (c *Class) AddConstructor(fn func(c *Constructor)) {
	c.constructor = &Constructor{
		generator: c.generator,
	}
	fn(c.constructor)
}

func (c *Class) AddMember(name string, typeInfo string, scope string) {
	prop := &ClassProperty{
		name:      name,
		typeInfo:  typeInfo,
		scope:     scope,
		generator: c.generator,
	}

	c.properties = append(c.properties, prop)
}

func (c *Class) AddClassMethod(name string, builder func(method *ClassMethod)) {
	m := &ClassMethod{
		name:      name,
		generator: c.generator,
	}
	c.classMethods = append(c.classMethods, m)

	builder(m)
}

type Parameter struct {
	name      string
	typeInfo  string
	generator *Generator
}

func (p *Parameter) Generate() string {
	return fmt.Sprintf("%s: %s", p.name, p.typeInfo)
}

func (i *Interface) AddProperty(name string, propType string) {
	prop := Property{
		name:      name,
		typeInfo:  propType,
		generator: i.generator,
	}

	i.properties = append(i.properties, &prop)
}

func (sf *SourceFile) AddInterface(name string, export bool, builder func(i *Interface)) {
	i := &Interface{name: name, generator: sf.generator, export: export}
	sf.interfaces = append(sf.interfaces, i)

	builder(i)
}

func (sf *SourceFile) AddClass(name string, export bool, builder func(c *Class)) {
	c := &Class{
		name:      name,
		generator: sf.generator,
		export:    export,
	}

	sf.classes = append(sf.classes, c)

	builder(c)
}
