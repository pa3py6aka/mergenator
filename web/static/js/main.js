const logContainer = document.getElementById('output');
const messagesContainer = document.querySelector('.messages-container');
let wsClientID = null;

function addLog(text, type) {
    type = type ? type : 'default';
    logContainer.innerHTML += "<span class='time'>[" + (new Date).toLocaleTimeString() + "]</span> "
        + "<span class='msg-" + type + "'>" + text + "</span>" + "\n";

    logContainer.scrollTop = logContainer.scrollHeight;
}

// Создаём соединение с сервером
const socket = new WebSocket("wss://" + window.location.hostname + ":" + (window.location.port * 1 + 1) + "/ws");

// Обработчик открытия соединения
socket.onopen = function(event) {
    console.log("Соединение по WS установлено");
};

// Обработчик получения сообщения
socket.onmessage = function(event) {
    let msgData = JSON.parse(event.data);
    if (msgData.hasOwnProperty('clientID')) {
        wsClientID = msgData.clientID;
        addLog("Установлен ID для WS-соединения: " + wsClientID);
        return;
    }
    addLog(msgData.message, msgData.hasOwnProperty('type') ? msgData.type : null);
};

// Обработчик ошибок
socket.onerror = function(error) {
    console.log("Ошибка:", error);
};

// Отправка сообщения на сервер
function sendMessage() {
    const input = document.getElementById("messageInput");
    socket.send(input.value);
    input.value = "";
}


document.addEventListener('DOMContentLoaded', function() {
    // Получаем обе формы
    const forms = document.querySelectorAll('.merge-form');

    forms.forEach(form => {
        form.addEventListener('submit', function(event) {
            event.preventDefault(); // Блокируем стандартную отправку

            // Определяем ID формы, чтобы знать, куда выводить сообщения
            const formId = this.id;

            // Собираем данные формы
            const formData = new FormData(this);
            const action = event.submitter.value;// Значение нажатой кнопки
            formData.append('action', action);
            formData.append('ws_client_id', wsClientID)
            messagesContainer.style.display = 'block';

            // Отправляем запрос через fetch
            fetch('/merge', {
                method: 'POST',
                body: formData
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
                        addLog("MR успешно создан!\n\t   <a href=\""+data+"\" class=\"mr-link\" target=\"_blank\">"+data+"</a>");
                    } else {
                        // Обычный текстовый ответ
                        addLog(data);
                        // logContainer.textContent = data;
                        // messagesContainer.classList.add('success');
                    }
                })
                .catch(error => {
                    addLog(error.message || 'Произошла ошибка', 'error');
                });
        });
    });
});
