# NAS导航

一个轻量级的NAS服务导航页面，用于统一管理和快速访问NAS上的各种服务。

## 功能特性

### 书签管理
- 以卡片形式展示书签，直观美观
- 支持标题、链接、描述、账号、密码、图标等信息
- 支持拖拽排序，自定义书签顺序
- 预设32个常用图标供选择

### 分类管理
- 选项卡形式展示分类
- 第一个固定为"全部书签"
- 支持拖拽排序分类顺序
- 删除分类时自动删除下属书签

### 权限控制
- 通过URL参数实现权限验证
- 无权限时仅可浏览
- 有权限时可添加、编辑、删除书签和分类

### 响应式设计
- 适配PC端和移动端
- 移动端支持长按分类弹出编辑菜单
- PC端支持右键菜单编辑分类

## 技术栈

- **前端**: HTML + CSS + JavaScript（无框架）
- **后端**: Go 1.21+
- **数据库**: SQLite3（纯Go驱动，无需CGO）

## 项目结构

```
nasnav/
├── config.yaml          # 配置文件
├── main.go              # 主程序入口
├── go.mod               # Go模块定义
├── config/
│   └── config.go        # 配置加载
├── database/
│   └── database.go      # 数据库操作
├── handlers/
│   └── handlers.go      # API处理器
├── models/
│   └── models.go        # 数据模型
└── static/
    ├── index.html       # HTML模板
    ├── css/
    │   └── style.css    # 样式文件
    └── js/
        └── app.js       # 前端交互逻辑
```

## 安装部署

### 编译

```bash
# 克隆或下载项目
cd nasnav

# 下载依赖
go mod tidy

# 编译
go build -o nasnav .
```

### 部署

将以下文件部署到NAS：

```
/nasnav/
├── nasnav       # 可执行文件
├── config.yaml  # 配置文件
└── data/        # 数据目录（自动创建）
    └── nasnav.db
```

### 运行

```bash
# 直接运行
./nasnav

# 后台运行
nohup ./nasnav &
```

## 配置说明

编辑 `config.yaml` 文件：

```yaml
server:
  port: 8080              # 服务端口
  host: 0.0.0.0           # 监听地址

database:
  path: ./data/nasnav.db  # 数据库路径（相对路径基于可执行文件目录）

auth:
  password: "your-secret-password"  # 访问密码

site:
  title: "NAS导航"        # 网站标题
```

### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| server.port | HTTP服务端口 | 8080 |
| server.host | 监听地址，0.0.0.0表示所有网卡 | 0.0.0.0 |
| database.path | 数据库文件路径 | ./data/nasnav.db |
| auth.password | 管理密码，用于URL参数验证 | - |
| site.title | 网站标题，显示在浏览器标签页 | NAS导航 |

## 使用方法

### 访问方式

**浏览模式（无权限）**
```
http://your-nas-ip:8080
```

**管理模式（有权限）**
```
http://your-nas-ip:8080?pwd=your-secret-password
```

### 操作说明

#### PC端操作

| 操作 | 方法 |
|------|------|
| 切换分类 | 点击分类选项卡 |
| 添加分类/书签 | 点击顶部"添加"按钮 |
| 编辑分类 | 右键点击分类选项卡 |
| 删除分类 | 右键点击分类选项卡 |
| 编辑书签 | 鼠标悬浮在卡片上，点击"编辑"按钮 |
| 删除书签 | 鼠标悬浮在卡片上，点击"删除"按钮 |
| 排序分类 | 拖拽分类选项卡 |
| 排序书签 | 拖拽书签卡片 |

#### 移动端操作

| 操作 | 方法 |
|------|------|
| 切换分类 | 点击分类选项卡 |
| 添加分类/书签 | 点击顶部"添加"按钮 |
| 编辑/删除分类 | 长按分类选项卡（约0.5秒） |
| 编辑/删除书签 | 点击卡片上的按钮 |
| 排序分类 | 长按拖拽分类选项卡 |
| 排序书签 | 长按拖拽书签卡片 |

## API接口

### 公开接口（无需权限）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/categories | 获取分类列表 |
| GET | /api/bookmarks | 获取书签列表 |
| GET | /api/auth/check | 检查认证状态 |

### 需要权限的接口

所有写操作需要在URL中添加 `password` 参数：

```
POST /api/categories?password=your-secret-password
```

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/categories | 创建分类 |
| PUT | /api/categories/{id} | 更新分类 |
| DELETE | /api/categories/{id} | 删除分类 |
| POST | /api/categories/reorder | 重排分类顺序 |
| POST | /api/bookmarks | 创建书签 |
| PUT | /api/bookmarks/{id} | 更新书签 |
| DELETE | /api/bookmarks/{id} | 删除书签 |
| POST | /api/bookmarks/reorder | 重排书签顺序 |

## 开发说明

### 本地开发

```bash
# 运行服务
go run .

# 访问
http://localhost:8080
```

### 静态资源

静态资源（HTML、CSS、JS）通过Go的embed包嵌入到可执行文件中，无需额外部署静态文件目录。

### 数据库

使用SQLite3存储数据，表结构：

**categories 表**
```sql
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    "order" INTEGER NOT NULL DEFAULT 0
);
```

**bookmarks 表**
```sql
CREATE TABLE bookmarks (
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
```

## 安全建议

1. **修改默认密码** - 部署前务必修改 `config.yaml` 中的密码
2. **使用HTTPS** - 建议配合反向代理（如Nginx）使用HTTPS
3. **内网访问** - 建议仅在内网环境使用，或通过VPN访问
4. **定期备份** - 定期备份 `data/nasnav.db` 数据库文件

## 常见问题

### Q: 如何修改端口？

A: 编辑 `config.yaml` 中的 `server.port` 配置项。

### Q: 忘记密码怎么办？

A: 查看 `config.yaml` 文件中的 `auth.password` 配置项。

### Q: 如何迁移数据？

A: 复制 `data/nasnav.db` 数据库文件到新位置即可。

### Q: 如何备份数据？

A: 复制 `data/nasnav.db` 文件到安全位置。

## 许可证

MIT License
