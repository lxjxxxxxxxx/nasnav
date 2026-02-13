package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"nasnav/config"
	"nasnav/database"
	"nasnav/handlers"
)

//go:embed static/*
var staticFS embed.FS

var authPassword string

// getExeDir 获取可执行文件所在目录的绝对路径
func getExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	return filepath.Dir(exe)
}

func main() {
	exeDir := getExeDir()

	configPath := filepath.Join(exeDir, "config.yaml")
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	authPassword = cfg.Auth.Password

	dbPath := cfg.Database.Path
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(exeDir, dbPath)
	}

	if err := database.Init(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/categories/reorder", authWrap(handlers.ReorderCategories))
	mux.HandleFunc("/api/categories/", categoriesHandlerWithID)
	mux.HandleFunc("/api/categories", categoriesHandler)
	mux.HandleFunc("/api/bookmarks/reorder", authWrap(handlers.ReorderBookmarks))
	mux.HandleFunc("/api/bookmarks/", bookmarksHandlerWithID)
	mux.HandleFunc("/api/bookmarks", bookmarksHandler)
	mux.HandleFunc("/api/auth/check", authCheckHandler)

	staticContent, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Failed to get static files: %v", err)
	}
	fileServer := http.FileServer(http.FS(staticContent))
	mux.Handle("/", indexHandler{fileServer: fileServer, title: cfg.Site.Title})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	log.Printf("Config file: %s", configPath)
	log.Printf("Database file: %s", dbPath)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// authWrap 权限验证中间件，检查URL参数中的密码是否正确
func authWrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pwd := r.URL.Query().Get("password")
		if pwd != authPassword {
			handlers.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		next(w, r)
	}
}

// authCheckHandler 检查认证状态
func authCheckHandler(w http.ResponseWriter, r *http.Request) {
	password := r.URL.Query().Get("password")
	hasAuth := password == authPassword
	handlers.WriteJSON(w, http.StatusOK, map[string]bool{"authenticated": hasAuth})
}

// categoriesHandler 处理分类列表的GET和POST请求
func categoriesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handlers.GetCategories(w, r)
	case "POST":
		authWrap(handlers.CreateCategory)(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// categoriesHandlerWithID 处理单个分类的PUT和DELETE请求
func categoriesHandlerWithID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/api/categories/" {
		http.NotFound(w, r)
		return
	}

	idStr := strings.TrimPrefix(path, "/api/categories/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	switch r.Method {
	case "PUT":
		authWrap(func(w http.ResponseWriter, r *http.Request) {
			handlers.UpdateCategoryByID(w, r, id)
		})(w, r)
	case "DELETE":
		authWrap(func(w http.ResponseWriter, r *http.Request) {
			handlers.DeleteCategoryByID(w, r, id)
		})(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// bookmarksHandler 处理书签列表的GET和POST请求
func bookmarksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handlers.GetBookmarks(w, r)
	case "POST":
		authWrap(handlers.CreateBookmark)(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// bookmarksHandlerWithID 处理单个书签的PUT和DELETE请求
func bookmarksHandlerWithID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/api/bookmarks/" {
		http.NotFound(w, r)
		return
	}

	idStr := strings.TrimPrefix(path, "/api/bookmarks/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, "Invalid bookmark ID")
		return
	}

	switch r.Method {
	case "PUT":
		authWrap(func(w http.ResponseWriter, r *http.Request) {
			handlers.UpdateBookmarkByID(w, r, id)
		})(w, r)
	case "DELETE":
		authWrap(func(w http.ResponseWriter, r *http.Request) {
			handlers.DeleteBookmarkByID(w, r, id)
		})(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// indexHandler 首页处理器，处理静态文件服务
type indexHandler struct {
	fileServer http.Handler
	title      string
}

// ServeHTTP 处理HTTP请求，首页返回带动态标题的HTML，其他请求转发给文件服务器
func (h indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		data, err := staticFS.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(replaceTitle(string(data), h.title)))
		return
	}
	h.fileServer.ServeHTTP(w, r)
}

// replaceTitle 替换HTML模板中的标题
func replaceTitle(html, title string) string {
	return fmt.Sprintf("<!DOCTYPE html><html lang=\"zh-CN\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><title>%s</title><link rel=\"stylesheet\" href=\"/css/style.css\"></head><body><div id=\"app\"><div class=\"loading\">加载中...</div></div><script>const SITE_TITLE = \"%s\";</script><script src=\"/js/app.js\"></script></body></html>", title, title)
}
