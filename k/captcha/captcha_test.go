package captcha

import "testing"

func TestDriverDigitAndVerify(t *testing.T) {
	id, b64s, answer, err := DriverDigitFunc()
	if err != nil {
		t.Fatalf("DriverDigitFunc: %v", err)
	}
	if id == "" || b64s == "" || answer == "" {
		t.Fatalf("empty captcha: id=%q answer=%q b64len=%d", id, answer, len(b64s))
	}
	if !Verify(id, answer, true) {
		t.Fatal("正确验证码应通过")
	}
	if Verify(id, answer, true) {
		t.Fatal("clear 后不应再通过")
	}
}
