package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/senorMk/pdf-tools/apps/desktop/internal/pdf"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Desktop keeps selection and document contents in the native process.
// Only metadata and the entered password cross the frontend bridge.
type Desktop struct {
	ctx        context.Context
	mu         sync.Mutex
	selected   *pdf.Document
	openDialog func(context.Context, runtime.OpenDialogOptions) (string, error)
	saveDialog func(context.Context, runtime.SaveDialogOptions) (string, error)
}

func newDesktop() *Desktop {
	return &Desktop{openDialog: runtime.OpenFileDialog, saveDialog: runtime.SaveFileDialog}
}

var errBusy = errors.New("A file operation is already running. Please wait for it to finish.")
var pdfFilter = []runtime.FileFilter{{DisplayName: "PDF documents", Pattern: "*.pdf"}}

func (d *Desktop) SelectFile() (*pdf.Document, error) {
	if !d.mu.TryLock() {
		return nil, errBusy
	}
	defer d.mu.Unlock()
	path, err := d.openDialog(d.ctx, runtime.OpenDialogOptions{Title: "Choose a PDF", Filters: pdfFilter})
	if err != nil {
		return nil, errors.New("Could not open the file picker. Please try again.")
	}
	if path == "" {
		return nil, nil
	}
	return d.selectPath(path)
}

func (d *Desktop) UseDroppedFile(paths []string) (*pdf.Document, error) {
	if !d.mu.TryLock() {
		return nil, errBusy
	}
	defer d.mu.Unlock()
	if len(paths) != 1 {
		return nil, errors.New("Drop one PDF at a time.")
	}
	return d.selectPath(paths[0])
}

func (d *Desktop) selectPath(path string) (*pdf.Document, error) {
	document, err := pdf.Inspect(path)
	if err != nil {
		return nil, err
	}
	d.selected = document
	return document, nil
}

func (d *Desktop) ClearFile() error {
	if !d.mu.TryLock() {
		return errBusy
	}
	defer d.mu.Unlock()
	d.selected = nil
	return nil
}

type SaveResult struct {
	Cancelled bool   `json:"cancelled"`
	Path      string `json:"path,omitempty"`
}

func (d *Desktop) RemovePassword(password string) (SaveResult, error) {
	if !d.mu.TryLock() {
		return SaveResult{}, errBusy
	}
	defer d.mu.Unlock()
	if d.selected == nil {
		return SaveResult{}, errors.New("Choose a PDF first.")
	}
	data, err := pdf.Decrypt(d.selected.Path, password)
	password = ""
	if err != nil {
		return SaveResult{}, err
	}
	defer clear(data)
	name := strings.TrimSuffix(d.selected.Name, filepath.Ext(d.selected.Name)) + "-unlocked.pdf"
	path, err := d.saveDialog(d.ctx, runtime.SaveDialogOptions{
		Title: "Save your unlocked PDF", DefaultFilename: name,
		DefaultDirectory: filepath.Dir(d.selected.Path), Filters: pdfFilter,
		CanCreateDirectories: true,
	})
	if err != nil {
		return SaveResult{}, errors.New("Could not open the save dialog. Please try again.")
	}
	if path == "" {
		return SaveResult{Cancelled: true}, nil
	}
	if filepath.Ext(path) == "" {
		path += ".pdf"
		// The native panel did not confirm replacing this expanded filename.
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			return SaveResult{}, errors.New("A file with this name already exists. Save again using a full .pdf filename or a different name.")
		}
	}
	if err := pdf.Save(d.selected.Path, path, data); err != nil {
		return SaveResult{}, err
	}
	return SaveResult{Path: path}, nil
}
