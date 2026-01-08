package defs

import (
	"go/token"
)

// VariableInfo contiene información detallada sobre una variable.
type VariableInfo struct {
	Name     string
	Type     string
	Position token.Position
	IsGlobal bool
	Kind     string
}

// MethodInfo contiene información detallada sobre un método.
type MethodInfo struct {
	Name         string
	Receiver     string
	InputParams  []ParameterInfo
	OutputParams []ParameterInfo
}

// FunctionInfo contiene información detallada sobre una función.
type FunctionInfo struct {
	Name         string
	InputParams  []ParameterInfo
	OutputParams []ParameterInfo
}

// ParameterInfo contiene información sobre un parámetro.
type ParameterInfo struct {
	Name string
	Type string
}
