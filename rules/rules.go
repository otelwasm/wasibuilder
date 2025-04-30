package rules

import "slices"

type ExecContext struct {
	Command      string
	Args         []string
	Package      string
	PackageIndex int
}

func (ctx *ExecContext) Clone() *ExecContext {
	return &ExecContext{
		Command:      ctx.Command,
		Args:         slices.Clone(ctx.Args),
		Package:      ctx.Package,
		PackageIndex: ctx.PackageIndex,
	}
}

type Rule interface {
	Apply(ctx *ExecContext) error
	Name() string
}
