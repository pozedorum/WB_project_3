const API_BASE = 'http://localhost:8080/api';

// DOM Elements
const analyticsForm = document.getElementById('analytics-form');
const exportCsvBtn = document.getElementById('export-csv');
const analyticsLoading = document.getElementById('analytics-loading');
const analyticsResults = document.getElementById('analytics-results');
const noDataElement = document.getElementById('no-data');
const generalStats = document.getElementById('general-stats');
const groupedData = document.getElementById('grouped-data');
const groupedBody = document.getElementById('grouped-body');

// Initialize
document.addEventListener('DOMContentLoaded', function () {
    setDefaultDateRange();

    analyticsForm.addEventListener('submit', handleAnalyticsSubmit);
    exportCsvBtn.addEventListener('click', handleExportCSV);
});

// Set default date range (last 30 days)
function setDefaultDateRange() {
    const now = new Date();
    const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);

    document.getElementById('to-date').value = now.toISOString().slice(0, 16);
    document.getElementById('from-date').value = thirtyDaysAgo.toISOString().slice(0, 16);
}

// Handle analytics form submission
async function handleAnalyticsSubmit(event) {
    event.preventDefault();

    const filters = getAnalyticsFilters();
    if (!filters) return;

    await loadAnalytics(filters);
}

// Get analytics filters
function getAnalyticsFilters() {
    const fromDate = document.getElementById('from-date').value;
    const toDate = document.getElementById('to-date').value;

    if (!fromDate || !toDate) {
        alert('Пожалуйста, укажите период');
        return null;
    }

    const fromISO = new Date(fromDate).toISOString();
    const toISO = new Date(toDate).toISOString();

    if (fromISO >= toISO) {
        alert('Дата "С" должна быть раньше даты "По"');
        return null;
    }

    const filters = {
        from: fromISO,
        to: toISO
    };

    const type = document.getElementById('analytics-type').value;
    const category = document.getElementById('category-filter').value.trim();
    const groupBy = document.getElementById('group-by').value;

    if (type) filters.type = type;
    if (category) filters.category = category;
    if (groupBy) filters.group_by = groupBy;

    return filters;
}

// Load analytics data
async function loadAnalytics(filters) {
    try {
        showAnalyticsLoading();

        const queryString = new URLSearchParams(filters).toString();
        const response = await fetch(`${API_BASE}/analytics?${queryString}`);

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const analyticsData = await response.json();
        displayAnalytics(analyticsData, filters.group_by);

    } catch (error) {
        console.error('Error loading analytics:', error);
        showAnalyticsError('Ошибка загрузки аналитики');
    }
}

// Display analytics data
function displayAnalytics(data, groupBy) {
    hideAnalyticsLoading();

    if (!data || (data.count === 0 && (!data.grouped_data || data.grouped_data.length === 0))) {
        showNoData();
        return;
    }

    analyticsResults.style.display = 'block';
    noDataElement.style.display = 'none';

    displayGeneralStats(data);

    if (data.grouped_data && data.grouped_data.length > 0) {
        displayGroupedData(data.grouped_data, groupBy);
    } else {
        groupedData.style.display = 'none';
    }
}

// Display general statistics
function displayGeneralStats(data) {
    generalStats.innerHTML = `
        <div class="stat-card">
            <h3>📊 Общая сумма</h3>
            <div class="value">${formatCurrency(data.total || 0)}</div>
        </div>
        <div class="stat-card">
            <h3>📈 Средняя сумма</h3>
            <div class="value">${formatCurrency(data.average || 0)}</div>
        </div>
        <div class="stat-card">
            <h3>🔢 Количество операций</h3>
            <div class="value">${data.count || 0}</div>
        </div>
        <div class="stat-card">
            <h3>📐 Медиана</h3>
            <div class="value">${formatCurrency(data.median || 0)}</div>
        </div>
        <div class="stat-card">
            <h3>🎯 90-й перцентиль</h3>
            <div class="value">${formatCurrency(data.percentile_90 || 0)}</div>
        </div>
    `;
}

// Display grouped data
function displayGroupedData(groupedDataArray, groupBy) {
    groupedData.style.display = 'block';

    const groupName = getGroupDisplayName(groupBy);

    groupedBody.innerHTML = groupedDataArray.map(item => `
        <tr>
            <td><strong>${escapeHtml(item.group)}</strong></td>
            <td class="amount">${formatCurrency(item.total)}</td>
            <td class="amount">${formatCurrency(item.average)}</td>
            <td>${item.count}</td>
            <td class="amount">${formatCurrency(item.median)}</td>
            <td class="amount">${formatCurrency(item.percentile_90)}</td>
        </tr>
    `).join('');
}

// Get display name for group
function getGroupDisplayName(groupBy) {
    const names = {
        'day': 'День',
        'week': 'Неделя',
        'month': 'Месяц',
        'category': 'Категория'
    };
    return names[groupBy] || 'Группа';
}

// Handle CSV export
async function handleExportCSV() {
    const filters = getAnalyticsFilters();
    if (!filters) return;

    try {
        exportCsvBtn.disabled = true;
        exportCsvBtn.textContent = 'Экспорт...';

        const queryString = new URLSearchParams(filters).toString();
        const response = await fetch(`${API_BASE}/csv?${queryString}`);

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const csvData = await response.text();
        downloadCSV(csvData, `analytics_export_${new Date().toISOString().slice(0, 10)}.csv`);

    } catch (error) {
        console.error('Error exporting CSV:', error);
        alert('Ошибка экспорта CSV: ' + error.message);
    } finally {
        exportCsvBtn.disabled = false;
        exportCsvBtn.textContent = 'Экспорт в CSV';
    }
}

// Download CSV file
function downloadCSV(csvData, filename) {
    const blob = new Blob([csvData], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);

    link.setAttribute('href', url);
    link.setAttribute('download', filename);
    link.style.visibility = 'hidden';

    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
}

// Utility functions
function showAnalyticsLoading() {
    analyticsLoading.style.display = 'block';
    analyticsResults.style.display = 'none';
    noDataElement.style.display = 'none';
}

function hideAnalyticsLoading() {
    analyticsLoading.style.display = 'none';
}

function showNoData() {
    analyticsResults.style.display = 'none';
    noDataElement.style.display = 'block';
}

function showAnalyticsError(message) {
    analyticsResults.style.display = 'none';
    noDataElement.innerHTML = `
        <div style="color: #dc3545; text-align: center;">
            ${message}
        </div>
    `;
    noDataElement.style.display = 'block';
}

function formatCurrency(amount) {
    return new Intl.NumberFormat('ru-RU', {
        style: 'currency',
        currency: 'RUB',
        minimumFractionDigits: 2
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