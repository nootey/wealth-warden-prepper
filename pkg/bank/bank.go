package bank

import (
	"sort"

	"github.com/nootey/wealth-warden-prepper/pkg/rules/generic"
	"github.com/nootey/wealth-warden-prepper/pkg/rules/n26"
	"github.com/nootey/wealth-warden-prepper/pkg/rules/nlb"
	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

var registry = map[string]statement.Parser{
	"nlb":     nlb.NLB{},
	"n26":     n26.N26{},
	"generic": generic.Generic{},
}

func Get(name string) (statement.Parser, bool) {
	p, ok := registry[name]
	return p, ok
}

func Names() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
