package sensitivematcher

import "unicode/utf8"

func bytesToRunes(data []byte) []rune {
	tempData := make([]byte, len(data))
	copy(tempData, data)
	runes := make([]rune, 0, utf8.RuneCount(tempData)) // 预分配 rune 切片
	for len(tempData) > 0 {
		r, size := utf8.DecodeRune(tempData)
		runes = append(runes, r)
		tempData = tempData[size:]
	}
	return runes
}
