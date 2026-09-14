package bank

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	"github.com/nootey/wealth-warden-prepper/pkg/pdftext"
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

func DetectCSV(header []string) string {
	best, bestScore := "generic", 0
	for name, p := range registry {
		d, ok := p.(statement.Detector)
		if !ok {
			continue
		}
		if score := d.DetectCSV(header); score > bestScore {
			best, bestScore = name, score
		}
	}
	return best
}

func DetectPDF(lines []string) (name string, ok bool) {
	bestScore := 0
	for n, p := range registry {
		d, detOk := p.(statement.Detector)
		if !detOk {
			continue
		}
		if score := d.DetectPDF(lines); score > bestScore {
			name, bestScore = n, score
		}
	}
	return name, bestScore > 0
}

func ParseCSVAuto(r io.Reader) ([]statement.Transaction, string, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, "", err
	}
	header, err := statement.PeekCSVHeader(raw)
	if err != nil {
		return nil, "", fmt.Errorf("read csv header: %w", err)
	}
	name := DetectCSV(header)
	parser, _ := Get(name)
	txns, err := parser.ParseCSV(bytes.NewReader(raw))
	return txns, name, err
}

func ParsePDFAuto(r io.Reader) ([]statement.Transaction, string, error) {
	lines, err := pdftext.ExtractReader(r)
	if err != nil {
		return nil, "", err
	}
	name, ok := DetectPDF(lines)
	if !ok {
		return nil, "", fmt.Errorf("could not detect bank from pdf layout; specify a bank explicitly")
	}
	parser, _ := Get(name)
	lp, ok := parser.(statement.LinesParser)
	if !ok {
		return nil, "", fmt.Errorf("bank %q has no pdf line parser", name)
	}
	txns, err := lp.ParsePDFLines(lines)
	return txns, name, err
}
