package ed2ksrv

import "testing"

func TestNormalizeDisplayTextPreservesUTF8(t *testing.T) {
	got := normalizeDisplayText("测试客户端")
	if got != "测试客户端" {
		t.Fatalf("unexpected utf8 normalization: %q", got)
	}
}

func TestNormalizeDisplayTextDecodesGB18030(t *testing.T) {
	got := normalizeDisplayText(string([]byte{0xB2, 0xE2, 0xCA, 0xD4, 0xBF, 0xCD, 0xBB, 0xA7, 0xB6, 0xCB}))
	if got != "测试客户端" {
		t.Fatalf("unexpected gb18030 normalization: %q", got)
	}
}
