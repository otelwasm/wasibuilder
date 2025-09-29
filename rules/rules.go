package rules

import (
	"slices"

	"github.com/otelwasm/wasibuilder/internal/version"
)

type ExecContext struct {
	Command      string
	Args         []string
	Package      string
	PackageIndex int
	GoVersion    *version.Version
}

func (ctx *ExecContext) Clone() *ExecContext {
	return &ExecContext{
		Command:      ctx.Command,
		Args:         slices.Clone(ctx.Args),
		Package:      ctx.Package,
		PackageIndex: ctx.PackageIndex,
		GoVersion:    ctx.GoVersion,
	}
}

type Rule interface {
	Apply(ctx *ExecContext) error
	Name() string
}
