package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	BackendStandBranch  = "andagar/r2533-bh-a1"
	FrontendStandBranch = "andagar/pre-release"
	CIMainBranch        = "r2533-andagar/ci"
)

func handleMerge(c *gin.Context) {
	sourceBranch := c.PostForm("source_branch")
	action := c.PostForm("action")
	repo := c.PostForm("repo")
	wsClientID := c.PostForm("ws_client_id")

	gitlabUserId, err := c.Cookie("gitlab_user_id")
	if err != nil {
		c.HTML(200, "login.html", gin.H{})
		return
	}
	gitlabUserIdInt, err := strconv.Atoi(gitlabUserId)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}

	repository := Repository{AssigneeId: gitlabUserIdInt}

	if repo == "backend" {
		repository.StandBranch = BackendStandBranch
		repository.ProjectId = "140"
	} else {
		repository.StandBranch = FrontendStandBranch
		repository.ProjectId = "141"
	}

	switch action {
	case "createMR":
		err := handleCreateMR(sourceBranch, c, repository, wsClientID)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		} else {
			// В handleCreateMR мы уже отправляем полный ответ (включая URL)
		}
	case "merge":
		err := handleMergeAction(sourceBranch, c, repository)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		} else {
			c.String(http.StatusOK, fmt.Sprintf("Ветка %s готова к мерджу в %s.", sourceBranch, repository.StandBranch))
		}
	default:
		c.String(http.StatusBadRequest, "Неизвестное действие")
	}
}

// Логика для кнопки "Создать MR"
func handleCreateMR(branch string, c *gin.Context, repo Repository, wsClientID string) error {
	sendMessageByID(wsClientID, fmt.Sprintf(
		"Запрос на создание MR ветки `%s` [%s]",
		branch, getProjectName(repo.ProjectId)), WSMessageTypeHeader)

	// 1. Проверка префикса исходной ветки
	sendMessageByID(wsClientID, "Проверка префикса исходной ветки", WSMessageTypeDefault)
	if err := validateBranchPrefix(branch); err != nil {
		return err
	}

	// 2. Проверка существования исходной ветки
	sendMessageByID(wsClientID, "Проверка существования исходной ветки", WSMessageTypeDefault)
	exists, err := branchExistsInRepo(branch, repo.ProjectId)
	if err != nil {
		return fmt.Errorf("ошибка проверки ветки: %v", err)
	}
	if !exists {
		return fmt.Errorf("ветка %s не найдена", branch)
	}

	// 3. Формирование имени CI‑ветки
	ciBranch := makeCIBranchName(branch)

	// 4. Проверка существования CI‑ветки
	sendMessageByID(wsClientID, fmt.Sprintf("Проверяем на существование CI-ветки `%s`", ciBranch), WSMessageTypeDefault)
	ciExists, err := branchExistsInRepo(ciBranch, repo.ProjectId)
	if err != nil {
		return fmt.Errorf("Ошибка проверки CI‑ветки: %v", err)
	}

	// 5. Проверка открытого MR для CI‑ветки
	sendMessageByID(wsClientID, "Проверяем есть ли уже открытый MR", WSMessageTypeDefault)
	hasMR, mrID, mrUrl, err := hasOpenMR(ciBranch, repo.StandBranch, repo)
	if err != nil {
		return fmt.Errorf("Ошибка проверки MR для CI‑ветки: %v", err)
	}
	if hasMR {
		return fmt.Errorf(
			"Для CI‑ветки %s уже есть открытый MR в %s:\n\t   %s", ciBranch, repo.StandBranch, getLink(Link{Href: mrUrl}))
	}

	// 6. Если CI‑ветка существует — удаляем её
	if ciExists {
		sendMessageByID(wsClientID, fmt.Sprintf("Удаляем старую CI-ветку `%s`", ciBranch), WSMessageTypeDefault)
		if err := deleteRemoteBranch(ciBranch, repo); err != nil {
			return fmt.Errorf("не удалось удалить CI‑ветку %s: %v", ciBranch, err)
		}
	}

	// 7. Создание CI‑ветки от исходной
	sendMessageByID(wsClientID, fmt.Sprintf("Создаём новую CI-ветку `%s`", ciBranch), WSMessageTypeDefault)
	if err := createRemoteBranch(branch, ciBranch, repo); err != nil {
		return fmt.Errorf("не удалось создать CI‑ветку %s: %v", ciBranch, err)
	}

	// 8. Проверка: есть ли уже открытый MR для этих веток?
	hasMR, mrID, mrUrl, err = hasOpenMR(CIMainBranch, ciBranch, repo)
	if err != nil {
		return fmt.Errorf("ошибка проверки существующих MR: %v", err)
	}
	if hasMR {
		// MR уже существует — не создаём новый, а используем существующий
		log.Printf("Уже есть открытый MR №%d для %s → %s", mrID, CIMainBranch, ciBranch)
	} else {
		// Создаём новый MR
		mrID, err = mergeBranchInto(CIMainBranch, ciBranch, repo)
		if err != nil {
			return fmt.Errorf("не удалось создать MR: %v", err)
		}
		log.Printf("Засыпаем...")
		time.Sleep(3 * time.Second)
		log.Printf("Просыпаемся...")
	}

	// 9. Принятие MR (фактическое слияние)
	if err := acceptMergeRequest(mrID, repo.ProjectId); err != nil {
		return fmt.Errorf("не удалось принять MR %d: %v", mrID, err)
	}

	// 10. Создание MR от CI‑ветки
	sendMessageByID(wsClientID, fmt.Sprintf("Создаём MR от CI-ветки `%s` в ветку стенда `%s`", ciBranch, repo.StandBranch), WSMessageTypeDefault)
	title := strings.TrimPrefix(ciBranch, PREFIX+CIPrefix)
	mrURL, err := createGitLabMR(ciBranch, title, repo)
	if err != nil {
		return err
	}

	c.String(http.StatusOK, mrURL)
	return nil
}

// Логика для кнопки "Смержить"
func handleMergeAction(branch string, c *gin.Context, project Repository) error {
	// 1. Проверка существования ветки в репозитории
	exists, err := branchExistsInRepo(branch, project.ProjectId)
	if err != nil {
		return fmt.Errorf("Ошибка проверки ветки в репозитории: %v", err)
	}
	if !exists {
		return fmt.Errorf("Ветка %s не найдена в репозитории", branch)
	}

	// 2. Проверка открытых MR
	hasMR, _, mrUrl, err := hasOpenMR(branch, project.StandBranch, project)
	if err != nil {
		return fmt.Errorf("Ошибка при проверке MR: %v", err)
	}
	if hasMR {
		return fmt.Errorf("Для ветки %s уже есть открытый MR в %s:\n<a href=''>%s</a>", branch, project.StandBranch, mrUrl)
	}

	return nil // Всё ок
}
