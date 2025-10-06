let currentUser = null;
let currentToken = null;
let currentItemId = null;

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function () {
    checkAuthStatus();
});

// Проверка статуса аутентификации
function checkAuthStatus() {
    const token = localStorage.getItem('jwtToken');
    const user = localStorage.getItem('user');

    if (token && user) {
        currentToken = token;
        currentUser = JSON.parse(user);
        showUserInterface();
    }
}

// API вызовы
async function apiCall(url, options = {}) {
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };

    if (currentToken) {
        headers['Authorization'] = `Bearer ${currentToken}`;
    }

    try {
        const response = await fetch(url, {
            ...options,
            headers
        });

        if (response.status === 401) {
            logout();
            throw new Error('Требуется авторизация');
        }

        if (response.status === 403) {
            throw new Error('Недостаточно прав для выполнения операции');
        }

        // Для статусов 204 (No Content) не пытаемся парсить JSON
        if (response.status === 204) {
            return null;
        }

        // Проверяем наличие контента перед парсингом JSON
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || 'Ошибка сервера');
            }

            return data;
        } else {
            // Если ответ не JSON, но статус успешный
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return await response.text();
        }
    } catch (error) {
        showNotification(error.message, true);
        throw error;
    }
}

// Авторизация
async function login() {
    const username = document.getElementById('login-username').value;
    const role = document.getElementById('login-role').value;

    if (!username || !role) {
        showNotification('Заполните все поля', true);
        return;
    }

    try {
        const data = await apiCall('/login', {
            method: 'POST',
            body: JSON.stringify({ username, role })
        });

        currentToken = data.token;
        currentUser = {
            username: data.username,
            role: data.role
        };

        localStorage.setItem('jwtToken', currentToken);
        localStorage.setItem('user', JSON.stringify(currentUser));

        showNotification(`Добро пожаловать, ${username}!`);
        showUserInterface();
    } catch (error) {
        console.error('Login error:', error);
    }
}

// Выход из системы
function logout() {
    localStorage.removeItem('jwtToken');
    localStorage.removeItem('user');
    currentToken = null;
    currentUser = null;

    document.getElementById('user-info').style.display = 'none';
    document.getElementById('login-form').style.display = 'block';
    document.getElementById('items-section').style.display = 'none';
    document.getElementById('history-section').style.display = 'none';
    document.getElementById('main-nav').style.display = 'none';

    // Сбрасываем форму логина
    document.getElementById('login-username').value = '';
    document.getElementById('login-role').value = '';

    showNotification('Вы вышли из системы');
}

// Показать интерфейс пользователя
function showUserInterface() {
    document.getElementById('login-form').style.display = 'none';
    document.getElementById('user-info').style.display = 'block';
    document.getElementById('user-name').textContent = currentUser.username;
    document.getElementById('user-role').textContent = currentUser.role;
    document.getElementById('main-nav').style.display = 'flex';

    // Показываем кнопки в зависимости от роли
    const createBtn = document.getElementById('create-item-btn');
    createBtn.style.display = currentUser.role === 'admin' ? 'block' : 'none';

    // Скрываем секцию истории для viewer
    const historyNavBtn = document.querySelector('.nav-btn[onclick*="history-section"]');
    if (currentUser.role === 'viewer') {
        historyNavBtn.style.display = 'none';
    } else {
        historyNavBtn.style.display = 'block';
    }

    showSection('items-section');
    loadItems();
}

// Переключение секций
function showSection(sectionName) {
    // Скрываем все секции
    document.getElementById('items-section').style.display = 'none';
    document.getElementById('history-section').style.display = 'none';

    // Показываем выбранную секцию
    document.getElementById(sectionName).style.display = 'block';

    // Обновляем активную кнопку навигации
    document.querySelectorAll('.nav-btn').forEach(btn => {
        btn.classList.remove('active');
        // Показываем все кнопки навигации (кроме скрытых для viewer)
        if (btn.style.display !== 'none') {
            btn.style.display = 'block';
        }
    });
    event.target.classList.add('active');

    // Загружаем данные для секции
    if (sectionName === 'items-section') {
        loadItems();
    } else if (sectionName === 'history-section') {
        loadHistory();
    }
}

// Работа с товарами
async function loadItems() {
    try {
        showLoading('items-list', 'Загрузка товаров...');
        const items = await apiCall('/items');
        displayItems(items);
    } catch (error) {
        console.error('Error loading items:', error);
        document.getElementById('items-list').innerHTML = '<p style="text-align: center; color: #7f8c8d;">Ошибка загрузки товаров</p>';
    }
}

function displayItems(items) {
    const itemsList = document.getElementById('items-list');
    itemsList.innerHTML = '';

    if (items.length === 0) {
        itemsList.innerHTML = '<p style="text-align: center; color: #7f8c8d;">Товары не найдены</p>';
        return;
    }

    items.forEach(item => {
        const itemCard = document.createElement('div');
        itemCard.className = 'item-card';
        itemCard.innerHTML = `
            <h3>${item.name}</h3>
            <div class="price">${formatPrice(item.price)}</div>
            <div class="meta">
                Создано: ${formatDate(item.created_at)}<br>
                Автор: ${item.created_by}
            </div>
            <div class="actions">
                <button onclick="showItemDetails(${item.id})">Просмотр</button>
                <button onclick="showItemHistoryModal(${item.id}, '${item.name}')">История</button>
            </div>
        `;
        itemsList.appendChild(itemCard);
    });
}

async function showItemDetails(itemId) {
    try {
        showLoading('item-info', 'Загрузка информации...');
        const item = await apiCall(`/items/${itemId}`);
        const itemInfo = document.getElementById('item-info');

        itemInfo.innerHTML = `
            <h4>${item.name}</h4>
            <div class="price">${formatPrice(item.price)}</div>
            <div class="meta">
                <p><strong>ID:</strong> ${item.id}</p>
                <p><strong>Создано:</strong> ${formatDate(item.created_at)}</p>
                <p><strong>Обновлено:</strong> ${formatDate(item.updated_at)}</p>
                <p><strong>Автор:</strong> ${item.created_by}</p>
            </div>
        `;

        // Показываем кнопки действий в зависимости от роли
        const updateBtn = document.getElementById('update-item-btn');
        const deleteBtn = document.getElementById('delete-item-btn');

        updateBtn.style.display = (currentUser.role === 'admin' || currentUser.role === 'manager') ? 'block' : 'none';
        deleteBtn.style.display = currentUser.role === 'admin' ? 'block' : 'none';

        currentItemId = itemId;
        document.getElementById('item-details').style.display = 'flex';
    } catch (error) {
        console.error('Error loading item details:', error);
    }
}

function hideItemDetails() {
    document.getElementById('item-details').style.display = 'none';
    currentItemId = null;
}

// Создание товара
function showCreateItemForm() {
    document.getElementById('create-item-form').style.display = 'block';
}

function hideCreateItemForm() {
    document.getElementById('create-item-form').style.display = 'none';
    document.getElementById('item-name').value = '';
    document.getElementById('item-price').value = '';
}

async function createItem() {
    const name = document.getElementById('item-name').value;
    const price = document.getElementById('item-price').value;

    if (!name || !price) {
        showNotification('Заполните все поля', true);
        return;
    }

    if (parseInt(price) <= 0) {
        showNotification('Цена должна быть положительным числом', true);
        return;
    }

    try {
        const response = await apiCall('/items', {
            method: 'POST',
            body: JSON.stringify({
                name: name,
                price: parseInt(price)
            })
        });

        showNotification('Товар успешно создан!');
        hideCreateItemForm();
        loadItems();
        return response;
    } catch (error) {
        console.error('Error creating item:', error);
        throw error;
    }
}

// Редактирование товара
function showUpdateItemForm() {
    document.getElementById('update-item-form').style.display = 'flex';
    // Заполняем форму текущими значениями
    const itemName = document.querySelector('#item-info h4').textContent;
    const itemPrice = document.querySelector('#item-info .price').textContent.replace(/[^\d.]/g, '');
    document.getElementById('update-item-name').value = itemName;
    document.getElementById('update-item-price').value = Math.round(parseFloat(itemPrice) * 100); // Конвертируем обратно в копейки
}

function hideUpdateItemForm() {
    document.getElementById('update-item-form').style.display = 'none';
}

async function updateItem() {
    const name = document.getElementById('update-item-name').value;
    const price = document.getElementById('update-item-price').value;

    if (!name || !price) {
        showNotification('Заполните все поля', true);
        return;
    }

    if (parseInt(price) <= 0) {
        showNotification('Цена должна быть положительным числом', true);
        return;
    }

    try {
        await apiCall(`/items/${currentItemId}`, {
            method: 'PUT',
            body: JSON.stringify({
                name: name,
                price: parseInt(price)
            })
        });

        showNotification('Товар успешно обновлен!');
        hideUpdateItemForm();
        hideItemDetails();
        loadItems();
    } catch (error) {
        console.error('Error updating item:', error);
    }
}

// Удаление товара
async function deleteItemPrompt() {
    if (confirm('Вы уверены, что хотите удалить этот товар?')) {
        try {
            const response = await apiCall(`/items/${currentItemId}`, {
                method: 'DELETE'
            });

            // response может быть null для статуса 204
            showNotification('Товар успешно удален!');
            hideItemDetails();
            loadItems();
        } catch (error) {
            console.error('Error deleting item:', error);
        }
    }
}
// История товара (модальное окно)
async function showItemHistoryModal(itemId, itemName) {
    try {
        // Создаем модальное окно для истории товара
        const modal = document.createElement('div');
        modal.className = 'modal';
        modal.id = 'item-history-modal';
        modal.style.display = 'flex';
        modal.innerHTML = `
            <div class="modal-content" style="max-width: 600px;">
                <div class="modal-header">
                    <h3>История товара: ${itemName}</h3>
                    <button class="close-btn" onclick="closeItemHistoryModal()">×</button>
                </div>
                <div id="item-history-content" style="max-height: 400px; overflow-y: auto; margin: 20px 0;">
                    <div style="text-align: center; color: #7f8c8d; padding: 20px;">Загрузка истории...</div>
                </div>
                <div class="form-actions">
                    <button class="secondary" onclick="closeItemHistoryModal()">Закрыть</button>
                </div>
            </div>
        `;
        document.body.appendChild(modal);

        // Загружаем историю товара
        const history = await apiCall(`/history/item/${itemId}`);
        displayItemHistory(history, itemId);
    } catch (error) {
        console.error('Error loading item history:', error);
        const content = document.getElementById('item-history-content');
        if (content) {
            content.innerHTML = `<p style="text-align: center; color: #e74c3c;">Ошибка загрузки истории: ${error.message}</p>`;
        }
    }
}

function displayItemHistory(history, itemId) {
    const content = document.getElementById('item-history-content');
    if (!content) return;

    content.innerHTML = '';

    if (history.length === 0) {
        content.innerHTML = '<p style="text-align: center; color: #7f8c8d;">История изменений не найдена</p>';
        return;
    }

    history.forEach(record => {
        const historyItem = document.createElement('div');
        historyItem.className = `history-item ${record.action.toLowerCase()}`;
        historyItem.style.margin = '10px 0';
        historyItem.style.padding = '15px';
        historyItem.style.borderLeft = `4px solid ${getActionColor(record.action)}`;
        historyItem.style.background = '#f8f9fa';
        historyItem.style.borderRadius = '0 5px 5px 0';

        const actionText = getActionText(record.action);

        historyItem.innerHTML = `
            <div class="action" style="font-weight: bold; margin-bottom: 8px; color: ${getActionColor(record.action)}">
                ${actionText}
            </div>
            <div style="margin-bottom: 5px;">
                <strong>Пользователь:</strong> ${record.changed_by}
            </div>
            <div style="color: #7f8c8d; font-size: 12px;">
                <strong>Время:</strong> ${formatDate(record.changed_at)}
            </div>
        `;
        content.appendChild(historyItem);
    });
}

function closeItemHistoryModal() {
    const modal = document.getElementById('item-history-modal');
    if (modal) {
        modal.remove();
    }
}

// Работа с общей историей (на странице истории)
async function loadHistory() {
    try {
        // Для viewer скрываем всю историю
        if (currentUser.role === 'viewer') {
            document.getElementById('history-list').innerHTML = '<p style="text-align: center; color: #7f8c8d;">Просмотр всей истории доступен только администраторам и менеджерам</p>';
            return;
        }

        showLoading('history-list', 'Загрузка истории...');
        const changedBy = document.getElementById('filter-changed-by').value;
        const action = document.getElementById('filter-action').value;

        const filters = {};
        if (changedBy) filters.changed_by = changedBy;
        if (action) filters.action = action;

        let history = [];

        // Только admin и manager могут смотреть всю историю
        if (currentUser.role === 'admin' || currentUser.role === 'manager') {
            history = await apiCall('/history?' + new URLSearchParams(filters));
        }

        displayAllHistory(history);
    } catch (error) {
        console.error('Error loading history:', error);
        if (error.message.includes('403')) {
            document.getElementById('history-list').innerHTML = '<p style="text-align: center; color: #7f8c8d;">Недостаточно прав для просмотра всей истории</p>';
        } else {
            document.getElementById('history-list').innerHTML = '<p style="text-align: center; color: #7f8c8d;">Ошибка загрузки истории</p>';
        }
    }
}

function displayAllHistory(history) {
    const historyList = document.getElementById('history-list');
    historyList.innerHTML = '';

    if (history.length === 0) {
        historyList.innerHTML = '<p style="text-align: center; color: #7f8c8d;">История изменений не найдена</p>';
        return;
    }

    history.forEach(record => {
        const historyItem = document.createElement('div');
        historyItem.className = `history-item ${record.action.toLowerCase()}`;

        const actionText = getActionText(record.action);
        const actionColor = getActionColor(record.action);

        historyItem.innerHTML = `
            <div class="action" style="color: ${actionColor}">${actionText}</div>
            <div><strong>Товар ID:</strong> ${record.item_id}</div>
            <div class="meta">
                <strong>Пользователь:</strong> ${record.changed_by}<br>
                <strong>Время:</strong> ${formatDate(record.changed_at)}
            </div>
        `;
        historyList.appendChild(historyItem);
    });
}

// Вспомогательные функции
function formatPrice(price) {
    // Конвертируем копейки в рубли
    const rubles = (price / 100).toFixed(2);
    return `${rubles} руб.`;
}

function formatDate(dateString) {
    try {
        const date = new Date(dateString);
        if (isNaN(date.getTime())) {
            return 'Неверная дата';
        }
        return date.toLocaleString('ru-RU', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
    } catch (error) {
        return 'Неверная дата';
    }
}

function getActionText(action) {
    const actions = {
        'CREATE': 'Создание товара',
        'UPDATE': 'Изменение товара',
        'DELETE': 'Удаление товара'
    };
    return actions[action] || action;
}

function getActionColor(action) {
    const colors = {
        'CREATE': '#27ae60',
        'UPDATE': '#f39c12',
        'DELETE': '#e74c3c'
    };
    return colors[action] || '#3498db';
}

function showNotification(message, isError = false) {
    const notification = document.getElementById('notification');
    const notificationText = document.getElementById('notification-text');

    notification.style.background = isError ? '#e74c3c' : '#27ae60';
    notificationText.textContent = message;
    notification.style.display = 'block';

    setTimeout(hideNotification, 5000);
}

function hideNotification() {
    document.getElementById('notification').style.display = 'none';
}

function showLoading(elementId, message = 'Загрузка...') {
    const element = document.getElementById(elementId);
    if (element) {
        element.innerHTML = `<div style="text-align: center; color: #7f8c8d; padding: 20px;">${message}</div>`;
    }
}

// Обработчики клавиш
document.addEventListener('DOMContentLoaded', function () {
    // Обработчик Enter для формы логина
    document.getElementById('login-username').addEventListener('keypress', function (e) {
        if (e.key === 'Enter') {
            login();
        }
    });

    document.getElementById('login-role').addEventListener('keypress', function (e) {
        if (e.key === 'Enter') {
            login();
        }
    });

    // Обработчики Enter для форм товаров
    document.getElementById('item-name')?.addEventListener('keypress', function (e) {
        if (e.key === 'Enter') {
            createItem();
        }
    });

    document.getElementById('item-price')?.addEventListener('keypress', function (e) {
        if (e.key === 'Enter') {
            createItem();
        }
    });

    document.getElementById('update-item-name')?.addEventListener('keypress', function (e) {
        if (e.key === 'Enter') {
            updateItem();
        }
    });

    document.getElementById('update-item-price')?.addEventListener('keypress', function (e) {
        if (e.key === 'Enter') {
            updateItem();
        }
    });

    // Обработчики для фильтров истории
    document.getElementById('filter-changed-by')?.addEventListener('keypress', function (e) {
        if (e.key === 'Enter') {
            loadHistory();
        }
    });

    document.getElementById('filter-action')?.addEventListener('change', function (e) {
        loadHistory();
    });
});

// Закрытие модальных окон по клику вне области
document.addEventListener('DOMContentLoaded', function () {
    // Закрытие модальных окон при клике вне контента
    document.addEventListener('click', function (e) {
        if (e.target.classList.contains('modal')) {
            e.target.style.display = 'none';
        }
    });

    // Предотвращение закрытия при клике на контент модального окна
    document.querySelectorAll('.modal-content').forEach(content => {
        content.addEventListener('click', function (e) {
            e.stopPropagation();
        });
    });
});