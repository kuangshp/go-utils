package k

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// ExcelHorizontal 水平对齐方式枚举，用于 CellStyle/HeaderStyle 中的 Alignment.Horizontal 字段。
// 使用枚举代替裸字符串，避免拼写错误。
//
// 示例：
//
//	CellStyle: &excelize.Style{
//	    Alignment: &excelize.Alignment{Horizontal: string(ExcelHorizontalRight)},
//	}
type ExcelHorizontal string

const (
	ExcelHorizontalLeft   ExcelHorizontal = "left"   // 左对齐（默认）
	ExcelHorizontalCenter ExcelHorizontal = "center" // 居中对齐
	ExcelHorizontalRight  ExcelHorizontal = "right"  // 右对齐，常用于金额、数字列
)

// DefExcelColumn Excel 列定义（泛型），T 为导出数据的结构体类型。
//
// 最简用法（只需表头 + 取值函数）：
//
//	DefExcelColumn[Order]{
//	    Header: "订单号",
//	    GetValue: func(item Order, index int) interface{} {
//	        return item.OrderNo
//	    },
//	}
//
// 金额列（右对齐 + 千分位 + 底部合计）：
//
//	DefExcelColumn[Order]{
//	    Header:       "实付金额",
//	    NumberFormat: true,
//	    Decimal:      2,
//	    Sum:          true,
//	    CellStyle: &excelize.Style{
//	        Alignment: &excelize.Alignment{Horizontal: string(ExcelHorizontalRight)},
//	    },
//	    GetValue: func(item Order, index int) interface{} {
//	        return item.TotalAmount // 支持 float64 / decimal.Decimal 等类型
//	    },
//	}
//
// 时间列（不设 Width 则自动宽度）：
//
//	DefExcelColumn[Order]{
//	    Header: "支付时间",
//	    GetValue: func(item Order, index int) interface{} {
//	        return item.PayTime.Format(time.DateTime)
//	    },
//	}
//
// 隐藏列：
//
//	DefExcelColumn[Order]{
//	    Header: "内部ID",
//	    Hide:   true,
//	    GetValue: func(item Order, index int) interface{} {
//	        return item.ID
//	    },
//	}
type DefExcelColumn[T any] struct {
	Header       string                              // 表头名称
	Hide         bool                                // 是否隐藏该列，隐藏后不写入 Excel
	Width        float64                             // 固定列宽；不设置（传 0）则根据内容自动计算
	GetValue     func(item T, index int) interface{} // 获取单元格数据，index 为行下标（从 0 开始）
	HeaderStyle  *excelize.Style                     // 表头样式；不传则使用默认（加粗、居中、浅蓝底色）
	CellStyle    *excelize.Style                     // 单元格样式；不传则使用默认（左对齐）
	MergeRow     int                                 // 向下合并行数（预留字段）
	MergeCol     int                                 // 向右合并列数（预留字段）
	NumberFormat bool                                // 是否启用数字格式化（千分位）
	Decimal      int                                 // 小数位数，配合 NumberFormat 使用；0/1/2 对应 0、0.0、#,##0.00
	Sum          bool                                // 是否在数据末尾追加合计行，值为各行数据的算术和
}

// defaultHeaderStyle 默认表头样式：加粗、水平居中、垂直居中、浅蓝底色
func defaultHeaderStyle() *excelize.Style {
	return &excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Alignment: &excelize.Alignment{
			Horizontal: string(ExcelHorizontalCenter),
			Vertical:   "center",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D9E1F2"},
			Pattern: 1,
		},
	}
}

// defaultCellStyle 默认单元格样式：左对齐
func defaultCellStyle() *excelize.Style {
	return &excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: string(ExcelHorizontalLeft),
		},
	}
}

// getNumberFormatStyle 根据小数位数返回对应的数字格式样式。
// decimal=0 → "0"，decimal=1 → "0.0"，decimal=2 → "#,##0.00"（千分位）
func getNumberFormatStyle(decimal int) *excelize.Style {
	var numFmt int
	switch decimal {
	case 0:
		numFmt = 1 // 0
	case 1:
		numFmt = 2 // 0.0
	case 2:
		numFmt = 4 // #,##0.00
	default:
		numFmt = 2
	}
	return &excelize.Style{
		NumFmt: numFmt,
		Alignment: &excelize.Alignment{
			Horizontal: string(ExcelHorizontalLeft),
		},
	}
}

// toFloat64 将任意类型转为 float64，用于 Sum 合计累加。
// 支持所有 Go 原生数值类型、字符串，以及实现了 fmt.Stringer 的类型（如 shopspring/decimal.Decimal）。
// 无法转换时返回 0。
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		// 兜底：支持 decimal.Decimal 等实现了 fmt.Stringer 的类型
		// decimal.Decimal.String() 返回如 "123.45" 的字符串，可直接解析
		if s, ok := v.(fmt.Stringer); ok {
			f, _ := strconv.ParseFloat(s.String(), 64)
			return f
		}
	}
	return 0
}

// ExportToExcel 通用 Excel 导出函数，返回 *excelize.File，可自行保存或写入 HTTP 响应。
//
// 参数：
//   - fileName : 文件名（showTitle=true 时同时作为第一行大标题）
//   - showTitle: 是否在第一行显示大标题（合并单元格）
//   - columns  : 列定义列表，见 DefExcelColumn
//   - list     : 数据列表
//
// 特性：
//   - 表头默认加粗居中、浅蓝底色，数据行默认左对齐
//   - Width=0 时自动按内容计算列宽，中英文混排友好
//   - 设置 Sum=true 的列会在末尾追加合计行，直接写入计算值（兼容所有 Excel/WPS 版本）
//   - time.Time 类型自动格式化为 "2006-01-02 15:04:05"
//
// 示例：
//
//	type Order struct {
//	    No          string
//	    TotalAmount decimal.Decimal
//	    PayTime     time.Time
//	}
//
//	columns := []k.DefExcelColumn[Order]{
//	    {
//	        Header: "订单号",
//	        GetValue: func(item Order, index int) interface{} { return item.No },
//	    },
//	    {
//	        Header:       "实付金额",
//	        NumberFormat: true,
//	        Decimal:      2,
//	        Sum:          true,
//	        CellStyle: &excelize.Style{
//	            Alignment: &excelize.Alignment{Horizontal: string(k.ExcelHorizontalRight)},
//	        },
//	        GetValue: func(item Order, index int) interface{} { return item.TotalAmount },
//	    },
//	    {
//	        Header: "支付时间",
//	        GetValue: func(item Order, index int) interface{} {
//	            return item.PayTime.Format(time.DateTime)
//	        },
//	    },
//	}
//
//	f, err := k.ExportToExcel("订单列表", false, columns, list)
//	if err != nil { ... }
//	f.SaveAs("orders.xlsx")
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
				Horizontal: string(ExcelHorizontalCenter),
				Vertical:   "center",
			},
		})
		// 合并单元格作为标题
		colCount := len(columns)
		endCol, _ := excelize.ColumnNumberToName(colCount)
		f.MergeCell(sheet, "A1", endCol+"1")
		f.SetCellValue(sheet, "A1", fileName)
		f.SetCellStyle(sheet, "A1", endCol+"1", titleStyle)
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

	// 记录每列视觉最大宽度（用于自动列宽）
	maxWidth := map[int]int{}

	// 样式缓存（避免对同一列重复调用 NewStyle）
	headerStyleCache := map[int]int{}
	cellStyleCache := map[int]int{}

	colIndex := 1

	// 构建表头
	for _, col := range columns {

		if col.Hide {
			continue
		}

		showColumns = append(showColumns, col)

		style := col.HeaderStyle
		if style == nil {
			style = defaultHeaderStyle()
		}

		styleID, _ := f.NewStyle(style)
		headerStyleCache[colIndex-1] = styleID

		cell, _ := excelize.CoordinatesToCellName(colIndex, headerStartRow)
		f.SetCellValue(sheet, cell, col.Header)
		f.SetCellStyle(sheet, cell, cell, styleID)

		maxWidth[colIndex-1] = runeWidth(col.Header)

		colName, _ := excelize.ColumnNumberToName(colIndex)
		if col.Width > 0 {
			err = f.SetColWidth(sheet, colName, colName, col.Width)
			if err != nil {
				return nil, err
			}
		}

		colIndex++
	}

	// 各列求和累加器（key 为列下标）
	sumValues := map[int]float64{}

	// 写入数据行
	for rowIndex, item := range list {
		for colIndex, col := range showColumns {
			value := col.GetValue(item, rowIndex)

			// time.Time 自动格式化
			if t, ok := value.(time.Time); ok {
				value = t.Format("2006-01-02 15:04:05")
			}

			// 累加求和
			if col.Sum {
				sumValues[colIndex] += toFloat64(value)
			}

			// 记录该列最大视觉宽度
			str := fmt.Sprintf("%v", value)
			w := runeWidth(str)
			if w > maxWidth[colIndex] {
				maxWidth[colIndex] = w
			}

			// 单元格样式（NumberFormat 优先合并数字格式）
			style := col.CellStyle
			if col.NumberFormat {
				numStyle := getNumberFormatStyle(col.Decimal)
				if style == nil {
					style = numStyle
				} else {
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

			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+dataStartRow)
			f.SetCellValue(sheet, cell, value)
			f.SetCellStyle(sheet, cell, cell, styleID)
		}
	}

	// 检查是否需要合计行
	hasSum := false
	for _, col := range showColumns {
		if col.Sum {
			hasSum = true
			break
		}
	}

	// 写入合计行（直接写入累加值，不依赖 Excel 公式，兼容 WPS / 程序读取等场景）
	if hasSum {
		sumRowIndex := len(list) + dataStartRow
		for colIndex, col := range showColumns {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, sumRowIndex)
			if colIndex == 0 {
				f.SetCellValue(sheet, cell, "合计")
				continue
			}
			if col.Sum {
				style := col.CellStyle
				if col.NumberFormat {
					numStyle := getNumberFormatStyle(col.Decimal)
					if style == nil {
						style = numStyle
					} else {
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

				f.SetCellValue(sheet, cell, sumValues[colIndex])
				f.SetCellStyle(sheet, cell, cell, styleID)
			} else {
				f.SetCellValue(sheet, cell, "")
			}
		}
	}

	// 自动列宽：中文字符按 2 个英文字符宽度计算，仅对未手动设置 Width 的列生效
	for i := range showColumns {
		if showColumns[i].Width > 0 {
			continue
		}
		width := calcColumnWidth(maxWidth[i])
		colName, _ := excelize.ColumnNumberToName(i + 1)
		err = f.SetColWidth(sheet, colName, colName, width)
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}

// ExportExcelToHttp 将 Excel 文件直接写入 HTTP 响应，触发浏览器下载。
// 内部调用 ExportToExcel，参数含义相同。
//
// 示例（Gin）：
//
//	func DownloadOrders(c *gin.Context) {
//	    err := k.ExportExcelToHttp(c.Writer, "订单列表", false, columns, list)
//	    if err != nil {
//	        c.JSON(500, gin.H{"error": err.Error()})
//	    }
//	}
//
// 示例（标准库）：
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    err := k.ExportExcelToHttp(w, "订单列表", false, columns, list)
//	    if err != nil {
//	        http.Error(w, err.Error(), 500)
//	    }
//	}
func ExportExcelToHttp[T any](w http.ResponseWriter, fileName string, showTitle bool, columns []DefExcelColumn[T], list []T) error {
	f, err := ExportToExcel(fileName, showTitle, columns, list)
	if err != nil {
		return err
	}

	w.Header().Add(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename*=utf-8''%s", url.QueryEscape(fmt.Sprintf("%s.xlsx", fileName))),
	)
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Transfer-Encoding", "binary")

	return f.Write(w)
}

// ReadExcelFromRequest 解析前端通过 multipart/form-data 上传的 Excel 文件，返回二维字符串数组。
// 第一行通常为表头，后续行为数据行，所有单元格均转为字符串。
// 每行列数已补齐为最大列数，不会因尾部空单元格导致长度不一致。
//
// 参数：
//   - r         : HTTP 请求，Content-Type 须为 multipart/form-data
//   - fieldName : 表单字段名，与前端 <input name="xxx"> 或 FormData.append("xxx", file) 对应
//   - sheetIndex: 读取第几个 Sheet，从 0 开始；传 0 即读第一个 Sheet
//
// 示例（Gin）：
//
//	func ImportOrders(c *gin.Context) {
//	    rows, err := k.ReadExcelFromRequest(c.Request, "file", 0)
//	    if err != nil {
//	        c.JSON(400, gin.H{"error": err.Error()})
//	        return
//	    }
//	    // rows[0] 为表头，rows[1:] 为数据
//	    for i, row := range rows[1:] {
//	        fmt.Printf("第%d行: 订单号=%s 金额=%s\n", i+1, row[0], row[1])
//	    }
//	}
//
// 示例（标准库）：
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    rows, err := k.ReadExcelFromRequest(r, "file", 0)
//	    if err != nil {
//	        http.Error(w, err.Error(), 400)
//	        return
//	    }
//	    for _, row := range rows[1:] {
//	        fmt.Println(row)
//	    }
//	}
func ReadExcelFromRequest(r *http.Request, fieldName string, sheetIndex int) ([][]string, error) {
	// 限制最大内存 32MB，超出部分写临时文件
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, fmt.Errorf("解析表单失败: %w", err)
	}

	file, _, err := r.FormFile(fieldName)
	if err != nil {
		return nil, fmt.Errorf("获取上传文件失败: %w", err)
	}
	defer file.Close()

	return readExcelFromReader(file, sheetIndex)
}

// ReadExcelFromFile 读取本地 Excel 文件，返回二维字符串数组。
// 第一行通常为表头，后续行为数据行，所有单元格均转为字符串。
// 每行列数已补齐为最大列数，不会因尾部空单元格导致长度不一致。
//
// 参数：
//   - filePath  : 本地文件路径，支持相对路径和绝对路径
//   - sheetIndex: 读取第几个 Sheet，从 0 开始；传 0 即读第一个 Sheet
//
// 示例：
//
//	rows, err := k.ReadExcelFromFile("./data/orders.xlsx", 0)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// rows[0] 为表头
//	headers := rows[0]  // ["订单号", "金额", "时间"]
//	// rows[1:] 为数据
//	for _, row := range rows[1:] {
//	    fmt.Println(row[0], row[1], row[2])
//	}
func ReadExcelFromFile(filePath string, sheetIndex int) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	return readExcelFromReader(file, sheetIndex)
}

// readExcelFromReader 内部通用方法，从任意 io.Reader 解析 Excel，返回二维字符串数组。
// excelize 读取行时会省略尾部空单元格，此处统一补齐各行到最大列数。
func readExcelFromReader(reader io.Reader, sheetIndex int) ([][]string, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("解析 Excel 失败: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel 文件中没有 Sheet")
	}
	if sheetIndex < 0 || sheetIndex >= len(sheets) {
		return nil, fmt.Errorf("sheetIndex %d 超出范围，共 %d 个 Sheet", sheetIndex, len(sheets))
	}

	rows, err := f.GetRows(sheets[sheetIndex])
	if err != nil {
		return nil, fmt.Errorf("读取 Sheet[%s] 失败: %w", sheets[sheetIndex], err)
	}

	// 补齐每行列数，保证各行长度一致，防止业务侧按列下标取值时越界
	maxCol := 0
	for _, row := range rows {
		if len(row) > maxCol {
			maxCol = len(row)
		}
	}
	result := make([][]string, 0, len(rows))
	for _, row := range rows {
		normalized := make([]string, maxCol)
		copy(normalized, row)
		result = append(result, normalized)
	}

	return result, nil
}

// runeWidth 计算字符串的视觉宽度，用于自动列宽计算。
// ASCII 字符（英文、数字、半角符号）算 1，其余（中文、日文、全角符号等）算 2。
func runeWidth(s string) int {
	width := 0
	for _, r := range s {
		if r > 0x7F {
			width += 2
		} else {
			width += 1
		}
	}
	return width
}

// calcColumnWidth 将视觉宽度换算为 Excel 列宽单位，加左右 padding 并限制最小/最大值。
// 最小宽度 8，最大宽度 60。
func calcColumnWidth(visualWidth int) float64 {
	// +2 留出左右 padding，* 1.1 补偿字体渲染差异
	width := float64(visualWidth+2) * 1.1
	if width < 8 {
		width = 8
	}
	if width > 60 {
		width = 60
	}
	return width
}
