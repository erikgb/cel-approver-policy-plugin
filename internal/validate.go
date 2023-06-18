package internal

import (
	"errors"
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"reflect"
)

type Validator struct {
	Expression string
	Program    cel.Program
}

func (v *Validator) Compile() error {
	env, err := cel.NewEnv(
		cel.Variable("self", cel.StringType),
		cel.Variable("cr", cel.MapType(cel.StringType, cel.StringType)),
		ext.Strings(),
	)
	if err != nil {
		return err
	}

	ast, iss := env.Compile(v.Expression)
	if iss.Err() != nil {
		return iss.Err()
	}
	if !reflect.DeepEqual(ast.OutputType(), cel.BoolType) {
		return fmt.Errorf(
			"got %v, wanted %v result type", ast.OutputType(), cel.BoolType)
	}

	v.Program, err = env.Program(ast)
	return err
}

func (v *Validator) Validate(val string, cr CertificateRequest) (bool, error) {
	if v.Program == nil {
		return false, errors.New("must compile first")
	}

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
