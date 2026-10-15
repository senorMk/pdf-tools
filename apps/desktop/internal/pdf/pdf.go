// Package pdf implements local PDF operations, independently of the desktop UI.
package pdf

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

const MaxFileSize = 100 << 20

var (
	ErrPassword    = errors.New("That password did not unlock this PDF. Check it and try again.")
	ErrUnprotected = errors.New("This PDF is already unlocked. It does not need password removal.")
	ErrInvalid     = errors.New("This PDF could not be read. It may be damaged or use unsupported encryption.")
	ErrTooLarge    = errors.New("This PDF is too large. Choose a file smaller than 100 MB.")
)

func init() {
	// Avoid library config/font caches, persistent state, and certificate lookups.
	api.DisableConfigDir()
}

type Document struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

func open(path string) (*os.File, os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, errors.New("Could not read this file. Check that it exists and you have permission to open it.")
	}
	if !info.Mode().IsRegular() {
		return nil, nil, errors.New("Choose a PDF file, rather than a folder or special file.")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, errors.New("Could not read this file. Check that it exists and you have permission to open it.")
	}
	info, err = f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		f.Close()
		return nil, nil, errors.New("Choose a PDF file, rather than a folder or special file.")
	}
	if info.Size() > MaxFileSize {
		f.Close()
		return nil, nil, ErrTooLarge
	}
	var header [1024]byte
	n, err := f.Read(header[:])
	if (err != nil && err != io.EOF) || !bytes.Contains(header[:n], []byte("%PDF-")) {
		f.Close()
		return nil, nil, errors.New("This file is not a PDF. Choose a valid PDF document.")
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		f.Close()
		return nil, nil, ErrInvalid
	}
	return f, info, nil
}

func Inspect(path string) (*Document, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, ErrInvalid
	}
	f, info, err := open(path)
	if err != nil {
		return nil, err
	}
	f.Close()
	return &Document{Name: filepath.Base(path), Path: path, Size: info.Size()}, nil
}

type limitedBuffer struct{ bytes.Buffer }

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if len(p) > MaxFileSize-b.Len() {
		return 0, ErrTooLarge
	}
	return b.Buffer.Write(p)
}

// Decrypt never writes plaintext to disk. The caller clears the returned buffer.
func Decrypt(path, password string) (data []byte, err error) {
	if password == "" {
		return nil, errors.New("Enter the known PDF password.")
	}
	f, _, err := open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var output limitedBuffer
	// A malformed PDF must not crash the application or expose library diagnostics.
	defer func() {
		if recover() != nil {
			data, err = nil, ErrInvalid
		}
		if err != nil {
			clear(output.Bytes())
		}
	}()
	conf := model.NewDefaultConfiguration()
	conf.Offline = true
	conf.ValidateLinks = false
	conf.Limits.MaxInputBytes = MaxFileSize
	conf.UserPW, conf.OwnerPW = password, password
	conf.ValidationMode = model.ValidationRelaxed
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	err = api.Decrypt(ctx, f, &output, conf)
	conf.UserPW, conf.OwnerPW = "", ""
	if err != nil {
		switch {
		case errors.Is(err, pdfcpu.ErrWrongPassword):
			return nil, ErrPassword
		case errors.Is(err, pdfcpu.ErrNotEncrypted):
			// The encryption check precedes full validation in the library.
			// Do not describe a broken unencrypted file as an unlocked PDF.
			if _, seekErr := f.Seek(0, io.SeekStart); seekErr != nil {
				return nil, ErrInvalid
			}
			if validateErr := api.Validate(ctx, f, conf, nil); validateErr != nil {
				return nil, ErrInvalid
			}
			return nil, ErrUnprotected
		case errors.Is(err, ErrTooLarge), errors.Is(err, model.ErrInputSizeLimit):
			return nil, ErrTooLarge
		case errors.Is(err, context.DeadlineExceeded):
			return nil, errors.New("This PDF took too long to process. Try a smaller document.")
		default:
			return nil, ErrInvalid
		}
	}
	return output.Bytes(), nil
}

// Save writes a private temporary file beside the destination and commits it only
// after a complete write. Source files, symlinks and aliases cannot be overwritten.
func Save(source, destination string, data []byte) error {
	if !strings.EqualFold(filepath.Ext(destination), ".pdf") {
		return errors.New("Save the unlocked document with a .pdf extension.")
	}
	src, err := os.Stat(source)
	if err != nil {
		return errors.New("The original PDF is no longer available. Select it again.")
	}
	sourcePath, err := filepath.EvalSymlinks(source)
	if err != nil {
		return ErrInvalid
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(destination))
	if err != nil {
		return errors.New("The save folder is unavailable. Choose another location.")
	}
	destination = filepath.Join(parent, filepath.Base(destination))
	sourcePath, _ = filepath.Abs(sourcePath)
	destination, _ = filepath.Abs(destination)
	if sourcePath == destination {
		return errors.New("Choose a different name or location to keep your original PDF.")
	}
	if existing, err := os.Lstat(destination); err == nil {
		if os.SameFile(src, existing) {
			return errors.New("Choose a different name or location to keep your original PDF.")
		}
		if !existing.Mode().IsRegular() {
			return errors.New("Choose a regular file location. Symbolic links and folders cannot be replaced.")
		}
	} else if !os.IsNotExist(err) {
		return errors.New("Could not access the save location. Choose another location.")
	}
	f, err := os.CreateTemp(parent, ".pdf-tools-*.tmp")
	if err != nil {
		return errors.New("Could not write to this folder. Choose a location with write permission.")
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	if err == nil {
		err = f.Close()
	}
	if err == nil {
		err = os.Rename(f.Name(), destination)
	}
	if err != nil {
		return errors.New("Could not save the unlocked PDF. Check available disk space and folder permissions, then try again.")
	}
	return nil
}
