const API_BASE = '/api/v1';

// 状态管理
let currentPage = 1;
let currentImageId = null;
let tagsPage = 1;
let galleryPage = 1;
let gallerySort = 'date_desc';
let totalImages = 0;
let currentTags = []; // 当前编辑的标签列表
let batchUploadedImages = []; // 批量上传的图片列表
let batchDeleteMode = false; // 批量删除模式
let selectedImages = new Set(); // 选中的图片 ID 集合
let allImageIds = []; // 所有图片的 ID 列表（用于全选）

// DOM 元素
const uploadArea = document.getElementById('uploadArea');
const fileInput = document.getElementById('fileInput');
const uploadResult = document.getElementById('uploadResult');
const searchInput = document.getElementById('searchInput');
const searchBtn = document.getElementById('searchBtn');
const searchResults = document.getElementById('searchResults');
const tagsList = document.getElementById('tagsList');
const loadMoreTags = document.getElementById('loadMoreTags');
const imageModal = document.getElementById('imageModal');
const modalClose = document.querySelector('.modal-close');

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    initUpload();
    initSearch();
    initGallery();
    loadTags();
    initModal();
});

// 上传功能
function initUpload() {
    uploadArea.addEventListener('click', () => fileInput.click());
    
    uploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadArea.classList.add('dragging');
    });
    
    uploadArea.addEventListener('dragleave', () => {
        uploadArea.classList.remove('dragging');
    });
    
    uploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadArea.classList.remove('dragging');
        const files = Array.from(e.dataTransfer.files).filter(f => f.type.startsWith('image/'));
        if (files.length > 0) {
            handleFiles(files);
        }
    });
    
    fileInput.addEventListener('change', (e) => {
        const files = Array.from(e.target.files);
        if (files.length > 0) {
            handleFiles(files);
        }
    });
}

// 处理文件上传（单个或多个）
function handleFiles(files) {
    if (files.length === 1) {
        uploadSingleFile(files[0]);
    } else {
        uploadMultipleFiles(files);
    }
}

// 单文件上传
async function uploadSingleFile(file) {
    const formData = new FormData();
    formData.append('file', file);
    
    // 添加描述（如果有）
    const description = document.getElementById('uploadDescription').value.trim();
    if (description) {
        formData.append('description', description);
    }
    
    try {
        uploadResult.classList.remove('hidden', 'error');
        uploadResult.textContent = '上传中...';
        
        const response = await fetch(`${API_BASE}/images/upload`, {
            method: 'POST',
            body: formData
        });
        
        const data = await response.json();
        
        if (data.code === 201) {
            uploadResult.innerHTML = `
                <p>✅ 上传成功！</p>
                <p><strong>图片 ID:</strong> ${data.data.image_id}</p>
                <p><strong>URL:</strong> <a href="${data.data.url}" target="_blank">${data.data.url}</a></p>
                ${data.data.description ? `<p><strong>描述:</strong> ${data.data.description}</p>` : ''}
                <button onclick="showImageDetail('${data.data.image_id}')" class="btn btn-secondary" style="margin-top: 0.5rem;">修改描述或添加标签</button>
            `;
            fileInput.value = '';
            document.getElementById('uploadDescription').value = '';
            // 刷新图片列表
            loadGallery();
        } else {
            throw new Error(data.message);
        }
    } catch (error) {
        uploadResult.classList.add('error');
        uploadResult.textContent = `上传失败: ${error.message}`;
    }
}

// 批量上传
async function uploadMultipleFiles(files) {
    const formData = new FormData();
    files.forEach(file => {
        formData.append('files', file);
    });
    
    try {
        uploadResult.classList.remove('hidden', 'error');
        uploadResult.textContent = `正在上传 ${files.length} 张图片...`;
        
        const response = await fetch(`${API_BASE}/images/batch-upload`, {
            method: 'POST',
            body: formData
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            const { total, success, failed, results } = data.data;
            
            uploadResult.innerHTML = `
                <p>✅ 批量上传完成！</p>
                <p>总计: ${total} 张，成功: ${success} 张，失败: ${failed} 张</p>
                ${failed > 0 ? `<p class="error">失败的文件：${results.filter(r => r.status === 'failed').map(r => r.filename).join(', ')}</p>` : ''}
                <button onclick="showBatchEdit(${JSON.stringify(results.filter(r => r.status === 'success')).replace(/"/g, '&quot;')})" class="btn btn-primary" style="margin-top: 0.5rem;">修改描述或添加标签</button>
            `;
            
            fileInput.value = '';
            document.getElementById('uploadDescription').value = '';
            
            // 刷新图片列表
            loadGallery();
        } else {
            throw new Error(data.message);
        }
    } catch (error) {
        uploadResult.classList.add('error');
        uploadResult.textContent = `批量上传失败: ${error.message}`;
    }
}

// 搜索功能
function initSearch() {
    searchBtn.addEventListener('click', performSearch);
    searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            performSearch();
        }
    });
}

async function performSearch() {
    const tags = searchInput.value.trim();
    if (!tags) {
        alert('请输入标签');
        return;
    }
    
    const mode = document.querySelector('input[name="searchMode"]:checked').value;
    const endpoint = mode === 'exact' ? '/search/exact' : '/search/relevance';
    
    try {
        const response = await fetch(`${API_BASE}${endpoint}?tags=${encodeURIComponent(tags)}&page=1&limit=20`);
        const data = await response.json();
        
        // 两种搜索模式返回格式一致，都在 data.results 中
        const results = data.data.results;
        displayResults(results);
    } catch (error) {
        alert(`搜索失败: ${error.message}`);
    }
}

function displayResults(results) {
    if (!results || results.length === 0) {
        searchResults.innerHTML = '<p style="text-align: center; color: var(--text-muted);">没有找到相关图片</p>';
        return;
    }
    
    searchResults.innerHTML = results.map(item => `
        <div class="result-item" onclick="showImageDetail('${item.image_id || item.id}')">
            <img src="${item.url}" alt="${item.image_id || item.id}">
            <div class="result-item-info">
                <div class="result-item-tags">
                    ${item.tags ? item.tags.map(tag => `<span class="tag">${tag}</span>`).join('') : ''}
                </div>
                ${item.matched_tag_count ? `<p style="margin-top: 0.5rem; color: var(--text-muted); font-size: 0.875rem;">匹配 ${item.matched_tag_count} 个标签</p>` : ''}
            </div>
        </div>
    `).join('');
}

// Gallery 功能
function initGallery() {
    const sortSelect = document.getElementById('sortSelect');
    const prevPageBtn = document.getElementById('prevPageBtn');
    const nextPageBtn = document.getElementById('nextPageBtn');
    
    sortSelect.addEventListener('change', (e) => {
        gallerySort = e.target.value;
        galleryPage = 1;
        loadGallery();
    });
    
    prevPageBtn.addEventListener('click', () => {
        if (galleryPage > 1) {
            galleryPage--;
            loadGallery();
        }
    });
    
    nextPageBtn.addEventListener('click', () => {
        const maxPage = Math.ceil(totalImages / 20);
        if (galleryPage < maxPage) {
            galleryPage++;
            loadGallery();
        }
    });
    
    // 批量删除按钮
    document.getElementById('batchDeleteBtn').addEventListener('click', enterBatchDeleteMode);
    document.getElementById('selectAllImagesBtn').addEventListener('click', toggleSelectAllImages);
    document.getElementById('cancelBatchDeleteBtn').addEventListener('click', exitBatchDeleteMode);
    document.getElementById('confirmDeleteBtn').addEventListener('click', confirmBatchDelete);
    
    // 初始加载
    loadGallery();
}

async function loadGallery() {
    const galleryGrid = document.getElementById('galleryGrid');
    galleryGrid.innerHTML = '<p style="text-align:center;color:#64748b;">加载中...</p>';
    
    try {
        const response = await fetch(
            `${API_BASE}/images?page=${galleryPage}&limit=20&sort=${gallerySort}`
        );
        const data = await response.json();
        
        if (data.code === 200) {
            totalImages = data.data.total;
            document.getElementById('totalImages').textContent = totalImages;
            document.getElementById('pageInfo').textContent = `第 ${galleryPage} 页`;
            
            // 更新分页按钮状态
            const prevBtn = document.getElementById('prevPageBtn');
            const nextBtn = document.getElementById('nextPageBtn');
            prevBtn.disabled = galleryPage === 1;
            nextBtn.disabled = galleryPage >= Math.ceil(totalImages / 20);
            
            // 显示图片
            if (data.data.images && data.data.images.length > 0) {
                displayGallery(data.data.images);
            } else {
                galleryGrid.innerHTML = '<p style="text-align:center;color:#64748b;">暂无图片</p>';
            }
        } else {
            throw new Error(data.message);
        }
    } catch (error) {
        galleryGrid.innerHTML = `<p style="text-align:center;color:#ef4444;">加载失败: ${error.message}</p>`;
    }
}

function displayGallery(images) {
    const galleryGrid = document.getElementById('galleryGrid');
    galleryGrid.innerHTML = images.map(img => `
        <div class="result-item ${batchDeleteMode ? 'batch-delete-mode' : ''}" 
             ${batchDeleteMode ? `onclick="toggleImageSelectionByClick('${img.id}')"` : `onclick="showImageDetail('${img.id}')"`}>
            ${batchDeleteMode ? `
                <div class="batch-checkbox-wrapper">
                    <input type="checkbox" 
                           class="batch-checkbox" 
                           id="checkbox_${img.id}"
                           data-image-id="${img.id}"
                           ${selectedImages.has(img.id) ? 'checked' : ''}>
                </div>
            ` : ''}
            <img src="${img.url}" alt="${img.description || '图片'}">
            <div class="result-item-info">
                ${img.description ? `<p style="margin-bottom: 0.5rem; color: var(--text-color); font-size: 0.9rem;">${img.description}</p>` : ''}
                <div class="result-item-tags">
                    ${img.tags && img.tags.length > 0 ? img.tags.map(tag => `<span class="tag">${tag}</span>`).join('') : ''}
                </div>
                <p style="margin-top: 0.5rem; color: var(--text-muted); font-size: 0.875rem;">${formatDate(img.upload_date)}</p>
            </div>
        </div>
    `).join('');
}

function formatDate(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleDateString('zh-CN', { 
        year: 'numeric', 
        month: '2-digit', 
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}

// 标签功能
async function loadTags() {
    try {
        const response = await fetch(`${API_BASE}/tags?page=${tagsPage}&limit=50`);
        const data = await response.json();
        
        if (data.code === 200 && data.data.tags) {
            const tagsHTML = data.data.tags.map(tag => `
                <div class="tag-item" onclick="searchByTag('${tag.name}')">
                    <span>${tag.name}</span>
                    <span class="tag-count">${tag.count}</span>
                </div>
            `).join('');
            
            if (tagsPage === 1) {
                tagsList.innerHTML = tagsHTML;
            } else {
                tagsList.innerHTML += tagsHTML;
            }
            
            // 如果还有更多标签，显示加载更多按钮
            if (tagsPage * 50 < data.data.total) {
                loadMoreTags.style.display = 'block';
            } else {
                loadMoreTags.style.display = 'none';
            }
        }
    } catch (error) {
        console.error('加载标签失败:', error);
    }
}

loadMoreTags.addEventListener('click', () => {
    tagsPage++;
    loadTags();
});

function searchByTag(tag) {
    searchInput.value = tag;
    performSearch();
}

// 图片详情模态框
function initModal() {
    modalClose.addEventListener('click', closeModal);
    imageModal.addEventListener('click', (e) => {
        if (e.target === imageModal) {
            closeModal();
        }
    });
    
    document.getElementById('updateDescriptionBtn').addEventListener('click', updateDescription);
    document.getElementById('addTagBtn').addEventListener('click', addNewTags);
    document.getElementById('saveTagsBtn').addEventListener('click', saveTags);
    document.getElementById('deleteImageBtn').addEventListener('click', deleteImage);
    document.getElementById('aiGenerateSingleBtn').addEventListener('click', generateSingleImageTags);
    
    // 支持回车键添加标签
    document.getElementById('newTagInput').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            addNewTags();
        }
    });
}

async function showImageDetail(imageId) {
    currentImageId = imageId;
    
    try {
        const response = await fetch(`${API_BASE}/images/${imageId}`);
        const data = await response.json();
        
        if (data.code === 200) {
            const info = data.data;
            document.getElementById('modalImage').src = info.url;
            document.getElementById('modalImageId').textContent = info.image_id;
            document.getElementById('modalImageUrl').textContent = info.url;
            document.getElementById('modalImageUrl').href = info.url;
            document.getElementById('modalUploadDate').textContent = new Date(info.upload_date).toLocaleString();
            
            // 显示描述
            document.getElementById('modalDescription').textContent = info.description || '';
            document.getElementById('newDescription').value = info.description || '';
            
            // 初始化标签编辑器
            currentTags = info.tags ? [...info.tags] : [];
            renderEditableTags();
            document.getElementById('newTagInput').value = '';
            
            // 清空 AI 生成状态
            document.getElementById('aiPromptSingle').value = '';
            document.getElementById('aiDelimiterSingle').value = ',';
            document.getElementById('aiStatusSingle').innerHTML = '';
            
            imageModal.classList.remove('hidden');
        }
    } catch (error) {
        alert(`获取图片详情失败: ${error.message}`);
    }
}

function closeModal() {
    imageModal.classList.add('hidden');
    currentImageId = null;
}

async function updateDescription() {
    const newDescription = document.getElementById('newDescription').value.trim();
    if (!newDescription) {
        alert('请输入描述');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/images/${currentImageId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ description: newDescription })
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            alert('描述更新成功！');
            showImageDetail(currentImageId);
            loadGallery(); // 刷新图片列表
        } else {
            throw new Error(data.message);
        }
    } catch (error) {
        alert(`更新描述失败: ${error.message}`);
    }
}

// 渲染可编辑的标签列表
function renderEditableTags() {
    const container = document.getElementById('editableTags');
    
    if (currentTags.length === 0) {
        container.innerHTML = '';
        return;
    }
    
    container.innerHTML = currentTags.map((tag, index) => `
        <div class="editable-tag">
            <span>${escapeHtml(tag)}</span>
            <span class="tag-remove" data-index="${index}">✕</span>
        </div>
    `).join('');
    
    // 添加删除事件监听
    container.querySelectorAll('.tag-remove').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const index = parseInt(e.target.dataset.index);
            currentTags.splice(index, 1);
            renderEditableTags();
        });
    });
}

// HTML 转义函数，防止 XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// 添加新标签
function addNewTags() {
    const input = document.getElementById('newTagInput');
    const value = input.value.trim();
    
    if (!value) {
        return;
    }
    
    // 支持多种分隔符：逗号、分号、空格
    const newTags = value.split(/[,;，；\s]+/)
        .map(t => t.trim())
        .filter(t => t && !currentTags.includes(t)); // 去重
    
    if (newTags.length === 0) {
        alert('标签已存在或无效');
        return;
    }
    
    currentTags.push(...newTags);
    renderEditableTags();
    input.value = '';
    input.focus();
}

// 保存标签更改
async function saveTags() {
    if (currentTags.length === 0) {
        if (!confirm('确定要清空所有标签吗？')) {
            return;
        }
    }
    
    try {
        const response = await fetch(`${API_BASE}/images/${currentImageId}/tags`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ 
                tags: currentTags, 
                mode: 'set' 
            })
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            alert('标签保存成功！');
            showImageDetail(currentImageId);
            loadTags(); // 重新加载标签列表
            loadGallery(); // 刷新图片列表
        } else {
            throw new Error(data.message);
        }
    } catch (error) {
        alert(`保存标签失败: ${error.message}`);
    }
}

async function deleteImage() {
    if (!confirm('确定要删除这张图片吗？此操作不可恢复！')) {
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/images/${currentImageId}`, {
            method: 'DELETE'
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            alert('图片已删除');
            closeModal();
            // 刷新搜索结果或标签
            loadTags();
            loadGallery();
        } else {
            throw new Error(data.message);
        }
    } catch (error) {
        alert(`删除失败: ${error.message}`);
    }
}

// ========== 批量编辑功能 ==========

// 显示批量编辑模态框
function showBatchEdit(images) {
    batchUploadedImages = images;
    
    const batchEditModal = document.getElementById('batchEditModal');
    const batchTotalCount = document.getElementById('batchTotalCount');
    const batchEditList = document.getElementById('batchEditList');
    
    batchTotalCount.textContent = images.length;
    
    // 渲染批量编辑列表
    batchEditList.innerHTML = images.map((img, index) => `
        <div class="batch-edit-item" data-index="${index}">
            <div class="batch-edit-preview">
                <img src="${img.url}" alt="${img.filename}">
                <p class="batch-edit-filename">${escapeHtml(img.filename)}</p>
            </div>
            <div class="batch-edit-description">
                <label>描述</label>
                <input type="text" 
                       class="batch-description-input" 
                       data-image-id="${img.image_id}" 
                       placeholder="输入图片描述">
            </div>
            <div class="batch-edit-tags">
                <label>标签</label>
                <div class="batch-tags-input-group">
                    <input type="text" 
                           class="batch-tags-input" 
                           data-index="${index}"
                           placeholder="添加标签（逗号分隔）">
                    <button class="btn btn-icon btn-sm" onclick="addBatchTags(${index})" title="添加">➕</button>
                </div>
                <div class="batch-tags-display" id="batchTags_${index}"></div>
            </div>
        </div>
    `).join('');
    
    // 初始化每个图片的标签数组
    images.forEach((img, index) => {
        img.batchTags = [];
    });
    
    batchEditModal.classList.remove('hidden');
    
    // 绑定保存按钮
    document.getElementById('batchSaveAllBtn').onclick = saveBatchEdits;
    
    // 绑定 AI 生成按钮
    document.getElementById('aiGenerateBtn').onclick = batchGenerateTags;
}

// 关闭批量编辑模态框
function closeBatchEdit() {
    document.getElementById('batchEditModal').classList.add('hidden');
    batchUploadedImages = [];
}

// 为单个图片添加标签
function addBatchTags(index) {
    const input = document.querySelector(`.batch-tags-input[data-index="${index}"]`);
    const value = input.value.trim();
    
    if (!value) {
        return;
    }
    
    // 解析标签
    const newTags = value.split(/[,;，；\s]+/)
        .map(t => t.trim())
        .filter(t => t);
    
    if (newTags.length === 0) {
        return;
    }
    
    // 添加到图片的标签数组
    const img = batchUploadedImages[index];
    if (!img.batchTags) {
        img.batchTags = [];
    }
    
    newTags.forEach(tag => {
        if (!img.batchTags.includes(tag)) {
            img.batchTags.push(tag);
        }
    });
    
    // 更新显示
    renderBatchTags(index);
    input.value = '';
}

// 渲染批量标签
function renderBatchTags(index) {
    const container = document.getElementById(`batchTags_${index}`);
    const img = batchUploadedImages[index];
    
    if (!img.batchTags || img.batchTags.length === 0) {
        container.innerHTML = '<span class="batch-tags-empty">暂无标签</span>';
        return;
    }
    
    container.innerHTML = img.batchTags.map((tag, tagIndex) => `
        <div class="editable-tag">
            <span>${escapeHtml(tag)}</span>
            <span class="tag-remove" onclick="removeBatchTag(${index}, ${tagIndex})">✕</span>
        </div>
    `).join('');
}

// 删除批量标签
function removeBatchTag(imageIndex, tagIndex) {
    const img = batchUploadedImages[imageIndex];
    if (img.batchTags) {
        img.batchTags.splice(tagIndex, 1);
        renderBatchTags(imageIndex);
    }
}

// 保存所有批量编辑
async function saveBatchEdits() {
    const saveBtn = document.getElementById('batchSaveAllBtn');
    saveBtn.disabled = true;
    saveBtn.textContent = '保存中...';
    
    let successCount = 0;
    let errorCount = 0;
    const errors = [];
    
    try {
        // 获取所有描述输入
        const descriptionInputs = document.querySelectorAll('.batch-description-input');
        
        // 逐个保存图片信息
        for (let i = 0; i < batchUploadedImages.length; i++) {
            const img = batchUploadedImages[i];
            const descInput = descriptionInputs[i];
            const description = descInput.value.trim();
            const tags = img.batchTags || [];
            
            try {
                // 如果有描述，更新描述
                if (description) {
                    const descResponse = await fetch(`${API_BASE}/images/${img.image_id}`, {
                        method: 'PUT',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ description })
                    });
                    
                    if (!descResponse.ok) {
                        throw new Error('描述更新失败');
                    }
                }
                
                // 如果有标签，更新标签
                if (tags.length > 0) {
                    const tagsResponse = await fetch(`${API_BASE}/images/${img.image_id}/tags`, {
                        method: 'PUT',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ tags, mode: 'set' })
                    });
                    
                    if (!tagsResponse.ok) {
                        throw new Error('标签更新失败');
                    }
                }
                
                successCount++;
            } catch (error) {
                errorCount++;
                errors.push(`${img.filename}: ${error.message}`);
            }
        }
        
        // 显示结果
        if (errorCount === 0) {
            alert(`✅ 全部保存成功！共 ${successCount} 张图片`);
            closeBatchEdit();
            loadTags();
            loadGallery();
        } else {
            alert(`⚠️ 保存完成\n成功: ${successCount} 张\n失败: ${errorCount} 张\n\n失败详情:\n${errors.join('\n')}`);
        }
    } catch (error) {
        alert(`保存失败: ${error.message}`);
    } finally {
        saveBtn.disabled = false;
        saveBtn.textContent = '保存全部';
    }
}

// ==================== AI 自动生成标签功能 ====================

// 单图片 AI 生成标签
async function generateSingleImageTags() {
    const prompt = document.getElementById('aiPromptSingle').value.trim();
    const delimiter = document.getElementById('aiDelimiterSingle').value.trim() || ',';
    const btn = document.getElementById('aiGenerateSingleBtn');
    const statusDiv = document.getElementById('aiStatusSingle');
    const tagsEditor = document.querySelector('.tags-editor');
    
    if (!currentImageId) {
        statusDiv.innerHTML = '<div class="ai-status error">✗ 没有选择图片</div>';
        return;
    }
    
    // 禁用按钮和标签编辑器
    btn.disabled = true;
    btn.textContent = '生成中...';
    tagsEditor.classList.add('ai-loading-single');
    statusDiv.innerHTML = '<div class="ai-status loading">⏳ AI 正在分析图片...</div>';
    
    try {
        // 调用 AI 生成标签 API
        const response = await fetch(`${API_BASE}/images/${currentImageId}/tags/generate`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                prompt: prompt || undefined,
                delimiter: delimiter,
                mode: 'append'
            })
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            // 成功：更新标签列表
            const generatedTags = data.data.generated_tags || [];
            
            // 合并到 currentTags
            generatedTags.forEach(tag => {
                if (!currentTags.includes(tag)) {
                    currentTags.push(tag);
                }
            });
            
            // 更新显示
            renderEditableTags();
            
            // 显示成功状态
            statusDiv.innerHTML = `<div class="ai-status success">✓ 成功生成 ${generatedTags.length} 个标签: ${generatedTags.join(', ')}</div>`;
            
            // 3秒后清空状态
            setTimeout(() => {
                statusDiv.innerHTML = '';
            }, 5000);
        } else {
            throw new Error(data.message || '生成失败');
        }
    } catch (error) {
        // 失败：显示错误
        statusDiv.innerHTML = `<div class="ai-status error">✗ ${error.message}</div>`;
    } finally {
        // 恢复按钮和标签编辑器
        btn.disabled = false;
        btn.textContent = '✨ 生成';
        tagsEditor.classList.remove('ai-loading-single');
    }
}

// 并发控制辅助函数
async function runWithConcurrency(tasks, concurrency) {
    const results = [];
    const executing = [];
    
    for (const [index, task] of tasks.entries()) {
        const promise = task().then(result => ({
            index,
            result,
            success: true
        })).catch(error => ({
            index,
            error,
            success: false
        }));
        
        results.push(promise);
        
        if (concurrency <= tasks.length) {
            const executing_promise = promise.then(() => {
                executing.splice(executing.indexOf(executing_promise), 1);
            });
            executing.push(executing_promise);
            
            if (executing.length >= concurrency) {
                await Promise.race(executing);
            }
        }
    }
    
    return Promise.all(results);
}

// 批量 AI 生成标签
async function batchGenerateTags() {
    const prompt = document.getElementById('aiPrompt').value.trim();
    const delimiter = document.getElementById('aiDelimiter').value.trim() || ',';
    const concurrency = parseInt(document.getElementById('aiConcurrency').value) || 5;
    const btn = document.getElementById('aiGenerateBtn');
    
    if (batchUploadedImages.length === 0) {
        alert('没有图片需要生成标签');
        return;
    }
    
    // 验证并发数
    if (concurrency < 1 || concurrency > 10) {
        alert('并发数应在 1-10 之间');
        return;
    }
    
    // 禁用按钮
    btn.disabled = true;
    btn.textContent = `生成中(并发${concurrency})...`;
    
    let successCount = 0;
    let failCount = 0;
    
    try {
        // 为每个图片设置加载状态
        for (let i = 0; i < batchUploadedImages.length; i++) {
            const tagContainer = document.getElementById(`batchTags_${i}`).parentElement;
            tagContainer.classList.add('ai-loading');
            
            // 移除之前的状态提示
            const oldStatus = tagContainer.querySelector('.ai-status');
            if (oldStatus) {
                oldStatus.remove();
            }
        }
        
        // 创建任务数组
        const tasks = batchUploadedImages.map((img, i) => async () => {
            const tagContainer = document.getElementById(`batchTags_${i}`).parentElement;
            
            try {
                // 调用 AI 生成标签 API
                const response = await fetch(`${API_BASE}/images/${img.image_id}/tags/generate`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        prompt: prompt || undefined,
                        delimiter: delimiter,
                        mode: 'append'
                    })
                });
                
                const data = await response.json();
                
                if (data.code === 200) {
                    // 成功：更新标签列表
                    const generatedTags = data.data.generated_tags || [];
                    
                    // 合并到 batchTags
                    if (!img.batchTags) {
                        img.batchTags = [];
                    }
                    generatedTags.forEach(tag => {
                        if (!img.batchTags.includes(tag)) {
                            img.batchTags.push(tag);
                        }
                    });
                    
                    // 更新显示
                    renderBatchTags(i);
                    
                    // 显示成功状态
                    const statusDiv = document.createElement('div');
                    statusDiv.className = 'ai-status success';
                    statusDiv.textContent = `✓ 已生成 ${generatedTags.length} 个标签`;
                    tagContainer.appendChild(statusDiv);
                    
                    return { success: true };
                } else {
                    throw new Error(data.message || '生成失败');
                }
            } catch (error) {
                // 失败：显示错误
                const statusDiv = document.createElement('div');
                statusDiv.className = 'ai-status error';
                statusDiv.textContent = `✗ ${error.message}`;
                tagContainer.appendChild(statusDiv);
                
                throw error;
            } finally {
                // 移除加载状态
                tagContainer.classList.remove('ai-loading');
            }
        });
        
        // 并发执行任务
        const results = await runWithConcurrency(tasks, concurrency);
        
        // 统计结果
        results.forEach(result => {
            if (result.success) {
                successCount++;
            } else {
                failCount++;
            }
        });
        
        // 显示总结
        if (failCount === 0) {
            alert(`✅ 全部生成成功！共为 ${successCount} 张图片生成了标签`);
        } else {
            alert(`⚠️ 生成完成\n成功: ${successCount} 张\n失败: ${failCount} 张`);
        }
    } catch (error) {
        alert(`批量生成失败: ${error.message}`);
    } finally {
        btn.disabled = false;
        btn.textContent = '✨ 自动生成标签';
    }
}

// ==================== 批量删除功能 ====================

// 进入批量删除模式
function enterBatchDeleteMode() {
    batchDeleteMode = true;
    selectedImages.clear();
    allImageIds = []; // 重置全选缓存
    
    // 重置全选按钮文本
    const selectAllBtn = document.getElementById('selectAllImagesBtn');
    if (selectAllBtn) {
        selectAllBtn.textContent = '全选';
    }
    
    // 显示/隐藏相关按钮
    document.getElementById('batchDeleteBtn').classList.add('hidden');
    document.getElementById('batchDeleteActions').classList.remove('hidden');
    
    // 更新选中计数
    updateSelectedCount();
    
    // 重新渲染图片列表（显示复选框）
    loadGallery();
}

// 退出批量删除模式
function exitBatchDeleteMode() {
    batchDeleteMode = false;
    selectedImages.clear();
    allImageIds = []; // 清空全选缓存
    
    // 显示/隐藏相关按钮
    document.getElementById('batchDeleteBtn').classList.remove('hidden');
    document.getElementById('batchDeleteActions').classList.add('hidden');
    
    // 重新渲染图片列表（隐藏复选框）
    loadGallery();
}

// 通过点击图片切换选择状态
function toggleImageSelectionByClick(imageId) {
    const checkbox = document.getElementById(`checkbox_${imageId}`);
    if (checkbox) {
        // 切换复选框状态
        checkbox.checked = !checkbox.checked;
        
        // 更新选中集合
        if (checkbox.checked) {
            selectedImages.add(imageId);
        } else {
            selectedImages.delete(imageId);
        }
        updateSelectedCount();
    }
}

// 切换图片选择状态（保留用于其他场景）
function toggleImageSelection(imageId, checked) {
    if (checked) {
        selectedImages.add(imageId);
    } else {
        selectedImages.delete(imageId);
    }
    updateSelectedCount();
}

// 更新选中计数显示
function updateSelectedCount() {
    document.getElementById('selectedCount').textContent = selectedImages.size;
    
    // 更新全选按钮文本
    const selectAllBtn = document.getElementById('selectAllImagesBtn');
    if (selectAllBtn && allImageIds.length > 0) {
        const allSelected = allImageIds.every(id => selectedImages.has(id));
        selectAllBtn.textContent = allSelected ? '取消全选' : '全选';
    }
}

// 全选/取消全选所有图片
async function toggleSelectAllImages() {
    const btn = document.getElementById('selectAllImagesBtn');
    
    // 如果还没有加载所有图片 ID，先加载
    if (allImageIds.length === 0) {
        btn.disabled = true;
        btn.textContent = '加载中...';
        
        try {
            // 获取所有图片（不分页）
            const response = await fetch(`${API_BASE}/images?page=1&limit=10000`);
            const data = await response.json();
            
            if (data.code === 200 && data.data.images) {
                allImageIds = data.data.images.map(img => img.id);
            } else {
                alert('获取图片列表失败');
                return;
            }
        } catch (error) {
            alert(`加载失败: ${error.message}`);
            return;
        } finally {
            btn.disabled = false;
        }
    }
    
    // 检查是否所有图片都已选中
    const allSelected = allImageIds.every(id => selectedImages.has(id));
    
    if (allSelected) {
        // 取消全选：清空所有选中
        selectedImages.clear();
    } else {
        // 全选：添加所有图片
        allImageIds.forEach(id => selectedImages.add(id));
    }
    
    // 更新当前页的复选框状态
    allImageIds.forEach(id => {
        const checkbox = document.getElementById(`checkbox_${id}`);
        if (checkbox) {
            checkbox.checked = selectedImages.has(id);
        }
    });
    
    updateSelectedCount();
}

// 确认批量删除
async function confirmBatchDelete() {
    if (selectedImages.size === 0) {
        alert('请先选择要删除的图片');
        return;
    }
    
    if (!confirm(`确定要删除选中的 ${selectedImages.size} 张图片吗？\n此操作不可恢复！`)) {
        return;
    }
    
    const confirmBtn = document.getElementById('confirmDeleteBtn');
    confirmBtn.disabled = true;
    confirmBtn.textContent = '删除中...';
    
    try {
        const imageIds = Array.from(selectedImages);
        
        const response = await fetch(`${API_BASE}/images/batch-delete`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ image_ids: imageIds })
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            const { total, success, failed } = data.data;
            
            if (failed === 0) {
                alert(`✅ 删除成功！共删除 ${success} 张图片`);
            } else {
                const failedList = data.data.results
                    .filter(r => r.status === 'failed')
                    .map(r => `${r.image_id}: ${r.error}`)
                    .join('\n');
                alert(`⚠️ 删除完成\n成功: ${success} 张\n失败: ${failed} 张\n\n失败详情:\n${failedList}`);
            }
            
            // 退出批量删除模式并刷新
            exitBatchDeleteMode();
            loadTags();
        } else {
            throw new Error(data.message);
        }
    } catch (error) {
        alert(`批量删除失败: ${error.message}`);
    } finally {
        confirmBtn.disabled = false;
        confirmBtn.textContent = '删除选中';
    }
}

