package fileutil

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func zipFile(filePath string, destPath string) error {
	zipFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	return addFileToArchive1(filePath, archive)
}

func zipFolder(folderPath string, destPath string) error {
	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)

	err = addFileToArchive2(w, folderPath, "")
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return nil
}

func addFileToArchive1(fpath string, archive *zip.Writer) error {
	err := filepath.Walk(fpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, filepath.Dir(fpath)+"/")

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(writer, file); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func addFileToArchive2(w *zip.Writer, basePath, baseInZip string) error {
	files, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(basePath, "/") {
		basePath = basePath + "/"
	}

	for _, file := range files {
		if !file.IsDir() {
			dat, err := os.ReadFile(basePath + file.Name())
			if err != nil {
				return err
			}

			f, err := w.Create(baseInZip + file.Name())
			if err != nil {
				return err
			}
			_, err = f.Write(dat)
			if err != nil {
				return err
			}
		} else if file.IsDir() {
			newBase := basePath + file.Name() + "/"
			addFileToArchive2(w, newBase, baseInZip+file.Name()+"/")
		}
	}
	return nil
}

func safeFilepathJoin(path1, path2 string) (string, error) {
	relPath, err := filepath.Rel(".", path2)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("(zipslip) filepath is unsafe %q: %v", path2, err)
	}
	if path1 == "" {
		path1 = "."
	}
	return filepath.Join(path1, filepath.Join("/", relPath)), nil
}

func escapeCSVField(field string, delimiter rune) string {
	// 替换所有的双引号为两个双引号
	escapedField := strings.ReplaceAll(field, "\"", "\"\"")
	// 如果字段包含分隔符、双引号或换行符，用双引号包裹整个字段
	if strings.ContainsAny(escapedField, string(delimiter)+"\"\n") {
		escapedField = fmt.Sprintf("\"%s\"", escapedField)
	}
	return escapedField
}

func isCsvSupportedType(v interface{}) bool {
	switch v.(type) {
	case bool, rune, string, int, int64, float32, float64, uint, byte, complex128, complex64, uintptr:
		return true
	default:
		return false
	}
}
