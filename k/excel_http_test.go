package k

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExportExcelToHttp(t *testing.T) {
	type Order struct {
		ID       int64
		UserName string
	}
	columns := []DefExcelColumn[Order]{
		{Header: "ID", GetValue: func(item Order, index int) interface{} { return item.ID }},
		{Header: "用户", GetValue: func(item Order, index int) interface{} { return item.UserName }},
	}
	list := []Order{{ID: 1, UserName: "alice"}, {ID: 2, UserName: "bob"}}

	recorder := httptest.NewRecorder()
	if err := ExportExcelToHttp(recorder, "订单列表", false, columns, list); err != nil {
		t.Fatalf("ExportExcelToHttp: %v", err)
	}
	disposition := recorder.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, "attachment;") || !strings.Contains(disposition, ".xlsx") {
		t.Fatalf("Content-Disposition = %q", disposition)
	}
	if ct := recorder.Header().Get("Content-Type"); ct == "" {
		t.Fatal("缺少 Content-Type")
	}
	if recorder.Body.Len() == 0 {
		t.Fatal("响应体为空")
	}
	_ = time.Now()
}
