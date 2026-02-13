package models

// Category 分类模型
type Category struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// Bookmark 书签模型
type Bookmark struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Account     string `json:"account"`
	Password    string `json:"password"`
	CategoryID  int64  `json:"category_id"`
	Icon        string `json:"icon"`
	Order       int    `json:"order"`
}

// BookmarkWithCategory 带分类名称的书签模型
type BookmarkWithCategory struct {
	Bookmark
	CategoryName string `json:"category_name"`
}
