const logContainer = document.getElementById('output');
const messagesContainer = document.querySelector('.messages-container');
let wsClientID = null;
let socket = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 50;
const INITIAL_RECONNECT_DELAY = 1000; // 1 секунда
let reconnectTimeout = null;

function connectWebSocket() {
    if (socket && socket.readyState !== WebSocket.CLOSED) {
        socket.close();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.hostname}:${window.location.port}/ws`;

    socket = new WebSocket(wsUrl);

    socket.onopen = function(event) {
        console.log("Соединение по WS установлено");
        addLog("WebSocket-соединение успешно установлено.", "success");
        reconnectAttempts = 0;
    };

    socket.onmessage = function(event) {
        try {
            let msgData = JSON.parse(event.data);

            if (msgData.type === 'pong') {
                console.log('Получен pong от сервера');
                return;
            }

            // Обработка ID клиента
            if (msgData.hasOwnProperty('clientID')) {
                wsClientID = msgData.clientID;
                addLog("Установлен ID для WS-соединения: " + wsClientID, "info");
                return;
            }

            // Пропускаем пустые сообщения
            if (!msgData.message && !msgData.type) {
                console.log('Получено служебное сообщение:', msgData);
                return;
            }

            // Обычное сообщение выводим в лог
            addLog(msgData.message || JSON.stringify(msgData),
                msgData.hasOwnProperty('type') ? msgData.type : "info");

        } catch (e) {
            console.error('Ошибка парсинга сообщения:', e);
            console.error('Сырые данные:', event.data);
        }
    };

    socket.onerror = function(error) {
        console.error("Критическая ошибка WebSocket (onerror):", error);
        if (reconnectAttempts === 0) {
            addLog("Потеря соединения с сервером. Попытка переподключения...", "warning");
        }
    };

    socket.onclose = function(event) {
        let reason = getCloseReason(event);

        if (event.wasClean) {
            addLog(`WebSocket отключен. Причина: ${reason}`, "info");
        } else {
            console.log(`WebSocket отключен: ${reason}`);
        }

        scheduleReconnect();
    };
}

function getCloseReason(event) {
    switch (event.code) {
        case 1000: return "Соединение закрыто нормально.";
        case 1001: return "Сервер ушёл вниз.";
        case 1002: return "Протокол нарушен, соединение разорвано.";
        case 1003: return "Сервер получил данные, которые не может обработать.";
        case 1005: return "Не получен ожидаемый код закрытия.";
        case 1006: return "Аномальное закрытие соединения (нет данных о причине).";
        case 1007: return "Данные сообщения некорректны (например, невалидный UTF-8).";
        case 1008: return "Сообщение нарушает политику сервера.";
        case 1009: return "Сообщение слишком большое для обработки.";
        case 1011: return "На сервере произошла непредвиденная ошибка.";
        case 1012: return "Сервер перезапускается.";
        case 1013: return "Слишком долго ждать ответа от сервера.";
        case 1014: return "Невозможно установить соединение через выбранный подпротокол.";
        default:
            if (event.code >= 3000 && event.code <= 4999) {
                return `Код ${event.code}: зарезервирован для библиотек/фреймворков.`;
            } else if (event.code >= 1000 && event.code <= 2999) {
                return `Код ${event.code}: стандартный код WebSocket.`;
            } else {
                return `Неизвестный код ошибки: ${event.code}.`;
            }
    }
}

function scheduleReconnect() {
    if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
    }

    if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
        addLog("Достигнут лимит попыток переподключения. Обновите страницу вручную.", "error");
        return;
    }

    const delay = Math.min(INITIAL_RECONNECT_DELAY * Math.pow(2, reconnectAttempts), 30000);

    reconnectAttempts++;

    console.log(`Попытка переподключения ${reconnectAttempts} через ${delay}ms...`);

    if (reconnectAttempts % 5 === 0 || delay >= 30000) {
        addLog(`Переподключение... Попытка ${reconnectAttempts}`, "warning");
    }

    reconnectTimeout = setTimeout(() => {
        connectWebSocket();
    }, delay);
}

function sendMessage(message) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(message);
    } else {
        addLog("Соединение с сервером потеряно. Попытка переподключения...", "warning");
        if (reconnectAttempts === 0) {
            scheduleReconnect();
        }
    }
}

function startHeartbeat() {
    setInterval(() => {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ type: 'ping', timestamp: Date.now() }));
        }
    }, 30000);
}

function addLog(text, type) {
    if (!text || text === 'undefined') {
        return;
    }

    type = type ? type : 'default';
    const time = new Date().toLocaleTimeString();
    const logEntry = `<span class='time'>[${time}]</span> <span class='msg-${type}'>${text}</span>\n`;

    logContainer.insertAdjacentHTML('beforeend', logEntry);
    logContainer.scrollTop = logContainer.scrollHeight;
}

function submitForm(formData) {
    messagesContainer.style.display = 'block';

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
            if (data.startsWith('http://') || data.startsWith('https://')) {
                addLog("MR успешно создан!\n\t   <a href=\"" + data + "\" class=\"mr-link\" target=\"_blank\">" + data + "</a>");
            } else {
                addLog(data);
            }
        })
        .catch(error => {
            addLog(error.message || 'Произошла ошибка', 'error');
        });
}

document.addEventListener('DOMContentLoaded', function() {
    connectWebSocket();
    startHeartbeat();

    const forms = document.querySelectorAll('.merge-form');
    forms.forEach(form => {
        form.addEventListener('submit', function(event) {
            event.preventDefault();

            const formData = new FormData(this);
            const action = event.submitter.value;
            formData.append('action', action);

            if (wsClientID) {
                formData.append('ws_client_id', wsClientID);
                submitForm(formData);
            } else {
                addLog("Ожидание WebSocket соединения...", "warning");
                let waitAttempts = 0;
                const waitForConnection = setInterval(() => {
                    if (wsClientID) {
                        clearInterval(waitForConnection);
                        formData.append('ws_client_id', wsClientID);
                        submitForm(formData);
                    }
                    waitAttempts++;
                    if (waitAttempts > 30) {
                        clearInterval(waitForConnection);
                        addLog("Не удалось установить WebSocket соединение. Обновите страницу.", "error");
                    }
                }, 100);
            }
        });
    });
});

window.addEventListener('online', function() {
    addLog("Сетевое соединение восстановлено. Переподключаемся...", "success");
    reconnectAttempts = 0;
    connectWebSocket();
});

window.addEventListener('offline', function() {
    addLog("Сетевое соединение потеряно. Ожидание восстановления...", "warning");
});

window.addEventListener('beforeunload', function() {
    if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
    }
    if (socket) {
        socket.close();
    }
});