class ImageProcessor {
    constructor() {
        this.baseUrl = window.location.origin;
        this.images = [];
        this.currentImageId = null;
        
        this.initializeEventListeners();
        this.loadImages();
    }

    initializeEventListeners() {
        // Форма загрузки
        document.getElementById('uploadForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.uploadImage();
        });

        // Модальное окно
        document.querySelector('.close').addEventListener('click', () => {
            this.closeModal();
        });

        document.getElementById('downloadBtn').addEventListener('click', () => {
            this.downloadImage();
        });

        document.getElementById('deleteModalBtn').addEventListener('click', () => {
            this.deleteImage(this.currentImageId);
        });

        // Закрытие модального окна по клику вне контента
        window.addEventListener('click', (e) => {
            const modal = document.getElementById('imageModal');
            if (e.target === modal) {
                this.closeModal();
            }
        });
    }

    async uploadImage() {
        const formData = new FormData(document.getElementById('uploadForm'));
        const uploadBtn = document.getElementById('uploadBtn');
        
        try {
            uploadBtn.disabled = true;
            uploadBtn.textContent = 'Загрузка...';

            const response = await fetch(`${this.baseUrl}/upload`, {
                method: 'POST',
                body: formData
            });

            if (response.ok) {
                const result = await response.json();
                this.showAlert('Изображение успешно загружено и поставлено в очередь на обработку!', 'success');
                this.addImageToList(result);
                document.getElementById('uploadForm').reset();
            } else {
                const error = await response.json();
                throw new Error(error.details || error.error || 'Ошибка загрузки');
            }
        } catch (error) {
            this.showAlert(`Ошибка: ${error.message}`, 'error');
            console.error('Upload error:', error);
        } finally {
            uploadBtn.disabled = false;
            uploadBtn.textContent = 'Загрузить и обработать';
        }
    }

    async loadImages() {
        // В реальном приложении здесь бы был endpoint для получения списка изображений
        // Поскольку его нет, мы будем хранить images в localStorage
        const savedImages = localStorage.getItem('processedImages');
        if (savedImages) {
            this.images = JSON.parse(savedImages);
            this.renderImages();
        }
    }

    addImageToList(imageData) {
        const image = {
            id: imageData.image_id,
            originalName: 'uploaded_image',
            status: 'processing',
            uploadedAt: new Date().toISOString(),
            metadata: imageData
        };

        this.images.unshift(image);
        this.saveImages();
        this.renderImages();

        // Запускаем опрос статуса для нового изображения
        this.pollImageStatus(image.id);
    }

    async pollImageStatus(imageId) {
        const maxAttempts = 60; // 5 минут с интервалом 5 секунд
        let attempts = 0;

        const checkStatus = async () => {
            if (attempts >= maxAttempts) {
                this.updateImageStatus(imageId, 'failed');
                return;
            }

            try {
                const response = await fetch(`${this.baseUrl}/image/${imageId}`);
                
                if (response.status === 404) {
                    this.updateImageStatus(imageId, 'failed');
                    return;
                }

                if (response.headers.get('content-type')?.includes('application/json')) {
                    const data = await response.json();
                    if (data.status === 'completed') {
                        this.updateImageStatus(imageId, 'completed', data.metadata);
                    } else if (data.status === 'processing') {
                        attempts++;
                        setTimeout(checkStatus, 5000); // Проверяем каждые 5 секунд
                    }
                } else {
                    // Получили бинарные данные - изображение готово
                    this.updateImageStatus(imageId, 'completed');
                }
            } catch (error) {
                console.error('Status poll error:', error);
                attempts++;
                setTimeout(checkStatus, 5000);
            }
        };

        checkStatus();
    }

    updateImageStatus(imageId, status, metadata = null) {
        const imageIndex = this.images.findIndex(img => img.id === imageId);
        if (imageIndex !== -1) {
            this.images[imageIndex].status = status;
            if (metadata) {
                this.images[imageIndex].metadata = metadata;
            }
            this.saveImages();
            this.renderImages();
        }
    }

    async viewImage(imageId) {
        this.currentImageId = imageId;
        
        try {
            const response = await fetch(`${this.baseUrl}/image/${imageId}`);
            
            if (response.ok) {
                if (response.headers.get('content-type')?.includes('application/json')) {
                    const data = await response.json();
                    if (data.status === 'processing') {
                        this.showAlert('Изображение еще обрабатывается...', 'info');
                        return;
                    }
                } else {
                    // Это бинарные данные изображения
                    const blob = await response.blob();
                    const url = URL.createObjectURL(blob);
                    
                    document.getElementById('modalImage').src = url;
                    document.getElementById('imageModal').style.display = 'block';
                }
            }
        } catch (error) {
            this.showAlert('Ошибка при загрузке изображения', 'error');
            console.error('View image error:', error);
        }
    }

    async deleteImage(imageId) {
        if (!confirm('Удалить это изображение?')) return;

        try {
            const response = await fetch(`${this.baseUrl}/image/${imageId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                this.images = this.images.filter(img => img.id !== imageId);
                this.saveImages();
                this.renderImages();
                this.closeModal();
                this.showAlert('Изображение удалено', 'success');
            } else {
                throw new Error('Ошибка удаления');
            }
        } catch (error) {
            this.showAlert('Ошибка при удалении изображения', 'error');
            console.error('Delete error:', error);
        }
    }

    downloadImage() {
        const imageUrl = document.getElementById('modalImage').src;
        const link = document.createElement('a');
        link.href = imageUrl;
        link.download = `processed-${this.currentImageId}`;
        link.click();
    }

    closeModal() {
        document.getElementById('imageModal').style.display = 'none';
        this.currentImageId = null;
    }

    renderImages() {
        const container = document.getElementById('imagesList');
        const emptyState = document.getElementById('emptyState');

        if (this.images.length === 0) {
            emptyState.style.display = 'block';
            container.innerHTML = '';
            return;
        }

        emptyState.style.display = 'none';
        container.innerHTML = this.images.map(image => `
            <div class="image-card">
                <div class="image-preview">
                    ${image.status === 'completed' ? 
                        `<img src="${this.baseUrl}/image/${image.id}" alt="Processed image" onerror="this.style.display='none'">` : 
                        `<div class="placeholder">🖼️</div>`
                    }
                </div>
                <div class="image-info">
                    <h3 title="${image.id}">${image.id.substring(0, 20)}...</h3>
                    <span class="status status-${image.status}">
                        ${this.getStatusText(image.status)}
                    </span>
                    <div class="image-actions">
                        <button class="btn-view" onclick="app.viewImage('${image.id}')" 
                                ${image.status !== 'completed' ? 'disabled' : ''}>
                            Просмотр
                        </button>
                        <button class="btn-delete" onclick="app.deleteImage('${image.id}')">
                            Удалить
                        </button>
                    </div>
                </div>
            </div>
        `).join('');
    }

    getStatusText(status) {
        const statusMap = {
            'processing': 'В обработке',
            'completed': 'Готово',
            'failed': 'Ошибка'
        };
        return statusMap[status] || status;
    }

    saveImages() {
        localStorage.setItem('processedImages', JSON.stringify(this.images));
    }

    showAlert(message, type) {
        // Создаем временное уведомление
        const alert = document.createElement('div');
        alert.className = `alert alert-${type}`;
        alert.textContent = message;
        
        document.body.appendChild(alert);
        
        setTimeout(() => {
            alert.remove();
        }, 3000);
    }
}

// Инициализация приложения
const app = new ImageProcessor();