package internal

import (
	"fmt"
	"github.com/google/cel-go/cel"
	"reflect"
)

type validator struct {
	Expression string
	Program    cel.Program
}

func NewValidator(expression string) (*validator, error) {
	env, err := cel.NewEnv(
		cel.Variable("self", cel.StringType),
		cel.Variable("name", cel.StringType),
		cel.Variable("namespace", cel.StringType),
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
