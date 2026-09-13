package bank

import (
	"sort"

	"github.com/nootey/wealth-warden-prepper/pkg/banknlb"
	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

var registry = map[string]statement.Parser{
	"nlb": banknlb.NLB{},
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
