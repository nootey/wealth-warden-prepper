package pdftext

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var ErrNotInstalled = errors.New("pdftotext is not installed (install poppler-utils)")

// Page breaks become lines containing "\f".
func Extract(path string) ([]string, error) {
	bin, err := exec.LookPath("pdftotext")
	if err != nil {
		return nil, ErrNotInstalled
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(bin, "-layout", "-enc", "UTF-8", path, "-")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdftotext %s: %w: %s", path, err, strings.TrimSpace(stderr.String()))
	}
	return SplitLines(stdout.String()), nil
}

// pdftotext needs a path.
func ExtractReader(r io.Reader) ([]string, error) {
	tmp, err := os.CreateTemp("", "prepper-*.pdf")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := io.Copy(tmp, r); err != nil {
		_ = tmp.Close()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}
	return Extract(tmp.Name())
}

func SplitLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\f", "\n\f\n")
	return strings.Split(text, "\n")
}
