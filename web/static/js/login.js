document.getElementById('loginForm').addEventListener('submit', function(event) {
    event.preventDefault();

    let value = document.getElementById('userIdInput').value;
    if (/^\d+$/.test(value)) {
        document.cookie = "gitlab_user_id=" + value + "; expires=Fri, 31 Dec 9999 23:59:59 GMT; path=/";
        document.location = "/mergenator"
        return;
    }

    alert("Неверный формат userId - должно быть число.");
})