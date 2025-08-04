package utils

func MaskSensitiveValue(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	if len(value) <= 8 {
		return value[:2] + "****" + value[len(value)-2:]
	}
	return value[:2] + "********" + value[len(value)-2:]
}
