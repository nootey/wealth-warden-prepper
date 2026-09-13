package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nootey/wealth-warden-prepper/pkg/bank"
	"github.com/nootey/wealth-warden-prepper/pkg/format"
	"github.com/nootey/wealth-warden-prepper/pkg/logger"
	"github.com/nootey/wealth-warden-prepper/pkg/statement"
	"go.uber.org/zap"
)

func main() {
	fs := flag.NewFlagSet("prepper", flag.ContinueOnError)
	bankName := fs.String("bank", "nlb", "bank format ("+strings.Join(bank.Names(), ", ")+")")
	id := fs.String("id", "", "import identifier (default: bank_<year range>)")
	debug := fs.Bool("debug", false, "log at debug level")
	fs.Usage = func() {
		_, _ = fmt.Fprintf(fs.Output(), "usage: prepper [flags] <input statements...> <output.json>\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if fs.NArg() < 2 {
		fs.Usage()
		os.Exit(2)
	}
	args := fs.Args()
	inputs, out := args[:len(args)-1], args[len(args)-1]

	log := logger.InitLogger(*debug)
	defer func() { _ = log.Sync() }()

	if err := run(log, *bankName, *id, inputs, out); err != nil {
		log.Fatal("prepper failed", zap.Error(err))
	}
}

func run(log *zap.Logger, bankName, id string, inputs []string, out string) error {
	parser, ok := bank.Get(bankName)
	if !ok {
		return fmt.Errorf("unsupported bank %q", bankName)
	}

	var txns []statement.Transaction
	for _, path := range inputs {
		parsed, err := parseFile(parser, path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		log.Info("Parsed statement", zap.String("file", path), zap.Int("transactions", len(parsed)))
		txns = append(txns, parsed...)
	}
	before := len(txns)
	txns = statement.Dedupe(txns)
	if dropped := before - len(txns); dropped > 0 {
		log.Info("Dropped duplicate transactions", zap.Int("count", dropped))
	}
	sort.SliceStable(txns, func(i, j int) bool { return txns[i].Date.Before(txns[j].Date) })
	if len(txns) == 0 {
		return fmt.Errorf("no transactions found")
	}

	if id == "" {
		id = defaultIdentifier(bankName, txns)
	}
	payload := format.Build(id, txns, time.Now())

	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(out, b, 0o644); err != nil {
		return err
	}
	log.Info("Wrote import file",
		zap.String("path", out),
		zap.String("identifier", payload.Identifier),
		zap.Int("transactions", len(payload.Transactions)),
	)
	return nil
}

func parseFile(parser statement.Parser, path string) ([]statement.Transaction, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	switch strings.ToLower(filepath.Ext(path)) {
	case ".csv":
		return parser.ParseCSV(f)
	case ".pdf":
		return parser.ParsePDF(f)
	default:
		return nil, fmt.Errorf("unsupported file type %q", filepath.Ext(path))
	}
}

func defaultIdentifier(bank string, txns []statement.Transaction) string {
	first, last := txns[0].Date.Year(), txns[len(txns)-1].Date.Year()
	if first == last {
		return fmt.Sprintf("%s_%d", bank, first)
	}
	return fmt.Sprintf("%s_%d_%d", bank, first, last)
}
