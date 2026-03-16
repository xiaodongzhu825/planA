package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidatorRule 验证规则
func ValidatorRule(err error, fieldCN map[string]string) string {
	if err == nil {
		return ""
	}

	// 断言为validator的验证错误类型
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return "参数验证失败：" + err.Error()
	}

	// 遍历错误（返回第一个错误，符合接口友好性；如需返回所有错误可改为拼接字符串）
	for _, e := range validationErrs {
		field := e.Field() // 获取当前错误的字段名（如 ShopType）
		tag := e.Tag()     // 获取验证规则（如 oneof）
		param := e.Param() // 获取规则参数（如 1 2 5）
		fieldName := field // 默认使用原字段名
		if cn, ok := fieldCN[field]; ok {
			fieldName = cn // 替换为中文名称
		}

		switch tag {
		case "required":
			return fmt.Sprintf("%s为必填项", fieldName)
		case "email":
			return fmt.Sprintf("%s格式不正确", fieldName)
		case "min":
			// 区分字符串长度min和数字值min
			if isNumericField(field) {
				return fmt.Sprintf("%s不能小于%s", fieldName, param)
			}
			return fmt.Sprintf("%s长度不能少于%s个字符", fieldName, param)
		case "max":
			if isNumericField(field) {
				return fmt.Sprintf("%s不能大于%s", fieldName, param)
			}
			return fmt.Sprintf("%s长度不能超过%s个字符", fieldName, param)
		case "gte":
			return fmt.Sprintf("%s必须大于等于%s", fieldName, param)
		case "lte":
			return fmt.Sprintf("%s必须小于等于%s", fieldName, param)
		case "phone":
			return fmt.Sprintf("%s格式不正确（请填写11位手机号）", fieldName)
		case "oneof":
			// 格式化oneof的可选值（如 "1 2 5" → "1、2、5"）
			options := strings.ReplaceAll(param, " ", "、")
			return fmt.Sprintf("%s只能填写%s中的一个", fieldName, options)
		case "numeric":
			return fmt.Sprintf("%s必须是数字格式", fieldName)
		case "shop_type_only_5":
			return fmt.Sprintf("%s仅允许填写5", fieldName)
		default:
			return fmt.Sprintf("%s验证失败（规则：%s）", fieldName, tag)
		}
	}

	return ""
}

// isNumericField 判断字段是否为数字类型（用于区分min/max是长度还是数值）
func isNumericField(field string) bool {
	numericFields := []string{"TaskCount", "Age"}
	for _, f := range numericFields {
		if f == field {
			return true
		}
	}
	return false
}
