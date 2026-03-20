package k

import (
	"github.com/xuri/excelize/v2"
	"testing"
	"time"
)

func TestExportToExcel(t *testing.T) {
	type Order struct {
		ID        int
		UserName  string
		Amount    float64
		CreatedAt time.Time
	}

	column := []DefExcelColumn[Order]{
		{
			Header: "ID",
			Width:  10,
			GetValue: func(item Order, index int) interface{} {
				return item.ID
			},
		},
		{
			Header: "用户名",
			Width:  20,
			GetValue: func(item Order, index int) interface{} {
				return item.UserName
			},
		},
		{
			Header:       "金额",
			Width:        20,
			NumberFormat: true,
			Decimal:      2,
			Sum:          true,
			CellStyle: &excelize.Style{
				Alignment: &excelize.Alignment{Horizontal: "right"},
			},
			GetValue: func(item Order, index int) interface{} {
				return item.Amount
			},
		},
		{
			Header: "创建时间",
			Width:  25,
			GetValue: func(item Order, index int) interface{} {
				return item.CreatedAt
			},
		},
	}
	dataList := []Order{
		{ID: 1, UserName: "张三", Amount: 123.456, CreatedAt: time.Now()},
		{ID: 2, UserName: "李四", Amount: 888.8, CreatedAt: time.Now()},
		{ID: 3, UserName: "王五", Amount: 11111166.6666, CreatedAt: time.Now()},
	}
	f, err := ExportToExcel("销售表", false, column, dataList)
	if err != nil {
		t.Fatal("导出excel失败", err)
	}
	f.SaveAs("销售表.xlsx")
}
