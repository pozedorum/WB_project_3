let currentUser = null;
let currentToken = null;

// Инициализация
document.addEventListener('DOMContentLoaded', function() {
    checkAuthStatus();
    loadEvents();
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

// Показать/скрыть формы авторизации
function showRegister() {
    document.getElementById('login-form').style.display = 'none';
    document.getElementById('register-form').style.display = 'block';
}

function showLogin() {
    document.getElementById('register-form').style.display = 'none';
    document.getElementById('login-form').style.display = 'block';
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

        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || 'Ошибка сервера');
        }

        return data;
    } catch (error) {
        alert(error.message);
        throw error;
    }
}

// Авторизация
async function login() {
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    try {
        const data = await apiCall('/login', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        });

        currentToken = data.token;
        currentUser = data.user;

        localStorage.setItem('jwtToken', currentToken);
        localStorage.setItem('user', JSON.stringify(currentUser));

        showUserInterface();
        loadEvents();
    } catch (error) {
        console.error('Login error:', error);
    }
}

async function register() {
    const name = document.getElementById('register-name').value;
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;
    const phone = document.getElementById('register-phone').value;

    try {
        const data = await apiCall('/register', {
            method: 'POST',
            body: JSON.stringify({ name, email, password, phone })
        });

        alert('Регистрация успешна! Теперь войдите в систему.');
        showLogin();
    } catch (error) {
        console.error('Register error:', error);
    }
}

function logout() {
    localStorage.removeItem('jwtToken');
    localStorage.removeItem('user');
    currentToken = null;
    currentUser = null;
    
    document.getElementById('user-info').style.display = 'none';
    document.getElementById('login-form').style.display = 'block';
    document.getElementById('admin-section').style.display = 'none';
    document.getElementById('user-section').style.display = 'block';
}

// Показать интерфейс пользователя
function showUserInterface() {
    document.getElementById('login-form').style.display = 'none';
    document.getElementById('register-form').style.display = 'none';
    document.getElementById('user-info').style.display = 'block';
    document.getElementById('user-name').textContent = currentUser.name;

    // Проверяем роль пользователя (простая логика - если создавал события, то организатор)
    loadAdminEvents();
}

// Работа с мероприятиями
async function loadEvents() {
    try {
        const data = await apiCall('/events');
        displayEvents(data.events);
    } catch (error) {
        console.error('Error loading events:', error);
    }
}

function displayEvents(events) {
    const eventsList = document.getElementById('events-list');
    eventsList.innerHTML = '';

    events.forEach(event => {
        const eventCard = document.createElement('div');
        eventCard.className = 'event-card';
        
        const canBook = currentUser && currentUser.id !== event.created_by;
        const isMyEvent = currentUser && currentUser.id === event.created_by;
        
        eventCard.innerHTML = `
            <h4>${event.name}</h4>
            <p>Дата: ${formatDate(event.date)}</p>
            <p>Стоимость: ${event.cost} руб.</p>
            <p>Свободных мест: ${event.availableSeats || 0}</p>
            ${isMyEvent ? `<p><small>Организатор: Вы</small></p>` : ''}
            <button onclick="showEventDetails(${event.id})">Подробнее</button>
            ${canBook ? `<button onclick="confirmBookingPrompt(${event.id})">Подтвердить бронь</button>` : ''}
        `;
        eventsList.appendChild(eventCard);
    });
}

async function showEventDetails(eventId) {
    try {
        const event = await apiCall(`/events/${eventId}`);
        const eventInfo = document.getElementById('event-info');
        
        // Правильно рассчитываем время брони в минутах
        const bookingTimeMinutes = Math.round(event.life_span / 60);
        
        eventInfo.innerHTML = `
            <h4>${event.name}</h4>
            <p>Дата: ${formatDate(event.date)}</p>
            <p>Стоимость: ${event.cost} руб.</p>
            <p>Всего мест: ${event.total_seats}</p>
            <p>Свободных мест: ${event.available_seats}</p>
            <p>Время брони: ${bookingTimeMinutes} минут</p>
        `;

        document.getElementById('event-details').style.display = 'block';
        
        // Проверяем, может ли пользователь бронировать это мероприятие
        const canBook = currentUser && currentUser.id !== event.created_by;
        document.getElementById('booking-form').style.display = canBook ? 'block' : 'none';
        
        // Сохраняем eventId для использования в bookEvent()
        eventInfo.dataset.eventId = eventId;
    } catch (error) {
        console.error('Error loading event details:', error);
    }
}

async function bookEvent() {
    const eventId = document.getElementById('event-info').dataset.eventId;
    const seatCount = document.getElementById('seat-count').value;

    try {
        const data = await apiCall(`/events/${eventId}/book`, {
            method: 'POST',
            body: JSON.stringify({ seat_count: parseInt(seatCount) })
        });

        alert(`Бронирование создано! Код брони: ${data.booking.booking_code}`);
        loadEvents();
        document.getElementById('booking-form').style.display = 'none';
        document.getElementById('seat-count').value = '';
    } catch (error) {
        console.error('Error booking event:', error);
    }
}

// Функции для организатора
async function loadAdminEvents() {
    try {
        const events = await apiCall('/events');
        const adminEvents = events.filter(event => event.created_by === currentUser.id);
        
        if (adminEvents.length > 0) {
            document.getElementById('admin-section').style.display = 'block';
            displayAdminEvents(adminEvents);
        }
    } catch (error) {
        console.error('Error loading admin events:', error);
    }
}

function displayAdminEvents(events) {
    const eventsList = document.getElementById('admin-events-list');
    eventsList.innerHTML = '';

    events.forEach(event => {
        const eventCard = document.createElement('div');
        eventCard.className = 'event-card';
        eventCard.innerHTML = `
            <h4>${event.name}</h4>
            <p>Дата: ${formatDate(event.date)}</p>
            <p>Свободных мест: ${event.availableSeats}</p>
            <button onclick="showEventBookings(${event.id})">Показать брони</button>
            <button class="danger" onclick="deleteEvent(${event.id})">Удалить</button>
        `;
        eventsList.appendChild(eventCard);
    });
}  

async function createEvent() {
    const name = document.getElementById('event-name').value;
    const dateInput = document.getElementById('event-date').value;
    const cost = document.getElementById('event-cost').value || 0;
    const seats = document.getElementById('event-seats').value;
    const lifespan = document.getElementById('event-lifespan').value;

    // Валидация
    if (!name || !dateInput || !seats || !lifespan) {
        showNotification('Заполните все обязательные поля', true);
        return;
    }

    const date = new Date(dateInput);
    if (isNaN(date.getTime()) || date <= new Date()) {
        showNotification('Дата должна быть в будущем', true);
        return;
    }

    if (parseInt(seats) <= 0) {
        showNotification('Количество мест должно быть положительным числом', true);
        return;
    }

    try {
        await apiCall('/events', {
            method: 'POST',
            body: JSON.stringify({
                name,
                date: date.toISOString(),
                cost: parseInt(cost),
                total_seats: parseInt(seats),
                life_span: lifespan
            })
        });

        showNotification('Мероприятие создано успешно!');
        
        // Очищаем форму
        document.getElementById('event-name').value = '';
        document.getElementById('event-date').value = '';
        document.getElementById('event-cost').value = '';
        document.getElementById('event-seats').value = '';
        document.getElementById('event-lifespan').value = '15m';
        
        loadEvents();
        loadAdminEvents();
    } catch (error) {
        console.error('Error creating event:', error);
}
}
// Функция для подтверждения бронирования
async function confirmBookingPrompt(eventId) {
    const bookingCode = prompt('Введите код бронирования:');
    if (bookingCode) {
        try {
            const data = await apiCall(`/events/${eventId}/confirm`, {
                method: 'POST',
                body: JSON.stringify({ booking_code: bookingCode })
            });
            
            alert('Бронирование подтверждено!');
            loadEvents(); // Обновляем список
        } catch (error) {
            console.error('Error confirming booking:', error);
        }
    }
}

// Функция для показа бронирований мероприятия (для организатора)
async function showEventBookings(eventId) {
    try {
        // Здесь нужно реализовать endpoint для получения бронирований по event_id
        // Пока просто заглушка
        alert('Функция показа бронирований будет реализована в следующей версии');
    } catch (error) {
        console.error('Error loading event bookings:', error);
    }
}

// Функция для удаления мероприятия
async function deleteEvent(eventId) {
    if (confirm('Вы уверены, что хотите удалить это мероприятие?')) {
        try {
            // Здесь нужно реализовать endpoint для удаления мероприятий
            // Пока просто заглушка
            alert('Функция удаления мероприятий будет реализована в следующей версии');
        } catch (error) {
            console.error('Error deleting event:', error);
        }
    }
}


// Добавляем кнопку для показа бронирований в интерфейс пользователя
function showUserInterface() {
    document.getElementById('login-form').style.display = 'none';
    document.getElementById('register-form').style.display = 'none';
    document.getElementById('user-info').style.display = 'block';
    document.getElementById('user-name').textContent = currentUser.name;


    loadAdminEvents();
}

// Функции для показа уведомлений
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

// Обновляем API вызовы для использования уведомлений
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

        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || 'Ошибка сервера');
        }

        return data;
    } catch (error) {
        showNotification(error.message, true);
        throw error;
    }
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


// Добавляем функцию для обновления интерфейса при загрузке
function updateUI() {
    loadEvents();
    if (currentUser) {
        showUserInterface();
        loadAdminEvents();
    }
}

// Обновляем функцию login для вызова updateUI
async function login() {
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    try {
        const data = await apiCall('/login', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        });

        currentToken = data.token;
        currentUser = data.user;

        localStorage.setItem('jwtToken', currentToken);
        localStorage.setItem('user', JSON.stringify(currentUser));

        showNotification('Вход выполнен успешно!');
        updateUI();
    } catch (error) {
        console.error('Login error:', error);
    }
}

// Обновляем функцию logout
function logout() {
    localStorage.removeItem('jwtToken');
    localStorage.removeItem('user');
    currentToken = null;
    currentUser = null;
    
    document.getElementById('user-info').style.display = 'none';
    document.getElementById('login-form').style.display = 'block';
    document.getElementById('admin-section').style.display = 'none';
    document.getElementById('user-section').style.display = 'block';
    
    // Очищаем списки
    document.getElementById('events-list').innerHTML = '';
    document.getElementById('event-details').style.display = 'none';
    
    showNotification('Вы вышли из системы');
}

// Добавляем обработчик для Enter в формах
document.addEventListener('DOMContentLoaded', function() {
    // Обработчик Enter для формы логина
    document.getElementById('login-password').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            login();
        }
    });
    
    // Обработчик Enter для формы регистрации
    document.getElementById('register-password').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            register();
        }
    });
    
    checkAuthStatus();
    loadEvents();
})