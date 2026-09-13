# wealth-warden-prepper

Turns bank statements into a JSON file that
[wealth-warden](https://github.com/nootey/wealth-warden) can import through its
custom transaction import. 

Supported banks:

| Bank | Format | Notes |
|------|--------|-------|
| NLB (Slovenia) | CSV from NLB Klik, 1 to 12 months | Inflow and outflow exports, any mix of files |
| NLB (Slovenia) | Monthly PDF statement (Izpisek) | Needs `pdftotext` from poppler |

## Requirements

- Go 1.26+
- `pdftotext` (package `poppler-utils`) for PDF statements. CSV files work without it.

## Usage

```bash
make build
./bin/prepper data/input/nlb/inflows_2021.csv data/input/nlb/outflows_2021.csv data/output/nlb_2021.json
./bin/prepper data/input/nlb/Izpisek_2026_08.pdf data/output/nlb_2026_08.json
```

Arguments are the input statements followed by the output path. Optional
flags go before them:

| Flag | Meaning |
|------|---------|
| `-bank nlb` | Bank format. |
| `-id name` | Import identifier. Default: `<bank>_<year>` or `<bank>_<first>_<last>`. |
| `-debug` | Log at debug level. |

Many files can be passed at once. Rows that share a bank transaction ID are
de-duplicated, so overlapping CSV exports are safe. The output is sorted by date.

## Output

```json
{
  "identifier": "nlb_2021",
  "generated_at": "2026-09-13T12:27:53Z",
  "transactions": [
    {
      "transaction_type": "expense",
      "amount": "19.90",
      "currency": "EUR",
      "txn_date": "2021-12-20T00:00:00Z",
      "category": "(uncategorized)",
      "description": "Food"
    }
  ]
}
```

Every transaction gets the `(uncategorized)` category.
Bank exports carry no useful categories, and the rule system for assigning
them is handled in wealth-warden. The `description` field carries the raw
counterparty and purpose for reference.

## Usage as a library

The packages under `pkg/` are importable as a library:

```go
import (
    "github.com/nootey/wealth-warden-prepper/pkg/bank"
    "github.com/nootey/wealth-warden-prepper/pkg/format"
)

parser, ok := bank.Get("nlb")
txns, err := parser.ParseCSV(reader) // or parser.ParsePDF(reader)
payload := format.Build("nlb_2021", txns, time.Now())
```

## Adding a bank

1. Create `pkg/bank<name>` with a type that implements `statement.Parser`.
2. Add it to the registry in `pkg/bank/bank.go`.

## Development

```bash
make test
make lint
```
