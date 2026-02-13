const ICONS = [
    '📁', '🏠', '💻', '🌐', '📧', '📷', '🎬', '🎵',
    '📚', '🎮', '🔧', '📊', '💼', '🔒', '☁️', '📱',
    '🎨', '✈️', '🚗', '🍕', '💰', '📰', '🎯', '⭐',
    '🔑', '💾', '🖥️', '📡', '⚡', '🔮', '🎪', '🌟'
];

let isAuthenticated = false;
let password = '';
let categories = [];
let bookmarks = [];
let currentCategoryId = 0;
let editingBookmark = null;
let editingCategory = null;
let draggedElement = null;

function getUrlPassword() {
    const params = new URLSearchParams(window.location.search);
    return params.get('pwd') || params.get('password') || '';
}

function setUrlPassword(pwd) {
    const url = new URL(window.location);
    url.searchParams.set('pwd', pwd);
    window.history.replaceState({}, '', url);
}

async function checkAuth() {
    password = getUrlPassword();
    if (!password) {
        isAuthenticated = false;
        return;
    }
    try {
        const res = await fetch(`/api/auth/check?password=${encodeURIComponent(password)}`);
        const data = await res.json();
        isAuthenticated = data.authenticated;
    } catch (e) {
        isAuthenticated = false;
    }
}

async function fetchCategories() {
    try {
        const res = await fetch('/api/categories');
        const data = await res.json();
        categories = Array.isArray(data) ? data : [];
    } catch (e) {
        categories = [];
    }
}

async function fetchBookmarks(categoryId = 0) {
    try {
        let url = '/api/bookmarks';
        if (categoryId > 0) {
            url += `?category_id=${categoryId}`;
        }
        const res = await fetch(url);
        const data = await res.json();
        bookmarks = Array.isArray(data) ? data : [];
    } catch (e) {
        bookmarks = [];
    }
}

function render() {
    const app = document.getElementById('app');
    app.innerHTML = `
        <div class="container">
            <header class="header">
                <h1>${SITE_TITLE || 'NAS导航'}</h1>
                <div class="header-actions">
                    ${isAuthenticated ? `
                        <button class="btn btn-primary" onclick="showAddCategoryModal()">
                            <span>+</span> 添加分类
                        </button>
                        <button class="btn btn-primary" onclick="showAddBookmarkModal()">
                            <span>+</span> 添加书签
                        </button>
                    ` : ''}
                </div>
            </header>
            
            <nav class="categories">
                <div class="category-tabs" id="categoryTabs">
                    ${renderCategoryTabs()}
                </div>
            </nav>
            
            <main class="bookmarks-grid" id="bookmarksGrid">
                ${renderBookmarks()}
            </main>
        </div>
        
        <div id="modalContainer"></div>
        <div id="contextMenu" class="context-menu"></div>
    `;
    
    setupDragAndDrop();
    setupContextMenu();
}

function renderCategoryTabs() {
    let html = `
        <button class="category-tab ${currentCategoryId === 0 ? 'active' : ''}" 
                data-id="0" onclick="selectCategory(0)">
            全部书签
        </button>
    `;
    
    categories.forEach(cat => {
        html += `
            <button class="category-tab ${currentCategoryId === cat.id ? 'active' : ''}" 
                    data-id="${cat.id}" 
                    draggable="${isAuthenticated ? 'true' : 'false'}"
                    onclick="selectCategory(${cat.id})">
                <span class="drag-handle">${cat.name}</span>
            </button>
        `;
    });
    
    return html;
}

function renderBookmarks() {
    if (bookmarks.length === 0) {
        return `
            <div class="empty-state">
                <div class="icon">📭</div>
                <p>暂无书签</p>
            </div>
        `;
    }
    
    return bookmarks.map(b => `
        <div class="bookmark-card" data-id="${b.id}" draggable="${isAuthenticated ? 'true' : 'false'}">
            <div class="card-header">
                <div class="card-icon">${b.icon || '📁'}</div>
                <div class="card-title">
                    <h3>${escapeHtml(b.title)}</h3>
                    <a href="${escapeHtml(b.url)}" target="_blank" onclick="event.stopPropagation()">
                        ${escapeHtml(b.url)}
                    </a>
                </div>
            </div>
            ${b.description ? `<p class="card-desc">${escapeHtml(b.description)}</p>` : ''}
            ${isAuthenticated && (b.account || b.password) ? `
                <div class="card-info">
                    ${b.account ? `<span><span class="label">账号:</span> ${escapeHtml(b.account)}</span>` : ''}
                    ${b.password ? `<span><span class="label">密码:</span> ${escapeHtml(b.password)}</span>` : ''}
                </div>
            ` : ''}
            ${isAuthenticated ? `
                <div class="card-actions">
                    <button class="btn btn-small btn-secondary" onclick="event.stopPropagation(); showEditBookmarkModal(${b.id})">编辑</button>
                    <button class="btn btn-small btn-danger" onclick="event.stopPropagation(); deleteBookmark(${b.id})">删除</button>
                </div>
            ` : ''}
        </div>
    `).join('');
}

function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

async function selectCategory(id) {
    currentCategoryId = id;
    await fetchBookmarks(id);
    render();
}

let longPressTimer = null;
let longPressTriggered = false;

function setupContextMenu() {
    if (!isAuthenticated) return;
    
    const contextMenu = document.getElementById('contextMenu');
    
    document.querySelectorAll('.category-tab[data-id]:not([data-id="0"])').forEach(tab => {
        // PC端右键菜单
        tab.addEventListener('contextmenu', (e) => {
            e.preventDefault();
            showCategoryContextMenu(e, tab);
        });
        
        // 移动端长按菜单
        tab.addEventListener('touchstart', (e) => {
            longPressTriggered = false;
            longPressTimer = setTimeout(() => {
                longPressTriggered = true;
                e.preventDefault();
                const touch = e.touches[0];
                showCategoryContextMenu({ pageX: touch.pageX, pageY: touch.pageY, preventDefault: () => {} }, tab);
            }, 500);
        }, { passive: false });
        
        tab.addEventListener('touchend', () => {
            if (longPressTimer) {
                clearTimeout(longPressTimer);
                longPressTimer = null;
            }
        });
        
        tab.addEventListener('touchmove', () => {
            if (longPressTimer) {
                clearTimeout(longPressTimer);
                longPressTimer = null;
            }
        });
    });
    
    document.addEventListener('click', hideContextMenu);
    document.addEventListener('touchstart', (e) => {
        if (!e.target.closest('.context-menu') && !e.target.closest('.category-tab')) {
            hideContextMenu();
        }
    });
}

function showCategoryContextMenu(e, tab) {
    const contextMenu = document.getElementById('contextMenu');
    const id = parseInt(tab.dataset.id);
    const cat = categories.find(c => c.id === id);
    if (!cat) return;
    
    contextMenu.innerHTML = `
        <div class="context-menu-item" onclick="showEditCategoryModal(${id}); hideContextMenu();">
            ✏️ 编辑分类
        </div>
        <div class="context-menu-item danger" onclick="deleteCategory(${id}); hideContextMenu();">
            🗑️ 删除分类
        </div>
    `;
    
    let x = e.pageX;
    let y = e.pageY;
    
    // 确保菜单不超出屏幕
    contextMenu.style.left = '0px';
    contextMenu.style.top = '0px';
    contextMenu.classList.add('show');
    
    const menuRect = contextMenu.getBoundingClientRect();
    const windowWidth = window.innerWidth;
    const windowHeight = window.innerHeight;
    
    if (x + menuRect.width > windowWidth) {
        x = windowWidth - menuRect.width - 10;
    }
    if (y + menuRect.height > windowHeight) {
        y = windowHeight - menuRect.height - 10;
    }
    
    contextMenu.style.left = x + 'px';
    contextMenu.style.top = y + 'px';
}

function hideContextMenu() {
    const contextMenu = document.getElementById('contextMenu');
    if (contextMenu) {
        contextMenu.classList.remove('show');
    }
}

function showAddCategoryModal() {
    editingCategory = null;
    showModal('添加分类', `
        <div class="form-group">
            <label>分类名称</label>
            <input type="text" id="categoryName" placeholder="请输入分类名称">
        </div>
    `, saveCategory);
}

function showEditCategoryModal(id) {
    const cat = categories.find(c => c.id === id);
    if (!cat) return;
    
    editingCategory = cat;
    showModal('编辑分类', `
        <div class="form-group">
            <label>分类名称</label>
            <input type="text" id="categoryName" value="${escapeHtml(cat.name)}" placeholder="请输入分类名称">
        </div>
    `, saveCategory);
}

async function saveCategory() {
    const name = document.getElementById('categoryName').value.trim();
    if (!name) {
        showToast('请输入分类名称', 'error');
        return;
    }
    
    try {
        let res;
        if (editingCategory) {
            res = await fetch(`/api/categories/${editingCategory.id}?password=${encodeURIComponent(password)}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name })
            });
        } else {
            res = await fetch(`/api/categories?password=${encodeURIComponent(password)}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name })
            });
        }
        
        if (!res.ok) throw new Error();
        
        closeModal();
        await fetchCategories();
        render();
        showToast(editingCategory ? '分类已更新' : '分类已创建', 'success');
    } catch (e) {
        showToast('操作失败', 'error');
    }
}

async function deleteCategory(id) {
    if (!confirm('确定要删除此分类吗？该分类下的所有书签也会被删除。')) return;
    
    try {
        const res = await fetch(`/api/categories/${id}?password=${encodeURIComponent(password)}`, {
            method: 'DELETE'
        });
        
        if (!res.ok) throw new Error();
        
        if (currentCategoryId === id) {
            currentCategoryId = 0;
        }
        
        await fetchCategories();
        await fetchBookmarks(currentCategoryId);
        render();
        showToast('分类已删除', 'success');
    } catch (e) {
        showToast('删除失败', 'error');
    }
}

function showAddBookmarkModal() {
    if (categories.length === 0) {
        showToast('请先创建分类', 'error');
        return;
    }
    
    editingBookmark = null;
    showModal('添加书签', getBookmarkFormHtml(), saveBookmark);
}

function showEditBookmarkModal(id) {
    const bookmark = bookmarks.find(b => b.id === id);
    if (!bookmark) return;
    
    editingBookmark = bookmark;
    showModal('编辑书签', getBookmarkFormHtml(bookmark), saveBookmark);
}

function getBookmarkFormHtml(b = null) {
    return `
        <div class="form-group">
            <label>标题 *</label>
            <input type="text" id="bookmarkTitle" value="${b ? escapeHtml(b.title) : ''}" placeholder="请输入标题">
        </div>
        <div class="form-group">
            <label>链接 *</label>
            <input type="url" id="bookmarkUrl" value="${b ? escapeHtml(b.url) : ''}" placeholder="请输入链接地址">
        </div>
        <div class="form-group">
            <label>描述</label>
            <textarea id="bookmarkDesc" placeholder="请输入描述（可选）">${b ? escapeHtml(b.description || '') : ''}</textarea>
        </div>
        <div class="form-group">
            <label>账号</label>
            <input type="text" id="bookmarkAccount" value="${b ? escapeHtml(b.account || '') : ''}" placeholder="请输入账号（可选）">
        </div>
        <div class="form-group">
            <label>密码</label>
            <input type="text" id="bookmarkPassword" value="${b ? escapeHtml(b.password || '') : ''}" placeholder="请输入密码（可选）">
        </div>
        <div class="form-group">
            <label>分类</label>
            <select id="bookmarkCategory">
                ${categories.map(c => `
                    <option value="${c.id}" ${b && b.category_id === c.id ? 'selected' : ''}>
                        ${escapeHtml(c.name)}
                    </option>
                `).join('')}
            </select>
        </div>
        <div class="form-group">
            <label>图标</label>
            <div class="icon-selector">
                ${ICONS.map(icon => `
                    <div class="icon-option ${b && b.icon === icon ? 'selected' : ''}" 
                         onclick="selectIcon('${icon}')" data-icon="${icon}">
                        ${icon}
                    </div>
                `).join('')}
            </div>
            <input type="hidden" id="bookmarkIcon" value="${b ? (b.icon || '📁') : '📁'}">
        </div>
    `;
}

function selectIcon(icon) {
    document.querySelectorAll('.icon-option').forEach(el => el.classList.remove('selected'));
    document.querySelector(`.icon-option[data-icon="${icon}"]`).classList.add('selected');
    document.getElementById('bookmarkIcon').value = icon;
}

async function saveBookmark() {
    const title = document.getElementById('bookmarkTitle').value.trim();
    const url = document.getElementById('bookmarkUrl').value.trim();
    const description = document.getElementById('bookmarkDesc').value.trim();
    const account = document.getElementById('bookmarkAccount').value.trim();
    const password_val = document.getElementById('bookmarkPassword').value.trim();
    const categoryId = parseInt(document.getElementById('bookmarkCategory').value);
    const icon = document.getElementById('bookmarkIcon').value;
    
    if (!title || !url) {
        showToast('标题和链接为必填项', 'error');
        return;
    }
    
    const data = {
        title,
        url,
        description,
        account,
        password: password_val,
        category_id: categoryId,
        icon
    };
    
    try {
        let res;
        if (editingBookmark) {
            res = await fetch(`/api/bookmarks/${editingBookmark.id}?password=${encodeURIComponent(password)}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
        } else {
            res = await fetch(`/api/bookmarks?password=${encodeURIComponent(password)}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
        }
        
        if (!res.ok) throw new Error();
        
        closeModal();
        await fetchBookmarks(currentCategoryId);
        render();
        showToast(editingBookmark ? '书签已更新' : '书签已创建', 'success');
    } catch (e) {
        showToast('操作失败', 'error');
    }
}

async function deleteBookmark(id) {
    if (!confirm('确定要删除此书签吗？')) return;
    
    try {
        const res = await fetch(`/api/bookmarks/${id}?password=${encodeURIComponent(password)}`, {
            method: 'DELETE'
        });
        
        if (!res.ok) throw new Error();
        
        await fetchBookmarks(currentCategoryId);
        render();
        showToast('书签已删除', 'success');
    } catch (e) {
        showToast('删除失败', 'error');
    }
}

function showModal(title, content, onSave) {
    const container = document.getElementById('modalContainer');
    container.innerHTML = `
        <div class="modal-overlay" onclick="closeModalOnOverlay(event)">
            <div class="modal" onclick="event.stopPropagation()">
                <div class="modal-header">
                    <h2>${title}</h2>
                    <button class="modal-close" onclick="closeModal()">&times;</button>
                </div>
                <div class="modal-body">
                    ${content}
                </div>
                <div class="modal-footer">
                    <button class="btn btn-secondary" onclick="closeModal()">取消</button>
                    <button class="btn btn-primary" id="modalSaveBtn">保存</button>
                </div>
            </div>
        </div>
    `;
    
    document.getElementById('modalSaveBtn').onclick = onSave;
    
    const firstInput = container.querySelector('input, textarea, select');
    if (firstInput) firstInput.focus();
}

function closeModal() {
    document.getElementById('modalContainer').innerHTML = '';
}

function closeModalOnOverlay(event) {
    if (event.target.classList.contains('modal-overlay')) {
        closeModal();
    }
}

function showToast(message, type = 'info') {
    const existing = document.querySelector('.toast');
    if (existing) existing.remove();
    
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    document.body.appendChild(toast);
    
    setTimeout(() => toast.remove(), 3000);
}

function setupDragAndDrop() {
    if (!isAuthenticated) return;
    
    setupCategoryDrag();
    setupBookmarkDrag();
}

function setupCategoryDrag() {
    const tabs = document.querySelectorAll('.category-tab[data-id]:not([data-id="0"])');
    
    tabs.forEach(tab => {
        tab.addEventListener('dragstart', (e) => {
            draggedElement = tab;
            tab.classList.add('dragging');
            e.dataTransfer.effectAllowed = 'move';
        });
        
        tab.addEventListener('dragend', () => {
            tab.classList.remove('dragging');
            draggedElement = null;
        });
        
        tab.addEventListener('dragover', (e) => {
            e.preventDefault();
            e.dataTransfer.dropEffect = 'move';
        });
        
        tab.addEventListener('drop', async (e) => {
            e.preventDefault();
            if (!draggedElement || draggedElement === tab) return;
            
            const container = document.getElementById('categoryTabs');
            const allTabs = Array.from(container.querySelectorAll('.category-tab[data-id]:not([data-id="0"])'));
            const draggedIdx = allTabs.indexOf(draggedElement);
            const targetIdx = allTabs.indexOf(tab);
            
            if (draggedIdx < targetIdx) {
                tab.parentNode.insertBefore(draggedElement, tab.nextSibling);
            } else {
                tab.parentNode.insertBefore(draggedElement, tab);
            }
            
            const newOrder = Array.from(container.querySelectorAll('.category-tab[data-id]'))
                .map(t => parseInt(t.dataset.id));
            
            try {
                await fetch(`/api/categories/reorder?password=${encodeURIComponent(password)}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ ids: newOrder })
                });
                await fetchCategories();
            } catch (e) {
                showToast('排序保存失败', 'error');
            }
        });
    });
}

function setupBookmarkDrag() {
    const cards = document.querySelectorAll('.bookmark-card');
    
    cards.forEach(card => {
        card.addEventListener('dragstart', (e) => {
            draggedElement = card;
            card.classList.add('dragging');
            e.dataTransfer.effectAllowed = 'move';
        });
        
        card.addEventListener('dragend', () => {
            card.classList.remove('dragging');
            draggedElement = null;
        });
        
        card.addEventListener('dragover', (e) => {
            e.preventDefault();
            e.dataTransfer.dropEffect = 'move';
        });
        
        card.addEventListener('drop', async (e) => {
            e.preventDefault();
            if (!draggedElement || draggedElement === card) return;
            
            const grid = document.getElementById('bookmarksGrid');
            const allCards = Array.from(grid.querySelectorAll('.bookmark-card'));
            const draggedIdx = allCards.indexOf(draggedElement);
            const targetIdx = allCards.indexOf(card);
            
            if (draggedIdx < targetIdx) {
                card.parentNode.insertBefore(draggedElement, card.nextSibling);
            } else {
                card.parentNode.insertBefore(draggedElement, card);
            }
            
            const newOrder = Array.from(grid.querySelectorAll('.bookmark-card'))
                .map(c => parseInt(c.dataset.id));
            
            try {
                await fetch(`/api/bookmarks/reorder?password=${encodeURIComponent(password)}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ ids: newOrder })
                });
                await fetchBookmarks(currentCategoryId);
            } catch (e) {
                showToast('排序保存失败', 'error');
            }
        });
    });
}

async function init() {
    await checkAuth();
    await fetchCategories();
    await fetchBookmarks();
    render();
}

init();
