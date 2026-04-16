package tool

import (
	"regexp"
	"strings"
)

// ExtractISBN978 从字符串中提取 978 开头的 ISBN-13
// 返回第一个匹配到的 ISBN，如果没有匹配则返回空字符串
func ExtractISBN978(text string) string {
	// 匹配 978 开头的13位数字（可能包含连字符或空格）
	// 格式：978-xxx-xxxxx-xxx 或 978xxxxxxxxxx
	isbnRegex := regexp.MustCompile(`978[\s-]?\d{1,5}[\s-]?\d{1,7}[\s-]?\d{1,6}[\s-]?\d`)

	matches := isbnRegex.FindAllString(text, -1)
	for _, match := range matches {
		cleaned := cleanISBN978(match)
		if isValidISBN13(cleaned) {
			return cleaned
		}
	}
	return ""
}

// ExtractAllISBN978 从字符串中提取所有 978 开头的 ISBN-13
func ExtractAllISBN978(text string) []string {
	isbnRegex := regexp.MustCompile(`978[\s-]?\d{1,5}[\s-]?\d{1,7}[\s-]?\d{1,6}[\s-]?\d`)

	var results []string
	matches := isbnRegex.FindAllString(text, -1)
	for _, match := range matches {
		cleaned := cleanISBN978(match)
		if isValidISBN13(cleaned) && !contains(results, cleaned) {
			results = append(results, cleaned)
		}
	}
	return results
}

// cleanISBN978 清理 ISBN，移除连字符和空格
func cleanISBN978(isbn string) string {
	re := regexp.MustCompile(`[-\s]`)
	return re.ReplaceAllString(isbn, "")
}

// isValidISBN13 验证 ISBN-13
func isValidISBN13(isbn string) bool {
	if len(isbn) != 13 {
		return false
	}

	var sum int
	for i, ch := range isbn {
		if ch < '0' || ch > '9' {
			return false
		}
		digit := int(ch - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	return sum%10 == 0
}

// contains 辅助函数
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 更简单的版本：只提取13位数字并检查是否以978开头
// ExtractISBN978Simple 简单版本，只匹配连续的数字
func ExtractISBN978Simple(text string) string {
	// 匹配13位连续的数字
	re := regexp.MustCompile(`\d{13}`)
	matches := re.FindAllString(text, -1)

	for _, match := range matches {
		if strings.HasPrefix(match, "978") && isValidISBN13(match) {
			return match
		}
	}
	return ""
}

// ExtractAllISBN978Simple 简单版本，提取所有13位数字中以978开头的
func ExtractAllISBN978Simple(text string) []string {
	re := regexp.MustCompile(`\d{13}`)
	matches := re.FindAllString(text, -1)

	var results []string
	for _, match := range matches {
		if strings.HasPrefix(match, "978") && isValidISBN13(match) && !contains(results, match) {
			results = append(results, match)
		}
	}
	return results
}
