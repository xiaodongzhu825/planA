package _type

// 书籍结构体

// Book 书籍信息
type Book struct {
	ID                 int64     `json:"id"`
	BookName           string    `json:"book_name"`
	BookPic            BookImage `json:"book_pic"`   //pddPath 轮播图第一张  官图
	BookPicS           BookImage `json:"book_pic_s"` //pddPath 轮播图第一张  实拍图
	BookPicObj         string    `json:"book_pic_obj"`
	BookDetailImage    BookImage `json:"book_detail_image"`    //pddPath 详情图
	BookPicB           string    `json:"book_pic_b"`           //pddPath 白底图
	BookDirectoryImage BookImage `json:"book_directory_image"` //pddPath 目录图
	ISBN               string    `json:"isbn"`
	Author             string    `json:"author"`           //作者
	Category           string    `json:"category"`         //分类
	Publisher          string    `json:"publisher"`        //出版社
	PublicationTime    string    `json:"publication_time"` //出版时间
	BindingLayout      string    `json:"binding_layout"`   //装帧
	FixPrice           float64   `json:"fix_price"`        //售价
	Content            string    `json:"content"`          //简介
	IsSuit             int64     `json:"is_suit"`          //套装
	IsIllegal          int64     `json:"is_illegal"`       //
	IsReturn           int64     `json:"is_return"`        //驳回
	IsFilter           string    `json:"is_filter"`        //过滤
}

type BookImage struct {
	LocalPath string `json:"localPath"`
	PddPath   string `json:"pddPath"` //轮播图第一张  官图
}
