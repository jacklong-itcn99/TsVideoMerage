package merger

import "strings"

// NaturalLess 做"自然排序":把字符串里的数字段按数值比较,使 seg2 < seg10。
// 用于让"添加文件夹"扫描到的分段按人类直觉排序,而非字典序(否则 seg10 会排在 seg2 前)。
func NaturalLess(a, b string) bool {
	a, b = strings.ToLower(a), strings.ToLower(b)
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if isDigit(a[i]) && isDigit(b[j]) {
			na, ni := readNum(a, i)
			nb, nj := readNum(b, j)
			if na != nb {
				return na < nb
			}
			i, j = ni, nj
			continue
		}
		if a[i] != b[j] {
			return a[i] < b[j]
		}
		i++
		j++
	}
	return len(a)-i < len(b)-j
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// readNum 从 s[i] 起读取连续数字,返回其数值和结束下标。
func readNum(s string, i int) (uint64, int) {
	var n uint64
	for i < len(s) && isDigit(s[i]) {
		n = n*10 + uint64(s[i]-'0')
		i++
	}
	return n, i
}
