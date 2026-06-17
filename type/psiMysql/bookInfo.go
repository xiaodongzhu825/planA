package psiMysql

// BookInfo 书籍信息表结构体
type BookInfo struct {
	ID              int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id" db:"id"`                                              // 自增ID
	FID             int64  `gorm:"column:fid;not null;default:0" json:"fid" db:"fid"`                                                 // 父级ID
	Type            int8   `gorm:"column:type;not null" json:"type" db:"type"`                                                        // 类型 1正常 2套装书 3 一号多书 4无书号
	ISBN            string `gorm:"column:isbn;not null;default:'';size:20" json:"isbn" db:"isbn"`                                     // ISBN
	BookName        string `gorm:"column:book_name;not null;default:'';size:100" json:"book_name" db:"book_name"`                     // 书名
	Author          string `gorm:"column:author;not null;default:'';size:100" json:"author" db:"author"`                              // 作者
	Publishing      string `gorm:"column:publishing;not null;default:'';size:50" json:"publishing" db:"publishing"`                   // 出版社
	PublicationDate string `gorm:"column:publication_date;not null;default:'';size:10" json:"publication_date" db:"publication_date"` // 出版日期
	PublicationTime int64  `gorm:"column:publication_time;not null;default:0" json:"publication_time" db:"publication_time"`          // 出版日期时间戳
	Binding         string `gorm:"column:binding;not null;default:'';size:10" json:"binding" db:"binding"`                            // 装帧
	PagesCount      int64  `gorm:"column:pages_count;not null;default:0" json:"pages_count" db:"pages_count"`                         // 页数
	WordsCount      int64  `gorm:"column:words_count;not null;default:0" json:"words_count" db:"words_count"`                         // 字数
	Format          int64  `gorm:"column:format;not null;default:0" json:"format" db:"format"`                                        // 开本
	Price           int64  `gorm:"column:price;not null;default:0" json:"price" db:"price"`                                           // 价格
	CatID           string `gorm:"column:cat_id;type:json;not null" json:"cat_id" db:"cat_id"`                                        // 类目json
	FISBN           string `gorm:"column:fisbn;not null;default:'';size:20" json:"fisbn" db:"fisbn"`                                  // FISBN
	FBookName       string `gorm:"column:f_book_name;not null;default:'';size:100" json:"f_book_name" db:"f_book_name"`               // 副书名
	LiveImage       string `gorm:"column:live_image;type:json;not null" json:"live_image" db:"live_image"`                            // 实拍图json
}

// TableName 指定表名
func (BookInfo) TableName() string {
	return "book_info"
}
