package excelutil

import (
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
)

// 错误定义
var (
	ErrFileNotFound = errors.New("file not found")
	ErrEmptySheet   = errors.New("empty sheet")
	ErrInvalidData  = errors.New("invalid data")
)

// Excel 工具结构体
type Excel struct {
	file *excelize.File
	mu   sync.RWMutex
}

// ReadOption 读取配置
type ReadOption struct {
	Sheet     string // 工作表名
	HasHeader bool   // 是否有表头
	StartRow  int    // 起始行
	StartCol  int    // 起始列
	MaxRows   int    // 最大行数
}

// WriteOption 写入配置
type WriteOption struct {
	Sheet    string   // 工作表名
	Headers  []string // 表头
	StartRow int      // 起始行
	StartCol int      // 起始列
}

// NewExcel 创建新的Excel文件
func NewExcel() *Excel {
	return &Excel{
		file: excelize.NewFile(),
	}
}

// Open 打开Excel文件
func Open(filename string) (*Excel, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, ErrFileNotFound
	}

	file, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, err
	}

	return &Excel{file: file}, nil
}

// OpenReader 从reader打开
func OpenReader(r io.Reader) (*Excel, error) {
	file, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	return &Excel{file: file}, nil
}

// Save 保存文件
func (e *Excel) Save() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.file.Save()
}

// SaveAs 另存为
func (e *Excel) SaveAs(filename string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.file.SaveAs(filename)
}

// Close 关闭
func (e *Excel) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.file.Close()
}

// ReadAll 读取所有数据
func (e *Excel) ReadAll(opt ReadOption) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	sheet := opt.Sheet
	if sheet == "" {
		sheets := e.file.GetSheetList()
		if len(sheets) == 0 {
			return nil, ErrEmptySheet
		}
		sheet = sheets[0]
	}

	return e.file.GetRows(sheet)
}

// ReadRows 读取指定范围的数据
func (e *Excel) ReadRows(opt ReadOption) ([][]string, error) {
	allRows, err := e.ReadAll(opt)
	if err != nil {
		return nil, err
	}

	startRow := opt.StartRow
	if startRow < 1 {
		startRow = 1
	}

	if startRow > len(allRows) {
		return [][]string{}, nil
	}

	rows := allRows[startRow-1:]

	if opt.MaxRows > 0 && len(rows) > opt.MaxRows {
		rows = rows[:opt.MaxRows]
	}

	return rows, nil
}

// ReadToSlice 读取到切片
func (e *Excel) ReadToSlice(slicePtr interface{}, opt ReadOption) error {
	rows, err := e.ReadRows(opt)
	if err != nil {
		return err
	}

	sliceVal := reflect.ValueOf(slicePtr)
	if sliceVal.Kind() != reflect.Ptr || sliceVal.Elem().Kind() != reflect.Slice {
		return ErrInvalidData
	}

	elemType := sliceVal.Elem().Type().Elem()
	slice := sliceVal.Elem()

	// 处理表头
	startIdx := 0
	if opt.HasHeader && len(rows) > 0 {
		startIdx = 1
	}

	for i := startIdx; i < len(rows); i++ {
		elem := reflect.New(elemType).Elem()
		if err := e.fillStruct(elem, rows[i]); err != nil {
			return err
		}
		slice = reflect.Append(slice, elem)
	}

	sliceVal.Elem().Set(slice)
	return nil
}

// Write 写入数据
func (e *Excel) Write(data [][]interface{}, opt WriteOption) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	sheet := opt.Sheet
	if sheet == "" {
		sheet = "Sheet1"
	} else {
		_ = e.file.SetSheetName("Sheet1", sheet)
	}
	// 确保工作表存在
	index, err := e.file.GetSheetIndex(sheet)
	if err != nil {
		return err
	}
	if index == -1 {
		// 如果是第一次创建工作表，需要删除默认的Sheet1
		if len(e.file.GetSheetList()) == 1 && e.file.GetSheetList()[0] == "Sheet1" {
			// 删除默认的Sheet1
			err = e.file.DeleteSheet("Sheet1")
			if err != nil {
				return err
			}
		}
		index, err = e.file.NewSheet(sheet)
		if err != nil {
			return err
		}
	}
	startRow := opt.StartRow
	if startRow < 1 {
		startRow = 1
	}
	startCol := opt.StartCol
	if startCol < 1 {
		startCol = 1
	}
	// 写入表头
	if opt.Headers != nil {
		for i, header := range opt.Headers {
			cell, _ := excelize.CoordinatesToCellName(startCol+i, startRow)
			_ = e.file.SetCellValue(sheet, cell, header)
		}
		startRow++
	}
	// 写入数据
	for i, row := range data {
		for j, value := range row {
			cell, _ := excelize.CoordinatesToCellName(startCol+j, startRow+i)
			_ = e.file.SetCellValue(sheet, cell, value)
		}
	}
	// 设置活动工作表
	e.file.SetActiveSheet(index)
	return nil
}

// WriteSlice 写入切片
func (e *Excel) WriteSlice(data interface{}, opt WriteOption) error {
	sliceVal := reflect.ValueOf(data)
	if sliceVal.Kind() != reflect.Slice {
		return ErrInvalidData
	}
	rows := make([][]interface{}, sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i)
		if elem.Kind() == reflect.Struct {
			row := e.structToRow(elem)
			rows[i] = row
		} else {
			rows[i] = []interface{}{elem.Interface()}
		}
	}
	return e.Write(rows, opt)
}

// GetSheetNames 获取所有工作表名
func (e *Excel) GetSheetNames() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.file.GetSheetList()
}

// GetSheetData 获取工作表数据
func (e *Excel) GetSheetData(sheet string) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.file.GetRows(sheet)
}

// SetCellValue 设置单元格值
func (e *Excel) SetCellValue(sheet, axis string, value interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.file.SetCellValue(sheet, axis, value)
}

// GetCellValue 获取单元格值
func (e *Excel) GetCellValue(sheet, axis string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.file.GetCellValue(sheet, axis)
}

// MergeCells 合并单元格
func (e *Excel) MergeCells(sheet, hCell, vCell string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.file.MergeCell(sheet, hCell, vCell)
}

// SetColWidth 设置列宽
func (e *Excel) SetColWidth(sheet, startCol, endCol string, width float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.file.SetColWidth(sheet, startCol, endCol, width)
}

// AddChart 添加图表
func (e *Excel) AddChart(sheet, cell string, chart *excelize.Chart) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.file.AddChart(sheet, cell, chart)
}

// ExportToFile 导出到文件
func ExportToFile(filename string, data [][]interface{}, headers []string) error {
	excel := NewExcel()
	defer excel.Close()

	// 使用第一个sheet名作为文件名基础
	sheetName := "Sheet1"
	if strings.Contains(filename, ".") {
		name := strings.Split(filename, ".")[0]
		if name != "" {
			sheetName = name
		}
	}

	opt := WriteOption{
		Sheet:   sheetName,
		Headers: headers,
	}

	if err := excel.Write(data, opt); err != nil {
		return err
	}

	return excel.SaveAs(filename)
}

// ImportFromFile 从文件导入
func ImportFromFile(filename string, slicePtr interface{}) error {
	excel, err := Open(filename)
	if err != nil {
		return err
	}
	defer excel.Close()

	opt := ReadOption{
		HasHeader: true,
	}

	return excel.ReadToSlice(slicePtr, opt)
}

// ExportStructs 导出结构体切片
func ExportStructs(filename string, data interface{}, headers []string) error {
	excel := NewExcel()
	defer excel.Close()

	// 智能生成sheet名称
	sheetName := "Data"
	if strings.Contains(filename, ".") {
		name := strings.Split(filename, ".")[0]
		if name != "" {
			sheetName = name
		}
	}

	opt := WriteOption{
		Sheet:   sheetName,
		Headers: headers,
	}

	if err := excel.WriteSlice(data, opt); err != nil {
		return err
	}

	return excel.SaveAs(filename)
}

// StreamWriter 流式写入器
type StreamWriter struct {
	writer *excelize.StreamWriter
	row    int
	mu     sync.Mutex
}

// NewStreamWriter 创建流式写入器
func (e *Excel) NewStreamWriter(sheet string) (*StreamWriter, error) {
	e.mu.Lock()
	writer, err := e.file.NewStreamWriter(sheet)
	if err != nil {
		e.mu.Unlock()
		return nil, err
	}
	e.mu.Unlock()

	return &StreamWriter{
		writer: writer,
		row:    1,
	}, nil
}

// WriteHeader 写入表头
func (sw *StreamWriter) WriteHeader(headers []string) error {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	cells := make([]interface{}, len(headers))
	for i, header := range headers {
		cells[i] = excelize.Cell{Value: header}
	}

	cell, _ := excelize.CoordinatesToCellName(1, sw.row)
	if err := sw.writer.SetRow(cell, cells); err != nil {
		return err
	}

	sw.row++
	return nil
}

// WriteRow 写入行
func (sw *StreamWriter) WriteRow(values []interface{}) error {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	cells := make([]interface{}, len(values))
	for i, value := range values {
		cells[i] = excelize.Cell{Value: value}
	}

	cell, _ := excelize.CoordinatesToCellName(1, sw.row)
	if err := sw.writer.SetRow(cell, cells); err != nil {
		return err
	}

	sw.row++
	return nil
}

// Flush 刷新
func (sw *StreamWriter) Flush() error {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.writer.Flush()
}

// RemoveDefaultSheet 删除默认的Sheet1
func (e *Excel) RemoveDefaultSheet() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	sheets := e.file.GetSheetList()
	if len(sheets) > 1 && sheets[0] == "Sheet1" {
		// 确保至少保留一个sheet
		return e.file.DeleteSheet("Sheet1")
	}
	return nil
}

// CleanSheets 清理工作表，只保留指定的sheet
func (e *Excel) CleanSheets(keepSheets []string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	allSheets := e.file.GetSheetList()
	sheetsToDelete := []string{}

	for _, sheet := range allSheets {
		keep := false
		for _, keepSheet := range keepSheets {
			if sheet == keepSheet {
				keep = true
				break
			}
		}
		if !keep {
			sheetsToDelete = append(sheetsToDelete, sheet)
		}
	}

	for _, sheet := range sheetsToDelete {
		if err := e.file.DeleteSheet(sheet); err != nil {
			return err
		}
	}

	return nil
}
