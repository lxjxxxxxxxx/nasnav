package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"nasnav/models"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// Init 初始化数据库连接，创建数据库文件目录和数据表
func Init(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// createTables 创建分类和书签数据表
func createTables() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			"order" INTEGER NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS bookmarks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			url TEXT NOT NULL,
			description TEXT,
			account TEXT,
			password TEXT,
			category_id INTEGER NOT NULL,
			icon TEXT,
			"order" INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (category_id) REFERENCES categories(id)
		);
	`)
	return err
}

// GetCategories 获取所有分类，按排序字段升序排列
func GetCategories() ([]models.Category, error) {
	rows, err := DB.Query(`SELECT id, name, "order" FROM categories ORDER BY "order" ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Order); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// CreateCategory 创建新分类，自动设置排序值为当前最大值加1
func CreateCategory(name string) (*models.Category, error) {
	var maxOrder int
	err := DB.QueryRow(`SELECT COALESCE(MAX("order"), 0) FROM categories`).Scan(&maxOrder)
	if err != nil {
		return nil, err
	}

	result, err := DB.Exec(`INSERT INTO categories (name, "order") VALUES (?, ?)`, name, maxOrder+1)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.Category{ID: id, Name: name, Order: maxOrder + 1}, nil
}

// UpdateCategory 更新分类名称
func UpdateCategory(id int64, name string) error {
	_, err := DB.Exec(`UPDATE categories SET name = ? WHERE id = ?`, name, id)
	return err
}

// DeleteCategory 删除分类及其下属的所有书签
func DeleteCategory(id int64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM bookmarks WHERE category_id = ?`, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM categories WHERE id = ?`, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateCategoryOrder 批量更新分类排序
func UpdateCategoryOrder(ids []int64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range ids {
		_, err := tx.Exec(`UPDATE categories SET "order" = ? WHERE id = ?`, i+1, id)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetBookmarks 获取书签列表，categoryID为0时获取所有书签，否则获取指定分类的书签
func GetBookmarks(categoryID int64, isAuthenticated bool) ([]models.BookmarkWithCategory, error) {
	var rows *sql.Rows
	var err error

	if categoryID == 0 {
		rows, err = DB.Query(`
			SELECT b.id, b.title, b.url, b.description, b.account, b.password, 
				   b.category_id, b.icon, b."order", c.name
			FROM bookmarks b
			LEFT JOIN categories c ON b.category_id = c.id
			ORDER BY b."order" ASC
		`)
	} else {
		rows, err = DB.Query(`
			SELECT b.id, b.title, b.url, b.description, b.account, b.password, 
				   b.category_id, b.icon, b."order", c.name
			FROM bookmarks b
			LEFT JOIN categories c ON b.category_id = c.id
			WHERE b.category_id = ?
			ORDER BY b."order" ASC
		`, categoryID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []models.BookmarkWithCategory
	for rows.Next() {
		var b models.BookmarkWithCategory
		var description, account, password, icon, categoryName sql.NullString
		if err := rows.Scan(&b.ID, &b.Title, &b.URL, &description, &account, &password,
			&b.CategoryID, &icon, &b.Order, &categoryName); err != nil {
			return nil, err
		}
		b.Description = description.String
		b.Account = account.String
		b.Password = password.String
		b.Icon = icon.String
		b.CategoryName = categoryName.String
		if !isAuthenticated {
			b.Account = ""
			b.Password = ""
		}
		bookmarks = append(bookmarks, b)
	}
	return bookmarks, nil
}

// CreateBookmark 创建新书签，自动设置排序值为当前最大值加1
func CreateBookmark(b *models.Bookmark) (*models.Bookmark, error) {
	var maxOrder int
	err := DB.QueryRow(`SELECT COALESCE(MAX("order"), 0) FROM bookmarks`).Scan(&maxOrder)
	if err != nil {
		return nil, err
	}

	result, err := DB.Exec(`
		INSERT INTO bookmarks (title, url, description, account, password, category_id, icon, "order")
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, b.Title, b.URL, b.Description, b.Account, b.Password, b.CategoryID, b.Icon, maxOrder+1)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	b.ID = id
	b.Order = maxOrder + 1
	return b, nil
}

// UpdateBookmark 更新书签信息
func UpdateBookmark(b *models.Bookmark) error {
	_, err := DB.Exec(`
		UPDATE bookmarks SET title = ?, url = ?, description = ?, account = ?, 
		password = ?, category_id = ?, icon = ? WHERE id = ?
	`, b.Title, b.URL, b.Description, b.Account, b.Password, b.CategoryID, b.Icon, b.ID)
	return err
}

// DeleteBookmark 删除书签
func DeleteBookmark(id int64) error {
	_, err := DB.Exec(`DELETE FROM bookmarks WHERE id = ?`, id)
	return err
}

// UpdateBookmarkOrder 批量更新书签排序
func UpdateBookmarkOrder(ids []int64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range ids {
		_, err := tx.Exec(`UPDATE bookmarks SET "order" = ? WHERE id = ?`, i+1, id)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
