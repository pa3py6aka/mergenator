window.ProjectComponent = {
    name: 'Mergenator',
    template: `
        <div>
            <h2>🔀 Mergenator</h2>
            <p>Текущая структура проекта (пример):</p>
            <pre>
my-project/
 ├── .idea/
 ├── src/
 │    ├── Controller/
 │    │    └── IndexController.php
 │    ├── Entity/
 │    │    └── User.php
 │    └── Repository/
 ├── templates/
 │    └── base.html.twig
 ├── config/
 │    └── routes.yaml
 ├── public/
 │    └── index.php
 └── composer.json
            </pre>
            <p><span class="tag">// 12 файлов, 3 папки</span></p>
        </div>
    `
};