package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/senorMk/pdf-tools/apps/desktop/internal/pdf"
	"github.com/senorMk/pdf-tools/apps/desktop/internal/testpdf"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func TestUnlockThenSave(t *testing.T) {
	source := testpdf.Encrypted(t, "known", "owner", true, 256)
	d := newDesktop()
	if _, err := d.UseDroppedFile([]string{source}); err != nil {
		t.Fatal(err)
	}
	dialogCalls := 0
	destination := filepath.Join(t.TempDir(), "chosen.pdf")
	d.saveDialog = func(_ context.Context, options runtime.SaveDialogOptions) (string, error) {
		dialogCalls++
		if options.DefaultFilename != "protected-unlocked.pdf" {
			t.Fatalf("default filename: %s", options.DefaultFilename)
		}
		return destination, nil
	}
	if _, err := d.RemovePassword("wrong"); !errors.Is(err, pdf.ErrPassword) {
		t.Fatalf("wrong password: %v", err)
	}
	if dialogCalls != 0 {
		t.Fatal("save dialog opened for wrong password")
	}
	result, err := d.RemovePassword("known")
	if err != nil || result.Cancelled || result.Path != destination {
		t.Fatalf("save: %+v, %v", result, err)
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if err := api.Validate(context.Background(), bytes.NewReader(data), model.NewDefaultConfiguration(), nil); err != nil {
		t.Fatalf("saved PDF: %v", err)
	}
	d.saveDialog = func(context.Context, runtime.SaveDialogOptions) (string, error) { return "", nil }
	result, err = d.RemovePassword("known")
	if err != nil || !result.Cancelled {
		t.Fatalf("cancel: %+v, %v", result, err)
	}
	d.saveDialog = func(context.Context, runtime.SaveDialogOptions) (string, error) { return source, nil }
	if _, err := d.RemovePassword("known"); err == nil {
		t.Fatal("source overwrite allowed")
	}
	if err := d.ClearFile(); err != nil {
		t.Fatal(err)
	}
	if _, err := d.RemovePassword("known"); err == nil {
		t.Fatal("unlock without selection accepted")
	}
}

func TestSelectionAndBusyOperations(t *testing.T) {
	source := testpdf.Encrypted(t, "known", "owner", true, 128)
	d := newDesktop()
	if _, err := d.UseDroppedFile([]string{source, source}); err == nil {
		t.Fatal("multiple files accepted")
	}
	selected, err := d.UseDroppedFile([]string{source})
	if err != nil {
		t.Fatal(err)
	}
	d.openDialog = func(context.Context, runtime.OpenDialogOptions) (string, error) { return "", nil }
	if result, err := d.SelectFile(); result != nil || err != nil {
		t.Fatalf("cancel selection: %v, %v", result, err)
	}
	if d.selected != selected {
		t.Fatal("cancel cleared prior selection")
	}
	if _, err := d.UseDroppedFile([]string{"missing.pdf"}); err == nil {
		t.Fatal("missing file accepted")
	}
	if d.selected != selected {
		t.Fatal("invalid selection cleared prior selection")
	}
	d.mu.Lock()
	if _, err := d.RemovePassword("known"); !errors.Is(err, errBusy) {
		t.Fatalf("overlap: %v", err)
	}
	if err := d.ClearFile(); !errors.Is(err, errBusy) {
		t.Fatalf("clear overlap: %v", err)
	}
	d.mu.Unlock()
}
