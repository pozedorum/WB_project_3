const API_BASE = 'http://localhost:8080/comments';
let currentParentId = '';

// Инициализация
document.addEventListener('DOMContentLoaded', function() {
    loadComments();
    setupEventListeners();
});

function setupEventListeners() {
    // Отправка формы
    document.getElementById('commentForm').addEventListener('submit', function(e) {
        e.preventDefault();
        addComment();
    });

    // Поиск по Enter
    document.getElementById('searchInput').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            searchComments();
        }
    });
}

// Загрузка комментариев
async function loadComments(parentId = '') {
    try {
        let url = parentId ? `${API_BASE}/${parentId}` : `${API_BASE}/all`;
        
        const response = await fetch(url);
        const data = await response.json();

        displayComments(data.comments, parentId);
        currentParentId = parentId;
        
    } catch (error) {
        console.error('Ошибка загрузки комментариев:', error);
        alert('Ошибка загрузки комментариев');
    }
}

// Отображение комментариев
function displayComments(comments, parentId) {
    const container = document.getElementById('commentsList');
    
    if (parentId) {
        // Показываем только детей конкретного комментария
        container.innerHTML = '';
    } else {
        // Показываем все дерево
        container.innerHTML = '';
    }

    comments.forEach(comment => {
        const commentElement = createCommentElement(comment);
        container.appendChild(commentElement);
    });

    // Добавляем кнопку "Назад" если мы внутри дерева
    if (parentId) {
        const backButton = document.createElement('button');
        backButton.textContent = '← Назад ко всем комментариям';
        backButton.onclick = () => loadComments('');
        backButton.style.marginBottom = '20px';
        container.insertBefore(backButton, container.firstChild);
    }
}

// Создание элемента комментария
function createCommentElement(comment) {
    const div = document.createElement('div');
    div.className = `comment comment-level-${comment.level || 0}`;
    
    div.innerHTML = `
        <div class="comment-header">
            <span class="author">${escapeHtml(comment.author)}</span>
        </div>
        <div class="comment-content">${escapeHtml(comment.content)}</div>
        <div class="comment-actions">
            <button onclick="replyToComment('${comment.id}', '${escapeHtml(comment.author)}')">Ответить</button>
            <button class="danger" onclick="deleteComment('${comment.id}')">Удалить</button>
            <button onclick="loadComments('${comment.id}')">Показать ответы</button>
        </div>
    `;
    
    return div;
}

// Добавление комментария
async function addComment() {
    const author = document.getElementById('author').value;
    const content = document.getElementById('content').value;
    const parentId = document.getElementById('parentId').value || '';

    if (!author || !content) {
        alert('Заполните все поля');
        return;
    }

    try {
        const response = await fetch(API_BASE, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                author: author,
                content: content,
                parent_id: parentId || undefined
            })
        });

        if (response.ok) {
            // Сброс формы
            document.getElementById('commentForm').reset();
            document.getElementById('parentId').value = '';
            document.getElementById('cancelReplyBtn').classList.add('hidden');
            
            // Убираем индикатор ответа
            const indicator = document.getElementById('replyingIndicator');
            if (indicator) {
                indicator.remove();
            }
            
            // Перезагрузка комментариев
            loadComments(currentParentId);
            alert('Комментарий добавлен!');
        } else {
            throw new Error('Ошибка добавления комментария');
        }
    } catch (error) {
        console.error('Ошибка:', error);
        alert('Ошибка добавления комментария');
    }
}

// Ответ на комментарий
function replyToComment(parentId, author) {
    document.getElementById('parentId').value = parentId;
    document.getElementById('cancelReplyBtn').classList.remove('hidden');
    
    // Показать индикатор ответа
    let indicator = document.getElementById('replyingIndicator');
    if (!indicator) {
        indicator = document.createElement('div');
        indicator.className = 'replying-to';
        indicator.id = 'replyingIndicator';
        
        const form = document.getElementById('commentForm');
        form.parentNode.insertBefore(indicator, form);
    }
    
    indicator.innerHTML = `Ответ на комментарий ${author}`;
    
    // Прокрутка к форме
    document.getElementById('commentForm').scrollIntoView({ behavior: 'smooth' });
}

// Отмена ответа
function cancelReply() {
    document.getElementById('parentId').value = '';
    document.getElementById('cancelReplyBtn').classList.add('hidden');
    
    const indicator = document.getElementById('replyingIndicator');
    if (indicator) {
        indicator.remove();
    }
}

// Удаление комментария
async function deleteComment(commentId) {
    if (!confirm('Удалить комментарий и все ответы?')) return;

    try {
        const response = await fetch(`${API_BASE}/${commentId}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            loadComments(currentParentId);
            alert('Комментарий удален!');
        } else {
            throw new Error('Ошибка удаления комментария');
        }
    } catch (error) {
        console.error('Ошибка:', error);
        alert('Ошибка удаления комментария');
    }
}

// Поиск комментариев
async function searchComments() {
    const query = document.getElementById('searchInput').value.trim();
    
    if (!query) {
        alert('Введите поисковый запрос');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/search?q=${encodeURIComponent(query)}`);
        const data = await response.json();

        // Показать результаты поиска
        document.getElementById('searchResults').classList.remove('hidden');
        document.getElementById('commentsTree').classList.add('hidden');
        
        const resultsContainer = document.getElementById('searchResultsList');
        resultsContainer.innerHTML = '';
        
        if (data.results.length === 0) {
            resultsContainer.innerHTML = '<p>Ничего не найдено</p>';
        } else {
            data.results.forEach(comment => {
                const commentElement = createCommentElement(comment);
                resultsContainer.appendChild(commentElement);
            });
        }
        
    } catch (error) {
        console.error('Ошибка поиска:', error);
        alert('Ошибка поиска');
    }
}

// Очистка поиска
function clearSearch() {
    document.getElementById('searchInput').value = '';
    document.getElementById('searchResults').classList.add('hidden');
    document.getElementById('commentsTree').classList.remove('hidden');
    loadComments(currentParentId);
}

// Вспомогательная функция для экранирования HTML
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}