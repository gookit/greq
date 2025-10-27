package requtil

import "strings"

// FilenameFromDisposition 从 Content-Disposition 提取文件名
func FilenameFromDisposition(disposition string) string {
	if disposition == "" {
		return ""
	}

	parts := strings.Split(disposition, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "filename=") {
			filename := strings.TrimPrefix(part, "filename=")
			return strings.Trim(filename, "\"'")
		}
	}

	return ""
}
