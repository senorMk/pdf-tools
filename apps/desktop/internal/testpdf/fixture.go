// Package testpdf generates synthetic PDFs for tests. It never reads user files.
package testpdf

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func Plain() []byte {
	var b bytes.Buffer
	b.WriteString("%PDF-1.7\n")
	stream := "BT /F1 18 Tf 40 100 Td (Private test document) Tj ET\n"
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R /Annots [6 0 R] >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(stream), stream),
		"<< /Type /Annot /Subtype /Link /Rect [40 80 160 120] /Border [0 0 0] /A << /S /URI /URI (https://privacy-check.invalid/document) >> >>",
	}
	offsets := []int{0}
	for i, object := range objects {
		offsets = append(offsets, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), xref)
	return b.Bytes()
}

func Encrypted(t *testing.T, user, owner string, aes bool, bits int) string {
	t.Helper()
	conf := model.NewAESConfiguration(user, owner, bits)
	conf.EncryptUsingAES = aes
	var encrypted bytes.Buffer
	if err := api.Encrypt(context.Background(), bytes.NewReader(Plain()), &encrypted, conf); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "protected.pdf")
	if err := os.WriteFile(path, encrypted.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
