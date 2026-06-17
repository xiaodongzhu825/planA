package psiMysql

import (
	"planA/initialization/golabl"
	psiMysqlType "planA/type/psiMysql"
)

// GetBookInfo 获取书籍信息
func GetBookInfo(isbn string, fisbn string) (bookInfo psiMysqlType.BookInfo, err error) {
	err = golabl.PsiMysqlDb.Where("isbn = ? and fisbn = ?", isbn, fisbn).First(&bookInfo).Error
	return bookInfo, err
}
