package generator_test

import (
	"os"
	"testing"

	generator "github.com/adaam2/ts-generator-go"
	"gotest.tools/v3/assert"
)

func TestGenerator(t *testing.T) {
	tmpDir := os.TempDir()

	t.Run("round trip", func(t *testing.T) {
		cg := generator.NewGenerator(tmpDir, &generator.GeneratorOptions{
			Indent: 2,
		})

		sf := cg.AddSourceFile("test.ts")

		sf.AddInterface("TestInterface", true, func(i *generator.Interface) {
			i.AddProperty("testProp", "string")
			i.AddProperty("anotherProp", "number")
		})

		sf.AddClass("MyClass", true, func(c *generator.Class) {
			c.AddMember("something", "string", "private")

			c.AddConstructor(func(cb *generator.Constructor) {
				cb.AddParameter("something", "string")
				cb.AddAssignment("this.something", "something")
			})

			c.AddClassMethod("doSomething", func(method *generator.ClassMethod) {
				method.SetReturnType("string")
				method.SetScope("private")
				method.AddParameter("input", "string")
				method.AddParameter("anotherInput", "number")
			})
		})

		out := cg.String()

		expected := `
// test.ts
export interface TestInterface {
  testProp: string;
  anotherProp: number;
}

export class MyClass {
  private something : string;
  constructor(something: string) {
    this.something = something;
  }
  private doSomething = (input: string, anotherInput: number) : string => {
  }

}

`
		assert.Equal(
			t,
			out,
			expected,
		)
	})
}
