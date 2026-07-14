package merger

import (
	"sort"
	"testing"
)

func TestNaturalLess(t *testing.T) {
	in := []string{"seg10.ts", "seg2.ts", "seg1.ts", "seg21.ts", "seg3.ts"}
	want := []string{"seg1.ts", "seg2.ts", "seg3.ts", "seg10.ts", "seg21.ts"}
	sort.Slice(in, func(i, j int) bool { return NaturalLess(in[i], in[j]) })
	for i := range want {
		if in[i] != want[i] {
			t.Fatalf("自然排序错误\n 得到 %v\n 期望 %v", in, want)
		}
	}
}

func TestNaturalLessMixedAndPadded(t *testing.T) {
	// 前导零与纯字典序对比:确保 out2 < out10,且 a < b
	if !NaturalLess("out2.ts", "out10.ts") {
		t.Error("out2 应排在 out10 前")
	}
	if !NaturalLess("clip_001.ts", "clip_002.ts") {
		t.Error("补零场景排序错误")
	}
	if NaturalLess("b.ts", "a.ts") {
		t.Error("纯字母比较错误")
	}
}
