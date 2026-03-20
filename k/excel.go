package k

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"net/http"
	"net/url"
	"time"
)

// DefExcelColumn Excel列定义（泛型）
// T 为导出数据的结构体类型
type DefExcelColumn[T any] struct {
	Header       string                              // 表头名称
	Hide         bool                                // 是否隐藏列
	Width        float64                             // 列宽（不设置则自动列宽）
	GetValue     func(item T, index int) interface{} // 获取单元格数据
	HeaderStyle  *excelize.Style                     // 表头样式（不传则使用默认样式）
	CellStyle    *excelize.Style                     // 单元格样式（不传则使用默认样式）
	MergeRow     int                                 // 向下合并行数
	MergeCol     int                                 // 向右合并列数
	NumberFormat bool                                // 是否数字格式化
	Decimal      int                                 // 小数位数
	Sum          bool                                // 是否在底部求和
}

// 默认表头样式
func defaultHeaderStyle() *excelize.Style {
	return &excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D9E1F2"},
			Pattern: 1,
		},
	}
}

// 默认单元格样式
func defaultCellStyle() *excelize.Style {
	return &excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	}
}

// 获取数字格式样式
func getNumberFormatStyle(decimal int) *excelize.Style {
	// 根据小数位数选择合适的预定义数字格式
	var numFmt int
	switch decimal {
	case 0:
		numFmt = 1 // 0
	case 1:
		numFmt = 2 // 0.0
	case 2:
		numFmt = 4 // #,##0.00
	default:
		numFmt = 2 // 默认 0.0
	}
	return &excelize.Style{
		NumFmt: numFmt,
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	}
}

// ExportToExcel
// 通用Excel导出函数
// 支持：
// - 自动列宽
// - 冻结表头
// - Excel公式求和
// - 自定义样式
// - 标题显示
func ExportToExcel[T any](fileName string, showTitle bool, columns []DefExcelColumn[T], list []T) (*excelize.File, error) {

	f := excelize.NewFile()

	sheet := "Sheet1"

	// 计算起始行位置
	headerStartRow := 1
	dataStartRow := 2
	if showTitle && fileName != "" {
		headerStartRow = 2
		dataStartRow = 3
		// 写入标题
		titleStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold: true,
				Size: 14,
			},
			Alignment: &excelize.Alignment{
				Horizontal: "center",
				Vertical:   "center",
			},
		})
		// 合并单元格作为标题
		colCount := len(columns)
		endCol, _ := excelize.ColumnNumberToName(colCount)
		f.MergeCell(sheet, "A1", endCol+"1")
		f.SetCellValue(sheet, "A1", fileName)
		f.SetCellStyle(sheet, "A1", endCol+"1", titleStyle)
		// 确保标题行被正确写入
		f.SetRowHeight(sheet, 1, 25)
	}

	// 冻结表头
	ySplit := headerStartRow
	err := f.SetPanes(sheet, &excelize.Panes{
		Freeze: true,
		YSplit: ySplit,
	})
	if err != nil {
		return nil, err
	}

	showColumns := make([]DefExcelColumn[T], 0)

	// 记录列最大宽度（用于自动列宽）
	maxWidth := map[int]int{}

	// 样式缓存（避免重复创建）
	headerStyleCache := map[int]int{}
	cellStyleCache := map[int]int{}

	colIndex := 1

	// 构建表头
	for _, col := range columns {

		if col.Hide {
			continue
		}

		showColumns = append(showColumns, col)

		// 如果没有自定义表头样式则使用默认
		style := col.HeaderStyle
		if style == nil {
			style = defaultHeaderStyle()
		}

		styleID, _ := f.NewStyle(style)

		headerStyleCache[colIndex-1] = styleID

		// 写入表头
		cell, _ := excelize.CoordinatesToCellName(colIndex, headerStartRow)
		f.SetCellValue(sheet, cell, col.Header)
		f.SetCellStyle(sheet, cell, cell, styleID)

		maxWidth[colIndex-1] = len(col.Header)

		colName, _ := excelize.ColumnNumberToName(colIndex)

		// 如果用户设置列宽则直接使用
		if col.Width > 0 {
			err = f.SetColWidth(sheet, colName, colName, col.Width)
			if err != nil {
				return nil, err
			}
		}

		colIndex++
	}

	// 写入数据
	for rowIndex, item := range list {
		for colIndex, col := range showColumns {
			value := col.GetValue(item, rowIndex)

			// 自动格式化时间
			if t, ok := value.(time.Time); ok {
				value = t.Format("2006-01-02 15:04:05")
			}

			// 计算自动列宽
			str := fmt.Sprintf("%v", value)
			if len(str) > maxWidth[colIndex] {
				maxWidth[colIndex] = len(str)
			}

			// 单元格样式
			style := col.CellStyle
			if col.NumberFormat {
				// 数字格式化使用专用样式
				numStyle := getNumberFormatStyle(col.Decimal)
				if style == nil {
					style = numStyle
				} else {
					// 合并样式，确保数字格式被正确应用
					style.NumFmt = numStyle.NumFmt
					if style.Alignment == nil {
						style.Alignment = numStyle.Alignment
					}
				}
			} else if style == nil {
				style = defaultCellStyle()
			}

			styleID, ok := cellStyleCache[colIndex]

			if !ok {
				styleID, _ = f.NewStyle(style)
				cellStyleCache[colIndex] = styleID
			}

			// 写入数据
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+dataStartRow)
			f.SetCellValue(sheet, cell, value)
			f.SetCellStyle(sheet, cell, cell, styleID)
		}
	}

	// 是否需要求和
	hasSum := false
	for _, col := range showColumns {
		if col.Sum {
			hasSum = true
			break
		}
	}

	// Excel公式求和
	if hasSum {
		sumRowIndex := len(list) + dataStartRow
		for colIndex, col := range showColumns {
			if colIndex == 0 {
				// 写入合计
				cell, _ := excelize.CoordinatesToCellName(colIndex+1, sumRowIndex)
				f.SetCellValue(sheet, cell, "合计")
				continue
			}

			if col.Sum {
				colName, _ := excelize.ColumnNumberToName(colIndex + 1)
				formula := fmt.Sprintf(
					"SUM(%s%d:%s%d)",
					colName,
					dataStartRow,
					colName,
					len(list)+dataStartRow-1,
				)

				// 为合计行应用与数据行相同的样式
				style := col.CellStyle
				if col.NumberFormat {
					// 数字格式化使用专用样式
					numStyle := getNumberFormatStyle(col.Decimal)
					if style == nil {
						style = numStyle
					} else {
						// 合并样式，确保数字格式被正确应用
						style.NumFmt = numStyle.NumFmt
						if style.Alignment == nil {
							style.Alignment = numStyle.Alignment
						}
					}
				} else if style == nil {
					style = defaultCellStyle()
				}

				// 获取或创建样式ID
				styleID, ok := cellStyleCache[colIndex]
				if !ok {
					styleID, _ = f.NewStyle(style)
					cellStyleCache[colIndex] = styleID
				}

				// 写入公式
				cell, _ := excelize.CoordinatesToCellName(colIndex+1, sumRowIndex)
				f.SetCellFormula(sheet, cell, formula)
				f.SetCellStyle(sheet, cell, cell, styleID)
			} else {
				// 空单元格
				cell, _ := excelize.CoordinatesToCellName(colIndex+1, sumRowIndex)
				f.SetCellValue(sheet, cell, "")
			}
		}
	}

	// 自动列宽
	for i := range showColumns {

		if showColumns[i].Width > 0 {
			continue
		}

		width := float64(maxWidth[i]) * 1.2

		if width < 10 {
			width = 10
		}

		if width > 50 {
			width = 50
		}

		colName, _ := excelize.ColumnNumberToName(i + 1)

		err = f.SetColWidth(sheet, colName, colName, width)
		if err != nil {
			return nil, err
		}
	}
	f.SetCalcProps(&excelize.CalcPropsOptions{
		FullCalcOnLoad: func() *bool {
			b := true
			return &b
		}(),
	})
	return f, nil
}

// ExportExcelToHttp
// 直接通过HTTP返回Excel文件
func ExportExcelToHttp[T any](w http.ResponseWriter, fileName string, showTitle bool, columns []DefExcelColumn[T], list []T) error {
	f, err := ExportToExcel(fileName, showTitle, columns, list)
	if err != nil {
		return err
	}

	// 设置下载头
	w.Header().Add(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename*=utf-8''%s", url.QueryEscape(fmt.Sprintf("%s.xlsx", fileName))),
	)
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Transfer-Encoding", "binary")

	// 写入响应
	return f.Write(w)
}
