package internal

import (
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"reflect"
)

type validator struct {
	Expression string
	Program    cel.Program
}

func newValidator(expression string) (*validator, error) {
	env, err := cel.NewEnv(
		cel.Variable("self", cel.StringType),
		cel.Variable("cr", cel.MapType(cel.StringType, cel.StringType)),
		ext.Strings(),
	)
	if err != nil {
		return nil, err
	}

	ast, iss := env.Compile(expression)
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	if !reflect.DeepEqual(ast.OutputType(), cel.BoolType) {
		return nil, fmt.Errorf(
			"got %v, wanted %v result type", ast.OutputType(), cel.BoolType)
	}

	program, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	return &validator{Expression: expression, Program: program}, nil
}

func (v validator) Validate(val string, cr CertificateRequest) (bool, error) {
	vars := map[string]interface{}{
		"self": val,
		"cr": map[string]string{
			"namespace": cr.GetNamespace(),
			"name":      cr.GetName(),
		},
	}

	out, _, err := v.Program.Eval(vars)
	if err != nil {
		return false, err
	}

	return out.Value().(bool), nil
}
