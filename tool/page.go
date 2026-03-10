package tool

// GetPage 获取页
// @param PageSiz 每页条数
// @return int 页码
func GetPage(pageNum int, PageSiz int) (int, int) {
	if pageNum < 1 {
		pageNum = 1
	}
	if PageSiz < 1 || PageSiz > 100 {
		PageSiz = 20
	}

	offset := (pageNum - 1) * PageSiz
	return PageSiz, offset
}
