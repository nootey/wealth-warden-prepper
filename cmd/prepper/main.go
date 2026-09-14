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
	bankName := fs.String("bank", "auto", "bank format: auto (default, detected per file), or one of: "+strings.Join(bank.Names(), ", "))
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
	var parser statement.Parser
	if bankName != "auto" {
		p, ok := bank.Get(bankName)
		if !ok {
			return fmt.Errorf("unsupported bank %q", bankName)
		}
		parser = p
	}

	var txns []statement.Transaction
	detected := bankName
	for _, path := range inputs {
		parsed, name, err := parseFile(parser, bankName, path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if name != "" {
			log.Info("Parsed statement", zap.String("file", path), zap.String("bank", name), zap.Int("transactions", len(parsed)))
			detected = name
		} else {
			log.Info("Parsed statement", zap.String("file", path), zap.Int("transactions", len(parsed)))
		}
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
		id = defaultIdentifier(detected, txns)
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

func parseFile(parser statement.Parser, bankName, path string) ([]statement.Transaction, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = f.Close() }()

	ext := strings.ToLower(filepath.Ext(path))
	if bankName == "auto" {
		return bank.ParseAuto(f, ext)
	}

	switch ext {
	case ".csv":
		txns, err := parser.ParseCSV(f)
		return txns, "", err
	case ".pdf":
		txns, err := parser.ParsePDF(f)
		return txns, "", err
	default:
		return nil, "", fmt.Errorf("unsupported file type %q", ext)
	}
}

func defaultIdentifier(bank string, txns []statement.Transaction) string {
	first, last := txns[0].Date.Year(), txns[len(txns)-1].Date.Year()
	if first == last {
		return fmt.Sprintf("%s_%d", bank, first)
	}
	return fmt.Sprintf("%s_%d_%d", bank, first, last)
}
