document.getElementById('mergeForm').addEventListener('submit', function(e) {
    e.preventDefault();

    const sourceBranch = document.getElementById('source_branch').value;
    const action = e.submitter.value; // Определяем, какая кнопка нажата
    const messagesDiv = document.getElementById('messages');

    // Очищаем предыдущие сообщения
    messagesDiv.style.display = 'none';
    messagesDiv.textContent = '';
    messagesDiv.className = 'messages';

    // Показываем статус "обработка"
    messagesDiv.textContent = 'Выполняется...';
    messagesDiv.classList.add('info');
    messagesDiv.style.display = 'block';

    fetch('/merge', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: `source_branch=${encodeURIComponent(sourceBranch)}&action=${encodeURIComponent(action)}`
    })
        .then(response => {
            if (response.ok) {
                return response.text();
            } else {
                return response.text().then(text => {
                    throw new Error(text);
                });
            }
        })
        .then(data => {
            // Если ответ содержит URL MR (для createMR)
            if (data.startsWith('http://') || data.startsWith('https://')) {
                messagesDiv.innerHTML = `
                        <strong>MR успешно создан!</strong><br>
                        <a href="${data}" class="mr-link" target="_blank">Перейти к MR</a>
                    `;
                messagesDiv.classList.add('success');
            } else {
                // Обычный текстовый ответ
                messagesDiv.textContent = data;
                messagesDiv.classList.add('success');
            }
        })
        .catch(error => {
            messagesDiv.textContent = error.message || 'Произошла ошибка';
            messagesDiv.classList.add('error');
        });
});