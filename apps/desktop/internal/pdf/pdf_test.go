package pdf

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/senorMk/pdf-tools/apps/desktop/internal/testpdf"
)

func TestDecryptKnownPasswords(t *testing.T) {
	for _, tc := range []struct {
		name string
		aes  bool
		bits int
	}{
		{"AES256", true, 256}, {"AES128", true, 128}, {"RC4128", false, 128}, {"RC440", false, 40},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := testpdf.Encrypted(t, " user password ", "owner-password", tc.aes, tc.bits)
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, password := range []string{" user password ", "owner-password"} {
				data, err := Decrypt(path, password)
				if err != nil {
					t.Fatalf("decrypt: %v", err)
				}
				ctx, err := api.ReadValidateAndOptimize(context.Background(), bytes.NewReader(data), model.NewDefaultConfiguration(), nil)
				if err != nil {
					t.Fatalf("open without password: %v", err)
				}
				if ctx.Encrypt != nil {
					t.Fatal("output remains encrypted")
				}
				pages, err := api.PageCount(context.Background(), bytes.NewReader(data), model.NewDefaultConfiguration())
				if err != nil || pages != 1 {
					t.Fatalf("page count %d, error %v", pages, err)
				}
				content, err := pdfcpu.ExtractPageContent(context.Background(), ctx, 1)
				if err != nil {
					t.Fatal(err)
				}
				text, err := io.ReadAll(content)
				if err != nil || !bytes.Contains(text, []byte("Private test document")) {
					t.Fatalf("page content lost: %s, %v", text, err)
				}
				clear(data)
			}
			if _, err := Decrypt(path, "wrong-password"); !errors.Is(err, ErrPassword) {
				t.Fatalf("wrong password: %v", err)
			}
			after, _ := os.ReadFile(path)
			if !bytes.Equal(original, after) {
				t.Fatal("source was modified")
			}
		})
	}
}

func TestRejectUnprotectedMalformedAndOversize(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.pdf")
	if err := os.WriteFile(path, testpdf.Plain(), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Decrypt(path, "anything"); !errors.Is(err, ErrUnprotected) {
		t.Fatalf("unprotected: %v", err)
	}
	if _, err := Decrypt(path, ""); err == nil {
		t.Fatal("empty password accepted")
	}
	if err := os.WriteFile(path, []byte("%PDF-1.7\nbroken document"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Decrypt(path, "secret"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("malformed: %v", err)
	}
	if err := os.WriteFile(path, []byte("not a pdf"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(path); err == nil {
		t.Fatal("non-PDF accepted")
	}
	if _, err := Inspect(filepath.Dir(path)); err == nil {
		t.Fatal("directory accepted")
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(MaxFileSize + 1); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, err := Inspect(path); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("oversize: %v", err)
	}
}

func TestSavePreservesSourceAndRejectsAliases(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.pdf")
	if err := os.WriteFile(source, testpdf.Plain(), 0600); err != nil {
		t.Fatal(err)
	}
	hardlink, symlink := filepath.Join(dir, "hardlink.pdf"), filepath.Join(dir, "symlink.pdf")
	if err := os.Link(source, hardlink); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, symlink); err != nil {
		t.Fatal(err)
	}
	for _, destination := range []string{source, hardlink, symlink, filepath.Join(dir, "output.txt")} {
		if err := Save(source, destination, []byte("new data")); err == nil {
			t.Fatalf("unsafe destination accepted: %s", destination)
		}
	}
	parentAlias := filepath.Join(dir, "parent-alias")
	if err := os.Symlink(dir, parentAlias); err != nil {
		t.Fatal(err)
	}
	if err := Save(source, filepath.Join(parentAlias, "source.pdf"), []byte("bad")); err == nil {
		t.Fatal("parent alias accepted")
	}
	original, _ := os.ReadFile(source)
	if !bytes.Equal(original, testpdf.Plain()) {
		t.Fatal("original changed")
	}
	destination := filepath.Join(dir, "unlocked.pdf")
	for _, data := range [][]byte{testpdf.Plain(), []byte("replacement")} {
		if err := Save(source, destination, data); err != nil {
			t.Fatal(err)
		}
		saved, _ := os.ReadFile(destination)
		if !bytes.Equal(saved, data) {
			t.Fatal("saved data differs")
		}
	}
	info, _ := os.Stat(destination)
	if info.Mode().Perm() != 0600 {
		t.Fatalf("permissions: %o", info.Mode().Perm())
	}
	temp, _ := filepath.Glob(filepath.Join(dir, ".pdf-tools-*.tmp"))
	if len(temp) != 0 {
		t.Fatalf("temporary files left: %v", temp)
	}
	if err := Save(source, filepath.Join(dir, "missing", "out.pdf"), testpdf.Plain()); err == nil {
		t.Fatal("missing folder accepted")
	}
}

func TestOutputSizeIsBounded(t *testing.T) {
	b := &limitedBuffer{}
	if _, err := b.Write(make([]byte, MaxFileSize)); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Write([]byte{1}); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("output limit: %v", err)
	}
}

type blockedTransport struct{ calls atomic.Int32 }

func (tr *blockedTransport) RoundTrip(*http.Request) (*http.Response, error) {
	tr.calls.Add(1)
	return nil, errors.New("test: all network requests are blocked")
}

func TestDecryptStaysOfflineWithLinksAndUnicodePasswords(t *testing.T) {
	transport := &blockedTransport{}
	originalTransport := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = originalTransport })
	source := testpdf.Encrypted(t, "päss word", "主人密码", true, 256)
	for _, password := range []string{"päss word", "主人密码"} {
		data, err := Decrypt(source, password)
		if err != nil {
			t.Fatal(err)
		}
		clear(data)
	}
	if calls := transport.calls.Load(); calls != 0 {
		t.Fatalf("decryption attempted %d network requests", calls)
	}
}
