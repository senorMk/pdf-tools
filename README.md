# PDF Tools

A small, privacy first desktop PDF toolkit. The first tool removes a PDF password you already know and saves an unlocked copy wherever you choose.

The project contains a Go/Wails desktop app in `apps/desktop`, a Svelte/TypeScript frontend, independent Go services under `internal`, and root Make/npm commands.

## Run on macOS

Requires Go 1.26+, Node 22.12+ (or a supported newer version), and Xcode Command Line Tools.

```sh
npm ci
make check
make test
make prototype
open "apps/desktop/build/bin/PDF Tools.app"
```

The app bundle includes the frontend, so it runs without a development server or internet connection. `make prototype` signs the bundle ad hoc for local development. It is not notarized for distribution. Set `PDFTOOLS_SIGN_IDENTITY` to use an installed signing identity instead.

Select **About** in the navigation area to see the app details, version, developer name, and license. The About page also describes how the app keeps file processing on your computer.

## Remove a password

1. Select a PDF with **browse files**, or drag one PDF onto the drop area.
2. Enter its known opening password or owner password. Password whitespace is preserved.
3. Choose **Remove password & save**. The PDF is decrypted locally before the native save dialog opens.
4. Choose a destination. The default name is `original-unlocked.pdf`.

The original is preserved, including when a chosen destination is a hard link or symbolic link to it. Save cancellation creates no output. Incorrect passwords, invalid PDFs, unsupported encryption, and save failures show a message in the interface. Already unlocked PDFs do not need this operation. This tool does not guess or recover unknown passwords.

Input and output PDFs are limited to 100 MiB each. Certificate-based encryption is unsupported. Removing encryption rewrites the document; existing digital signatures may no longer validate.

## Privacy and file handling

- PDF processing uses the Go [pdfcpu library](https://github.com/pdfcpu/pdfcpu) inside the desktop process. There are no uploads, remote processing endpoints, telemetry, account setup, external fonts, or runtime network requests in application code.
- PDF contents remain in Go. The frontend receives only file metadata and save results. Passwords cross the local native bridge and are not written to settings, logs, command arguments, or browser storage. The password field clears on submission.
- Decrypted output stays in memory until a save location is chosen. Cancelling clears that buffer. Successful saving writes a temporary file with private permissions in the destination folder, then renames it into place. Failed writes remove the temporary file and preserve the existing destination. The temporary file is also removed after successful saving; an abnormal process termination can leave it in that folder.
- pdfcpu's configuration and font caches are disabled, and its offline mode is explicitly enabled. The app has no document history or persistent preferences.
- The frontend embeds local assets and uses a content security policy that restricts resources and connections to its own origin.

Memory clearing is best effort: Go and JavaScript runtimes can retain earlier copies until garbage collection, and operating system swap/crash dumps are outside the app's control. Dependency installation requires internet access on a clean machine; working with PDFs does not.

## Development

```text
apps/desktop/
  main.go                 Wails window and lifecycle
  desktop.go              Native selection, unlock, and save flow
  internal/pdf/           PDF validation, decryption, safe saving
  internal/testpdf/       Synthetic PDF fixtures for tests
  frontend/src/           Svelte UI and typed native bridge
  scripts/package-macos.sh
  build/Info.plist
  build/appicon.svg        Editable vector icon (PNG rendering included)
```

`npm run preview` runs the interface in a browser at the local Vite address. Native selection, drag-and-drop file paths, and processing require the desktop app. `make build` produces a desktop executable; `make prototype` produces the macOS app bundle. Wails is pinned to v2.15.0; a global CLI installation is unnecessary for these commands.

`make check` builds the frontend, checks Svelte/TypeScript, and runs Go vet. `make test` runs Go tests with the race detector against generated PDFs, covering AES-128/AES-256 and RC4-40/RC4-128, user and owner passwords, incorrect passwords, malformed and oversized inputs, unlocked output validation, original preservation, save aliases, output permissions, selection cancellation, save cancellation, overlapping operations, Unicode passwords, and decryption with embedded links while HTTP requests are blocked.

The implementation uses Wails' documented [native drag-and-drop API](https://wails.io/docs/reference/runtime/draganddrop/) and pdfcpu's [decryption operation](https://pdfcpu.io/encrypt/decryptpdf/). To change the icon, edit `apps/desktop/build/appicon.svg` and regenerate its PNG with `resvg apps/desktop/build/appicon.svg apps/desktop/build/appicon.png` before packaging. The small Go icon packer uses the PNG sizes produced by macOS `sips`.

Packaging is currently provided for macOS. Windows/Linux native builds and real native dialog/drop interactions require platform testing.

## License

Licensed under the [MIT License](LICENSE).
