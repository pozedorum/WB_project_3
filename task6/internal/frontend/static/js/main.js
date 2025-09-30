const API_BASE = 'http://localhost:8080/api';

// DOM Elements
const saleForm = document.getElementById('sale-form');
const recordsBody = document.getElementById('records-body');
const loadingElement = document.getElementById('loading');
const emptyState = document.getElementById('empty-state');
const formTitle = document.getElementById('form-title');
const submitBtn = document.getElementById('submit-btn');
const cancelBtn = document.getElementById('cancel-btn');

let isEditing = false;
let currentEditId = null;

// Initialize
document.addEventListener('DOMContentLoaded', function () {
    loadRecords();
    setDefaultDateTime();

    // Form submission
    saleForm.addEventListener('submit', handleFormSubmit);
    cancelBtn.addEventListener('click', cancelEdit);
});

// Set default datetime to now
function setDefaultDateTime() {
    const now = new Date();
    const localDateTime = now.toISOString().slice(0, 16);
    document.getElementById('date').value = localDateTime;
}

// Load all records
async function loadRecords() {
    try {
        showLoading();

        const response = await fetch(`${API_BASE}/items`);

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const records = await response.json();
        displayRecords(records);

    } catch (error) {
        console.error('Error loading records:', error);
        showError('Ошибка загрузки записей');
    }
}

// Display records in table
function displayRecords(records) {
    hideLoading();

    if (!records || records.length === 0) {
        recordsBody.innerHTML = '';
        emptyState.style.display = 'block';
        return;
    }

    emptyState.style.display = 'none';

    recordsBody.innerHTML = records.map(record => `
        <tr>
            <td>${formatDateTime(record.date)}</td>
            <td>
                <span class="type-badge type-${record.type}">
                    ${record.type === 'income' ? '💰 Доход' : '💸 Расход'}
                </span>
            </td>
            <td>${escapeHtml(record.category)}</td>
            <td class="amount amount-${record.type}">
                ${formatCurrency(record.amount)}
            </td>
            <td>${escapeHtml(record.description || '—')}</td>
            <td>
                <div class="action-buttons">
                    <button class="btn btn-edit" onclick="editRecord(${record.id})">
                        ✏️ Изменить
                    </button>
                    <button class="btn btn-danger" onclick="deleteRecord(${record.id})">
                        🗑️ Удалить
                    </button>
                </div>
            </td>
        </tr>
    `).join('');
}

// Handle form submission
async function handleFormSubmit(event) {
    event.preventDefault();

    const formData = getFormData();
    if (!formData) return;

    try {
        submitBtn.disabled = true;
        submitBtn.textContent = 'Сохранение...';

        if (isEditing) {
            await updateRecord(currentEditId, formData);
        } else {
            await createRecord(formData);
        }

        resetForm();
        await loadRecords();

    } catch (error) {
        console.error('Error saving record:', error);
        alert('Ошибка сохранения: ' + error.message);
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = isEditing ? 'Обновить' : 'Добавить';
    }
}

// Get form data with validation
function getFormData() {
    const amount = document.getElementById('amount').value;
    const type = document.getElementById('type').value;
    const category = document.getElementById('category').value.trim();
    const date = document.getElementById('date').value;
    const description = document.getElementById('description').value.trim();

    if (!amount || !type || !category || !date) {
        alert('Пожалуйста, заполните все обязательные поля');
        return null;
    }

    if (parseFloat(amount) <= 0) {
        alert('Сумма должна быть больше 0');
        return null;
    }

    // Convert local datetime to ISO string
    const dateISO = new Date(date).toISOString();

    return {
        amount: parseFloat(amount).toFixed(2),
        type: type,
        category: category,
        description: description,
        date: dateISO
    };
}

// Create new record
async function createRecord(data) {
    const response = await fetch(`${API_BASE}/items`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data)
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Ошибка создания записи');
    }

    return await response.json();
}

// Update record
async function updateRecord(id, data) {
    const response = await fetch(`${API_BASE}/items/${id}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data)
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Ошибка обновления записи');
    }

    return await response.json();
}

// Delete record
async function deleteRecord(id) {
    if (!confirm('Вы уверены, что хотите удалить эту запись?')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/items/${id}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            throw new Error('Ошибка удаления записи');
        }

        await loadRecords();

    } catch (error) {
        console.error('Error deleting record:', error);
        alert('Ошибка удаления записи');
    }
}

// Edit record
async function editRecord(id) {
    try {
        const response = await fetch(`${API_BASE}/items/${id}`);

        if (!response.ok) {
            throw new Error('Ошибка загрузки записи');
        }

        const record = await response.json();

        // Fill form with record data
        document.getElementById('edit-id').value = record.id;
        document.getElementById('amount').value = record.amount;
        document.getElementById('type').value = record.type;
        document.getElementById('category').value = record.category;
        document.getElementById('description').value = record.description || '';

        // Convert ISO date to local datetime
        const localDate = new Date(record.date);
        const localDateTime = localDate.toISOString().slice(0, 16);
        document.getElementById('date').value = localDateTime;

        // Switch to edit mode
        isEditing = true;
        currentEditId = record.id;
        formTitle.textContent = 'Редактировать запись';
        submitBtn.textContent = 'Обновить';
        cancelBtn.style.display = 'inline-block';

        // Scroll to form
        document.querySelector('.form-section').scrollIntoView({
            behavior: 'smooth'
        });

    } catch (error) {
        console.error('Error loading record for edit:', error);
        alert('Ошибка загрузки записи для редактирования');
    }
}

// Cancel edit
function cancelEdit() {
    resetForm();
}

// Reset form
function resetForm() {
    saleForm.reset();
    setDefaultDateTime();

    isEditing = false;
    currentEditId = null;
    formTitle.textContent = 'Добавить запись';
    submitBtn.textContent = 'Добавить';
    cancelBtn.style.display = 'none';
}

// Utility functions
function showLoading() {
    loadingElement.style.display = 'block';
    recordsBody.innerHTML = '';
    emptyState.style.display = 'none';
}

function hideLoading() {
    loadingElement.style.display = 'none';
}

function showError(message) {
    recordsBody.innerHTML = `
        <tr>
            <td colspan="6" style="text-align: center; color: #dc3545;">
                ${message}
            </td>
        </tr>
    `;
}

function formatDateTime(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('ru-RU');
}

function formatCurrency(amount) {
    return new Intl.NumberFormat('ru-RU', {
        style: 'currency',
        currency: 'RUB'
    }).format(amount);
}

function escapeHtml(unsafe) {
    return unsafe
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}