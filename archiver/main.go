package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

//
// =========================
// CLI PATH HANDLING
// =========================
//

func normalizeInputPath(p string) string {
	p = strings.TrimSpace(p)

	if len(p) >= 2 {
		if (p[0] == '"' && p[len(p)-1] == '"') ||
			(p[0] == '\'' && p[len(p)-1] == '\'') {
			p = p[1 : len(p)-1]
		}
	}
	return p
}

func cleanCLIPath(p string) string {
	return normalizeInputPath(p)
}

//
// =========================
// OUTPUT LOGIC
// =========================
//

func sameDir(a, b string) bool {
	aa, err1 := filepath.Abs(a)
	bb, err2 := filepath.Abs(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return filepath.Clean(aa) == filepath.Clean(bb)
}

func ensureParentDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}

func resolveOutputPath(sourceDir, target string) (string, error) {
	target = cleanCLIPath(target)

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// nur Dateiname → working directory
	if filepath.Dir(target) == "." && !strings.Contains(target, string(os.PathSeparator)) {
		target = filepath.Join(wd, target)
	}

	// Konfliktregel:
	// wenn source == working dir → nicht dort schreiben
	if sameDir(sourceDir, wd) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		fallback := filepath.Join(home, "Documents", "archiver", filepath.Base(target))
		return fallback, nil
	}

	return target, nil
}

//
// =========================
// SAFE PATH CHECK
// =========================
//

func safeJoin(base, target string) (string, error) {
	base = filepath.Clean(base)
	target = filepath.Clean(filepath.Join(base, target))

	if !strings.HasPrefix(target, base+string(os.PathSeparator)) && target != base {
		return "", fmt.Errorf("path traversal blocked: %s", target)
	}
	return target, nil
}

//
// =========================
// ARCHIVER INTERFACE
// =========================
//

type Archiver interface {
	Create(ctx context.Context, src, dst string) error
	Extract(ctx context.Context, src, dst string) error
}

var registry = map[string]Archiver{}

func register(name string, a Archiver) {
	registry[name] = a
}

func get(name string) (Archiver, error) {
	a, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown format: %s", name)
	}
	return a, nil
}

//
// =========================
// TAR CORE
// =========================
//

func writeTar(ctx context.Context, w io.Writer, source string) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	buf := make([]byte, 64*1024)

	return filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			return err
		}

		if path == source {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		hdr.Name = filepath.ToSlash(rel)

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		_, copyErr := io.CopyBuffer(tw, f, buf)
		f.Close()

		if ctx.Err() != nil {
			return ctx.Err()
		}

		return copyErr
	})
}

func readTar(ctx context.Context, r io.Reader, dst string) error {
	tr := tar.NewReader(r)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target, err := safeJoin(dst, hdr.Name)
		if err != nil {
			return err
		}

		if hdr.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		os.MkdirAll(filepath.Dir(target), 0755)

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, hdr.FileInfo().Mode())
		if err != nil {
			return err
		}

		_, copyErr := io.Copy(out, tr)
		out.Close()

		if copyErr != nil {
			return copyErr
		}
	}

	return nil
}

//
// =========================
// TAR
// =========================
//

type TarArchiver struct{}

func (TarArchiver) Create(ctx context.Context, src, dst string) error {
	if err := ensureParentDir(dst); err != nil {
		return err
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	return writeTar(ctx, f, src)
}

func (TarArchiver) Extract(ctx context.Context, src, dst string) error {
	if err := ensureParentDir(dst); err != nil {
		return err
	}

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	return readTar(ctx, f, dst)
}

//
// =========================
// TGZ
// =========================
//

type TgzArchiver struct{}

func (TgzArchiver) Create(ctx context.Context, src, dst string) error {
	if err := ensureParentDir(dst); err != nil {
		return err
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	return writeTar(ctx, gz, src)
}

func (TgzArchiver) Extract(ctx context.Context, src, dst string) error {
	if err := ensureParentDir(dst); err != nil {
		return err
	}

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	return readTar(ctx, gz, dst)
}

//
// =========================
// ZIP
// =========================
//

type ZipArchiver struct{}

func (ZipArchiver) Create(ctx context.Context, src, dst string) error {
	if err := ensureParentDir(dst); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	buf := make([]byte, 64*1024)

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			return err
		}

		if path == src {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if d.IsDir() {
			_, err := zw.Create(filepath.ToSlash(rel) + "/")
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			f.Close()
			return err
		}

		header.Name = filepath.ToSlash(rel)

		w, err := zw.CreateHeader(header)
		if err != nil {
			f.Close()
			return err
		}

		_, copyErr := io.CopyBuffer(w, f, buf)
		f.Close()

		return copyErr
	})
}

func (ZipArchiver) Extract(ctx context.Context, src, dst string) error {
	if err := ensureParentDir(dst); err != nil {
		return err
	}

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		target, err := safeJoin(dst, f.Name)
		if err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		os.MkdirAll(filepath.Dir(target), 0755)

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			out.Close()
			return err
		}

		_, copyErr := io.Copy(out, rc)
		out.Close()
		rc.Close()

		if copyErr != nil {
			return copyErr
		}
	}

	return nil
}

func archiveBaseName(path string) string {
	name := filepath.Base(path)
	lower := strings.ToLower(name)

	switch {
	case strings.HasSuffix(lower, ".tar.gz"):
		return name[:len(name)-7]

	case strings.HasSuffix(lower, ".tgz"):
		return name[:len(name)-4]

	case strings.HasSuffix(lower, ".tar"):
		return name[:len(name)-4]

	case strings.HasSuffix(lower, ".zip"):
		return name[:len(name)-4]

	default:
		return strings.TrimSuffix(name, filepath.Ext(name))
	}
}

//
// =========================
// HELP
// =========================
//

func help() {
	fmt.Println(`
Usage:
  archiver create <zip|tar|tgz> <source> <target>
  archiver extract <zip|tar|tgz> <source> <target>

Rules:
  - default output = working directory
  - if working dir == source dir → fallback to Documents/archiver

Examples:
  archiver create zip c:\data backup.zip
  archiver extract tgz backup.tgz c:\out  
`)
}

//
// =========================
// MAIN (Ctrl+C SAFE)
// =========================
//

func main() {
	register("zip", ZipArchiver{})
	register("tar", TarArchiver{})
	register("tgz", TgzArchiver{})

	if len(os.Args) < 2 {
		help()
		return
	}

	switch os.Args[1] {
	case "-h", "--help", "help":
		help()
		return
	}

	if len(os.Args) < 5 {
		help()
		return
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	action := strings.ToLower(os.Args[1])
	format := strings.ToLower(os.Args[2])

	source := cleanCLIPath(os.Args[3])
	target := cleanCLIPath(os.Args[4])

	arch, err := get(format)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	var dst string

	switch action {
	case "create":
		dst, err = resolveOutputPath(source, target)
		if err != nil {
			fmt.Println("error:", err)
			return
		}

	case "extract":
		dst = filepath.Join(
			target,
			archiveBaseName(source),
		)

	default:
		help()
		return
	}

	switch action {
	case "create":
		err = arch.Create(ctx, source, dst)

	case "extract":
		err = arch.Extract(ctx, source, dst)
	}

	if err != nil {
		fmt.Println("error:", err)
		return
	}

	switch action {
	case "create":
		fmt.Println("created:", dst)

	case "extract":
		fmt.Println("extracted to:", dst)
	}
}
