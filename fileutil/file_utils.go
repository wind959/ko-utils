package fileutil

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/wind959/ko-utils/validator"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type FileReader struct {
	*bufio.Reader
	file   *os.File
	offset int64
}

func NewFileReader(path string) (*FileReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &FileReader{
		file:   f,
		Reader: bufio.NewReader(f),
		offset: 0,
	}, nil
}

func (f *FileReader) ReadLine() (string, error) {
	data, err := f.Reader.ReadBytes('\n')
	f.offset += int64(len(data))
	if err == nil || err == io.EOF {
		for len(data) > 0 && (data[len(data)-1] == '\r' || data[len(data)-1] == '\n') {
			data = data[:len(data)-1]
		}
		return string(data), err
	}
	return "", err
}

func (f *FileReader) Offset() int64 {
	return f.offset
}

func (f *FileReader) SeekOffset(offset int64) error {
	_, err := f.file.Seek(offset, 0)
	if err != nil {
		return err
	}
	f.Reader = bufio.NewReader(f.file)
	f.offset = offset
	return nil
}

func (f *FileReader) Close() error {
	return f.file.Close()
}

// IsExist 判断文件或目录是否存在
func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}

// CreateFile 创建文件，创建成功返回true, 否则返回false
func CreateFile(path string) bool {
	file, err := os.Create(path)
	if err != nil {
		return false
	}

	defer file.Close()
	return true
}

// CreateDir 使用绝对路径创建嵌套目录，例如/a/, /a/b
func CreateDir(absPath string) error {
	return os.MkdirAll(absPath, os.ModePerm)
}

// CopyDir 拷贝文件夹到目标路径，会递归复制文件夹下所有的文件及文件夹，
// 并且访问权限也与源文件夹保持一致。
// 当dstPath存在时会返回error
func CopyDir(srcPath string, dstPath string) error {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get source directory info: %w", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source path is not a directory: %s", srcPath)
	}

	err = os.MkdirAll(dstPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	entries, err := os.ReadDir(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcDir := filepath.Join(srcPath, entry.Name())
		dstDir := filepath.Join(dstPath, entry.Name())

		if entry.IsDir() {
			err := CopyDir(srcDir, dstDir)
			if err != nil {
				return err
			}
		} else {
			err := CopyFile(srcDir, dstDir)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// IsDir 判断参数是否是目录
func IsDir(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}
	return file.IsDir()
}

// RemoveFile 删除文件
func RemoveFile(path string, onDelete ...func(path string)) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}
	if len(onDelete) > 0 && onDelete[0] != nil {
		onDelete[0](path)
	}
	return os.Remove(path)
}

func RemoveDir(path string, onDelete ...func(path string)) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	var callback func(string)
	if len(onDelete) > 0 {
		callback = onDelete[0]
	}
	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err == nil && callback != nil {
			callback(p)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return os.RemoveAll(path)
}

// CopyFile 拷贝文件，会覆盖原有的文件
func CopyFile(srcPath string, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	distFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer distFile.Close()

	var tmp = make([]byte, 1024*4)
	for {
		n, err := srcFile.Read(tmp)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		_, err = distFile.Write(tmp[:n])
		if err != nil {
			return err
		}
	}
}

// ClearFile 清空文件内容
func ClearFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

// ReadFileToString 读取文件内容并返回字符串
func ReadFileToString(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ReadFileByLine 按行读取文件内容，返回字符串切片包含每一行
func ReadFileByLine(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make([]string, 0)
	buf := bufio.NewReader(f)

	for {
		line, _, err := buf.ReadLine()
		l := string(line)
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		result = append(result, l)
	}

	return result, nil
}

// ListFileNames 返回目录下所有文件名
func ListFileNames(path string) ([]string, error) {
	if !IsExist(path) {
		return []string{}, nil
	}

	fs, err := os.ReadDir(path)
	if err != nil {
		return []string{}, err
	}

	sz := len(fs)
	if sz == 0 {
		return []string{}, nil
	}

	result := []string{}
	for i := 0; i < sz; i++ {
		if !fs[i].IsDir() {
			result = append(result, fs[i].Name())
		}
	}

	return result, nil
}

// IsZipFile 判断文件是否是zip压缩文件
func IsZipFile(filepath string) bool {
	f, err := os.Open(filepath)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 4)
	if n, err := f.Read(buf); err != nil || n < 4 {
		return false
	}

	return bytes.Equal(buf, []byte("PK\x03\x04"))
}

// Zip  zip压缩文件, fpath参数可以是文件或目录
func Zip(path string, destPath string) error {
	if IsDir(path) {
		return zipFolder(path, destPath)
	}

	return zipFile(path, destPath)
}

// UnZip zip解压缩文件并保存在目录中
func UnZip(zipFile string, destPath string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		decodeName := f.Name
		if f.Flags == 0 {
			i := bytes.NewReader([]byte(f.Name))
			decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
			content, _ := io.ReadAll(decoder)
			decodeName = string(content)
		}
		// issue#62: fix ZipSlip bug
		path, err := safeFilepathJoin(destPath, decodeName)
		if err != nil {
			return err
		}
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				return err
			}
			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()
			outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ZipAppendEntry 通过将单个文件或目录追加到现有的zip文件
func ZipAppendEntry(fpath string, destPath string) error {
	tempFile, err := os.CreateTemp("", "temp.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	zipReader, err := zip.OpenReader(destPath)
	if err != nil {
		return err
	}
	archive := zip.NewWriter(tempFile)
	for _, zipItem := range zipReader.File {
		zipItemReader, err := zipItem.Open()
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(zipItem.FileInfo())
		if err != nil {
			return err
		}
		header.Name = zipItem.Name
		targetItem, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(targetItem, zipItemReader)
		if err != nil {
			return err
		}
	}
	err = addFileToArchive1(fpath, archive)
	if err != nil {
		return err
	}
	err = zipReader.Close()
	if err != nil {
		return err
	}
	err = archive.Close()
	if err != nil {
		return err
	}
	err = tempFile.Close()
	if err != nil {
		return err
	}
	return CopyFile(tempFile.Name(), destPath)
}

// IsLink 判断文件是否是符号链接
func IsLink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// FileMode 获取文件mode信息
func FileMode(path string) (fs.FileMode, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	return fi.Mode(), nil
}

// MiMeType 获取文件mime类型, 'file'参数的类型必须是string或者*os.File
func MiMeType(file any) string {
	var mediatype string

	readBuffer := func(f *os.File) ([]byte, error) {
		buffer := make([]byte, 512)
		_, err := f.Read(buffer)
		if err != nil {
			return nil, err
		}
		return buffer, nil
	}

	if filePath, ok := file.(string); ok {
		f, err := os.Open(filePath)
		if err != nil {
			return mediatype
		}
		buffer, err := readBuffer(f)
		if err != nil {
			return mediatype
		}
		return http.DetectContentType(buffer)
	}

	if f, ok := file.(*os.File); ok {
		buffer, err := readBuffer(f)
		if err != nil {
			return mediatype
		}
		return http.DetectContentType(buffer)
	}
	return mediatype
}

// CurrentPath 返回当前位置的绝对路径
func CurrentPath() string {
	var absPath string
	_, filename, _, ok := runtime.Caller(1)
	if ok {
		absPath = filepath.Dir(filename)
	}

	return absPath
}

// FileSize 返回文件字节大小
func FileSize(path string) (int64, error) {
	f, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return f.Size(), nil
}

func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			size += info.Size()
		}
		return err
	})
	return size, err
}

// MTime 返回文件修改时间(unix timestamp)
func MTime(filepath string) (int64, error) {
	f, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}
	return f.ModTime().Unix(), nil
}

// Sha  返回文件sha值，参数`shaType` 应传值为: 1, 256，512.
func Sha(filepath string, shaType ...int) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	h := sha1.New()
	if len(shaType) > 0 {
		if shaType[0] == 1 {
			h = sha1.New()
		} else if shaType[0] == 256 {
			h = sha256.New()
		} else if shaType[0] == 512 {
			h = sha512.New()
		} else {
			return "", errors.New("param `shaType` should be 1, 256 or 512")
		}
	}
	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}
	sha := fmt.Sprintf("%x", h.Sum(nil))
	return sha, nil
}

// ReadCsvFile 读取csv文件内容到切片
func ReadCsvFile(filepath string, delimiter ...rune) ([][]string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := csv.NewReader(f)
	if len(delimiter) > 0 {
		reader.Comma = delimiter[0]
	}
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

// WriteCsvFile 向csv文件写入内容
func WriteCsvFile(filepath string, records [][]string, append bool, delimiter ...rune) error {
	flag := os.O_RDWR | os.O_CREATE
	if append {
		flag = flag | os.O_APPEND
	}
	f, err := os.OpenFile(filepath, flag, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := csv.NewWriter(f)
	// 设置默认分隔符为逗号，除非另外指定
	if len(delimiter) > 0 {
		writer.Comma = delimiter[0]
	} else {
		writer.Comma = ','
	}
	// 遍历所有记录并处理包含分隔符或双引号的单元格
	for i := range records {
		for j := range records[i] {
			records[i][j] = escapeCSVField(records[i][j], writer.Comma)
		}
	}
	return writer.WriteAll(records)
}

// WriteStringToFile 将字符串写入文件
func WriteStringToFile(filepath string, content string, append bool) error {
	var flag int
	if append {
		flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	} else {
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	}
	f, err := os.OpenFile(filepath, flag, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

// WriteBytesToFile 将bytes写入文件
func WriteBytesToFile(filepath string, content []byte) error {
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(content)
	return err
}

// ReadFile 读取文件或者URL
func ReadFile(path string) (reader io.ReadCloser, closeFn func(), err error) {
	switch {
	case validator.IsUrl(path):
		resp, err := http.Get(path)
		if err != nil {
			return nil, func() {}, err
		}
		return resp.Body, func() { resp.Body.Close() }, nil
	case IsExist(path):
		reader, err := os.Open(path)
		if err != nil {
			return nil, func() {}, err
		}
		return reader, func() { reader.Close() }, nil
	default:
		return nil, func() {}, errors.New("unknown file type")
	}
}

// WriteMapsToCsv  将map切片写入csv文件中
func WriteMapsToCsv(filepath string, records []map[string]any, appendToExistingFile bool, delimiter rune,
	headers ...[]string) error {
	for _, record := range records {
		for _, value := range record {
			if !isCsvSupportedType(value) {
				return errors.New("unsupported value type detected; only basic types are supported: \nbool, rune, string, int, int64, float32, float64, uint, byte, complex128, complex64, uintptr")
			}
		}
	}

	var columnHeaders []string
	if len(headers) > 0 {
		columnHeaders = headers[0]
	} else {
		columnHeaders = make([]string, 0, len(records[0]))
		for key := range records[0] {
			columnHeaders = append(columnHeaders, key)
		}
		// sort keys in alphabeta order
		sort.Strings(columnHeaders)
	}
	var datasToWrite [][]string
	if !appendToExistingFile {
		datasToWrite = append(datasToWrite, columnHeaders)
	}
	for _, record := range records {
		row := make([]string, 0, len(columnHeaders))
		for _, h := range columnHeaders {
			row = append(row, fmt.Sprintf("%v", record[h]))
		}
		datasToWrite = append(datasToWrite, row)
	}
	return WriteCsvFile(filepath, datasToWrite, appendToExistingFile, delimiter)
}

// ChunkRead 从文件的指定偏移读取块并返回块内所有行
func ChunkRead(file *os.File, offset int64, size int, bufPool *sync.Pool) ([]string, error) {
	buf := bufPool.Get().([]byte)[:size] // 从Pool获取缓冲区并调整大小
	n, err := file.ReadAt(buf, offset)   // 从指定偏移读取数据到缓冲区
	if err != nil && err != io.EOF {
		return nil, err
	}
	buf = buf[:n] // 调整切片以匹配实际读取的字节数

	var lines []string
	var lineStart int
	for i, b := range buf {
		if b == '\n' {
			line := string(buf[lineStart:i]) // 不包括换行符
			lines = append(lines, line)
			lineStart = i + 1 // 设置下一行的开始
		}
	}

	if lineStart < len(buf) { // 处理块末尾的行
		line := string(buf[lineStart:])
		lines = append(lines, line)
	}
	bufPool.Put(buf) // 读取完成后，将缓冲区放回Pool
	return lines, nil
}

// ParallelChunkRead 并行读取文件并将每个块的行发送到指定通道
func ParallelChunkRead(filePath string, linesCh chan<- []string, chunkSizeMB, maxGoroutine int) error {
	if chunkSizeMB == 0 {
		chunkSizeMB = 100
	}
	chunkSize := chunkSizeMB * 1024 * 1024
	// 内存复用
	bufPool := sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, chunkSize)
		},
	}
	if maxGoroutine == 0 {
		maxGoroutine = runtime.NumCPU() // 设置为0时使用CPU核心数
	}
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	chunkOffsetCh := make(chan int64, maxGoroutine)

	// 分配工作
	go func() {
		for i := int64(0); i < info.Size(); i += int64(chunkSize) {
			chunkOffsetCh <- i
		}
		close(chunkOffsetCh)
	}()

	// 启动工作协程
	for i := 0; i < maxGoroutine; i++ {
		wg.Add(1)
		go func() {
			for chunkOffset := range chunkOffsetCh {
				chunk, err := ChunkRead(f, chunkOffset, chunkSize, &bufPool)
				if err == nil {
					linesCh <- chunk
				}
			}
			wg.Done()
		}()
	}
	// 等待所有解析完成后关闭行通道
	wg.Wait()
	close(linesCh)
	return nil
}

// SaveToFile 将数据保存到文件，支持自定义格式化函数
func SaveToFile(filename string, data interface{}, formatFunc func(interface{}, io.Writer) error) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	if err := formatFunc(data, file); err != nil {
		return err
	}
	return nil
}

// FileExt 获取文件扩展名
func FileExt(filename string) string {
	dotIndex := strings.LastIndex(filename, ".")
	if dotIndex == -1 {
		return ""
	}
	return filename[dotIndex:]
}
