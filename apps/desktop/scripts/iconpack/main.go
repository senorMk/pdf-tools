// iconpack wraps the generated PNG sizes in an ICNS container without requiring
// macOS image services, which may be unavailable in restricted build sessions.
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: iconpack iconset-directory output.icns")
		os.Exit(2)
	}
	entries := []struct{ kind, file string }{
		{"icp4", "icon_16x16.png"}, {"icp5", "icon_32x32.png"},
		{"icp6", "icon_32x32@2x.png"}, {"ic07", "icon_128x128.png"},
		{"ic08", "icon_256x256.png"}, {"ic09", "icon_512x512.png"},
		{"ic10", "icon_512x512@2x.png"}, {"ic11", "icon_16x16@2x.png"},
		{"ic12", "icon_32x32@2x.png"}, {"ic13", "icon_128x128@2x.png"},
		{"ic14", "icon_256x256@2x.png"},
	}
	var chunks bytes.Buffer
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(os.Args[1], entry.file))
		if err != nil {
			fail(err)
		}
		chunks.WriteString(entry.kind)
		binary.Write(&chunks, binary.BigEndian, uint32(len(data)+8))
		chunks.Write(data)
	}
	var result bytes.Buffer
	result.WriteString("icns")
	binary.Write(&result, binary.BigEndian, uint32(chunks.Len()+8))
	result.Write(chunks.Bytes())
	if err := os.WriteFile(os.Args[2], result.Bytes(), 0644); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "Could not package application icon:", err)
	os.Exit(1)
}
