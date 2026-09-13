package banknlb

import (
	"io"

	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

type NLB struct{}

var _ statement.Parser = NLB{}

func (NLB) ParseCSV(r io.Reader) ([]statement.Transaction, error) { return ParseCSV(r) }
func (NLB) ParsePDF(r io.Reader) ([]statement.Transaction, error) { return ParsePDF(r) }
