const API_BASE = '/api/v1';

// 状态管理
let currentPage = 1;
let currentImageId = null;
let tagsPage = 1;

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
        const file = e.dataTransfer.files[0];
        if (file && file.type.startsWith('image/')) {
            uploadFile(file);
        }
    });
    
    fileInput.addEventListener('change', (e) => {
        const file = e.target.files[0];
        if (file) {
            uploadFile(file);
        }
    });
}

async function uploadFile(file) {
    const formData = new FormData();
    formData.append('file', file);
    
    try {
        uploadResult.classList.remove('hidden', 'error');
        uploadResult.textContent = '上传中...';
        
        const response = await fetch(`${API_BASE}/images/upload`, {
            method: 'POST',
            body: formData
        });
        
        const data = await response.json();
        
        if (data.code === 201) {
            uploadResult.textContent = `上传成功！图片 ID: ${data.data.image_id}`;
            uploadResult.innerHTML = `
                <p>✅ 上传成功！</p>
                <p><strong>图片 ID:</strong> ${data.data.image_id}</p>
                <p><strong>URL:</strong> <a href="${data.data.url}" target="_blank">${data.data.url}</a></p>
                <button onclick="showImageDetail('${data.data.image_id}')" class="btn btn-secondary" style="margin-top: 0.5rem;">添加标签</button>
            `;
            fileInput.value = '';
        } else {
            throw new Error(data.message);
        }
    } catch (error) {
        uploadResult.classList.add('error');
        uploadResult.textContent = `上传失败: ${error.message}`;
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
    const endpoint = mode === 'exact' ? '/search/exact' : '/mcp/v1/search/relevance';
    
    try {
        const response = await fetch(`${API_BASE}${endpoint}?tags=${encodeURIComponent(tags)}&page=1&limit=20`);
        const data = await response.json();
        
        const results = mode === 'exact' ? data.data.results : data.results;
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
    
    document.getElementById('setTagsBtn').addEventListener('click', () => updateTags('set'));
    document.getElementById('appendTagsBtn').addEventListener('click', () => updateTags('append'));
    document.getElementById('deleteImageBtn').addEventListener('click', deleteImage);
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
            
            const tagsHTML = info.tags && info.tags.length > 0
                ? info.tags.map(tag => `<span class="tag">${tag}</span>`).join('')
                : '<p style="color: var(--text-muted);">暂无标签</p>';
            document.getElementById('modalTags').innerHTML = tagsHTML;
            
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

async function updateTags(mode) {
    const newTags = document.getElementById('newTags').value.trim();
    if (!newTags) {
        alert('请输入标签');
        return;
    }
    
    const tags = newTags.split(',').map(t => t.trim()).filter(t => t);
    
    try {
        const response = await fetch(`${API_BASE}/images/${currentImageId}/tags`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ tags, mode })
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            alert('标签更新成功！');
            document.getElementById('newTags').value = '';
            showImageDetail(currentImageId);
            loadTags(); // 重新加载标签列表
        } else {
            throw new Error(data.message);
        }
    } catch (error) {
        alert(`更新标签失败: ${error.message}`);
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
        } else {
            throw new Error(data.message);
        }
    } catch (error) {
        alert(`删除失败: ${error.message}`);
    }
}

