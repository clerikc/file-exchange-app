document.addEventListener('DOMContentLoaded', function() {
    const uploadForm = document.querySelector('form[enctype="multipart/form-data"]');
    
    if (uploadForm) {
        uploadForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const fileInput = this.querySelector('input[type="file"]');
            const button = this.querySelector('button[type="submit"]');
            const originalButtonText = button.textContent;
            
            if (!fileInput.files.length) {
                alert('Please select a file to upload');
                return;
            }
            
            const formData = new FormData();
            formData.append('file', fileInput.files[0]);
            
            // Показываем индикатор загрузки
            button.disabled = true;
            button.textContent = 'Uploading...';
            
            // Создаем прогресс-бар если его нет
            let progressBar = this.querySelector('.upload-progress');
            if (!progressBar) {
                progressBar = document.createElement('div');
                progressBar.className = 'upload-progress';
                progressBar.innerHTML = `
                    <div class="progress-container">
                        <div class="progress-bar"></div>
                        <div class="progress-text">0%</div>
                    </div>
                `;
                this.appendChild(progressBar);
            }
            
            const xhr = new XMLHttpRequest();
            
            // Отслеживаем прогресс загрузки
            xhr.upload.addEventListener('progress', function(e) {
                if (e.lengthComputable) {
                    const percentComplete = (e.loaded / e.total) * 100;
                    const progressBarElement = progressBar.querySelector('.progress-bar');
                    const progressText = progressBar.querySelector('.progress-text');
                    
                    progressBarElement.style.width = percentComplete + '%';
                    progressText.textContent = Math.round(percentComplete) + '%';
                }
            });
            
            xhr.addEventListener('load', function() {
                if (xhr.status === 200) {
                    // Успешная загрузка
                    progressBar.querySelector('.progress-text').textContent = 'Upload complete!';
                    setTimeout(() => {
                        // Перезагружаем страницу чтобы показать новый файл
                        window.location.reload();
                    }, 1000);
                } else {
                    // Ошибка загрузки
                    alert('Upload failed: ' + xhr.statusText);
                    button.disabled = false;
                    button.textContent = originalButtonText;
                    progressBar.style.display = 'none';
                }
            });
            
            xhr.addEventListener('error', function() {
                alert('Upload failed. Please try again.');
                button.disabled = false;
                button.textContent = originalButtonText;
                progressBar.style.display = 'none';
            });
            
            xhr.open('POST', this.action);
            xhr.send(formData);
        });
    }
    
    // Добавляем валидацию размера файла
    const fileInputs = document.querySelectorAll('input[type="file"]');
    fileInputs.forEach(input => {
        input.addEventListener('change', function() {
            const maxSize = 100 * 1024 * 1024; // 100MB в байтах
            if (this.files[0] && this.files[0].size > maxSize) {
                alert('File size exceeds the maximum limit of 100MB');
                this.value = '';
            }
        });
    });
});