package data

import (
	TestType "planA/test/type"
)

var DataArr = []string{
	"{\"book_info\":{\"isbn\":\"97875617998191\"},\"detail\":{\"price\":50000,\"stock\":0}}",
	"{\"book_info\":{\"isbn\":\"9787533160227\"},\"detail\":{\"price\":50000,\"stock\":0}}",
	"{\"book_info\":{\"isbn\":\"9787505734944\"},\"detail\":{\"price\":10000000,\"stock\":0}}",
	"{\"book_info\":{\"isbn\":\"9787801532244\"},\"detail\":{\"price\":50000,\"stock\":0}}",
}

var DataReturn = []TestType.BodyErr{
	{"9787561799819", "无书品信息"},
	{"9787533160227", "违规词命中"},
	{"9787505734944", "不在价格区间内"},
	{"9787801532244", "缺少官图"},
}
