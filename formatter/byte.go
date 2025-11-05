package formatter

// DecimalBytes 返回十进制标准（以1000为基数）下的可读字节单位字符串。
// precision参数指定小数点后的位数，默认为4
func DecimalBytes(size float64, precision ...int) string {
	pointPosition := 4
	if len(precision) > 0 {
		pointPosition = precision[0]
	}

	size, unit := calculateByteSize(size, 1000.0, decimalByteUnits)

	return roundToToString(size, pointPosition) + unit
}

// BinaryBytes 返回binary标准（以1024为基数）下的可读字节单位字符串。
// precision参数指定小数点后的位数，默认为4
func BinaryBytes(size float64, precision ...int) string {
	pointPosition := 4
	if len(precision) > 0 {
		pointPosition = precision[0]
	}

	size, unit := calculateByteSize(size, 1024.0, binaryByteUnits)

	return roundToToString(size, pointPosition) + unit
}

// ParseDecimalBytes 将字节单位字符串转换成其所表示的字节数（以1000为基数）
func ParseDecimalBytes(size string) (uint64, error) {
	return parseBytes(size, "decimal")
}

// ParseBinaryBytes 将字节单位字符串转换成其所表示的字节数（以1024为基数）
func ParseBinaryBytes(size string) (uint64, error) {
	return parseBytes(size, "binary")
}
