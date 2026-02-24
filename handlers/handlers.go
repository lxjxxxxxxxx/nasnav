package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"nasnav/database"
	"nasnav/models"
)

// WriteJSON 输出JSON格式的HTTP响应
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError 输出错误信息的JSON响应
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

// GetCategories 获取所有分类列表
func GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := database.GetCategories()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get categories")
		return
	}
	WriteJSON(w, http.StatusOK, categories)
}

// CreateCategoryRequest 创建分类的请求结构
type CreateCategoryRequest struct {
	Name string `json:"name"`
}

// CreateCategory 创建新分类
func CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "Name is required")
		return
	}

	category, err := database.CreateCategory(req.Name)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create category")
		return
	}
	WriteJSON(w, http.StatusCreated, category)
}

// UpdateCategoryRequest 更新分类的请求结构
type UpdateCategoryRequest struct {
	Name string `json:"name"`
}

// UpdateCategoryByID 根据ID更新分类名称
func UpdateCategoryByID(w http.ResponseWriter, r *http.Request, id int64) {
	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if err := database.UpdateCategory(id, req.Name); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to update category")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Category updated"})
}

// DeleteCategoryByID 根据ID删除分类
func DeleteCategoryByID(w http.ResponseWriter, r *http.Request, id int64) {
	if err := database.DeleteCategory(id); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to delete category")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Category deleted"})
}

// ReorderRequest 重排序的请求结构
type ReorderRequest struct {
	IDs []int64 `json:"ids"`
}

// ReorderCategories 重新排序分类
func ReorderCategories(w http.ResponseWriter, r *http.Request) {
	var req ReorderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := database.UpdateCategoryOrder(req.IDs); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to reorder categories")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Categories reordered"})
}

// GetBookmarks 获取书签列表，支持按分类ID筛选
func GetBookmarks(w http.ResponseWriter, r *http.Request, isAuthenticated bool) {
	categoryIDStr := r.URL.Query().Get("category_id")
	var categoryID int64
	if categoryIDStr != "" {
		var err error
		categoryID, err = strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid category ID")
			return
		}
	}

	bookmarks, err := database.GetBookmarks(categoryID, isAuthenticated)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get bookmarks")
		return
	}
	WriteJSON(w, http.StatusOK, bookmarks)
}

// CreateBookmark 创建新书签
func CreateBookmark(w http.ResponseWriter, r *http.Request) {
	var b models.Bookmark
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if b.Title == "" || b.URL == "" {
		WriteError(w, http.StatusBadRequest, "Title and URL are required")
		return
	}

	bookmark, err := database.CreateBookmark(&b)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create bookmark")
		return
	}
	WriteJSON(w, http.StatusCreated, bookmark)
}

// UpdateBookmarkByID 根据ID更新书签信息
func UpdateBookmarkByID(w http.ResponseWriter, r *http.Request, id int64) {
	var b models.Bookmark
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	b.ID = id
	if err := database.UpdateBookmark(&b); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to update bookmark")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Bookmark updated"})
}

// DeleteBookmarkByID 根据ID删除书签
func DeleteBookmarkByID(w http.ResponseWriter, r *http.Request, id int64) {
	if err := database.DeleteBookmark(id); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to delete bookmark")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Bookmark deleted"})
}

// ReorderBookmarks 重新排序书签
func ReorderBookmarks(w http.ResponseWriter, r *http.Request) {
	var req ReorderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := database.UpdateBookmarkOrder(req.IDs); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to reorder bookmarks")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Bookmarks reordered"})
}
