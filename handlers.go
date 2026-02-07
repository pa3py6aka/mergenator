package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	TARGET_BRANCH    = "andagar/r2533-bh-a1"
	ASSIGNEE_ID      = 67
	CI_COMMON_BRANCH = "andagar/CI"
)

func handleMerge(c *gin.Context) {
	sourceBranch := c.PostForm("source_branch")
	action := c.PostForm("action")

	switch action {
	case "createMR":
		err := handleCreateMR(sourceBranch, c)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		} else {
			// В handleCreateMR мы уже отправляем полный ответ (включая URL)
		}
	case "merge":
		err := handleMergeAction(sourceBranch, c)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		} else {
			c.String(http.StatusOK, fmt.Sprintf("Ветка %s готова к мерджу в %s.", sourceBranch, TARGET_BRANCH))
		}
	default:
		c.String(http.StatusBadRequest, "Неизвестное действие")
	}
}

// Логика для кнопки "Создать MR"
func handleCreateMR(branch string, c *gin.Context) error {
	// 1. Проверка префикса исходной ветки
	if err := validateBranchPrefix(branch); err != nil {
		return err
	}

	// 2. Проверка существования исходной ветки
	exists, err := branchExistsInRepo(branch)
	if err != nil {
		return fmt.Errorf("ошибка проверки ветки: %v", err)
	}
	if !exists {
		return fmt.Errorf("ветка %s не найдена", branch)
	}

	// 3. Формирование имени CI‑ветки
	ciBranch := makeCIBranchName(branch)

	// 4. Проверка существования CI‑ветки
	ciExists, err := branchExistsInRepo(ciBranch)
	if err != nil {
		return fmt.Errorf("Ошибка проверки CI‑ветки: %v", err)
	}

	// 5. Проверка открытого MR для CI‑ветки
	hasMR, err := hasOpenMergeRequestForCIBranch(ciBranch)
	if err != nil {
		return fmt.Errorf("Ошибка проверки MR для CI‑ветки: %v", err)
	}
	if hasMR {
		return fmt.Errorf(
			"Для CI‑ветки %s уже есть открытый MR в %s", ciBranch, TARGET_BRANCH)
	}

	// 6. Если CI‑ветка существует — удаляем её
	if ciExists {
		if err := deleteRemoteBranch(ciBranch); err != nil {
			return fmt.Errorf("не удалось удалить CI‑ветку %s: %v", ciBranch, err)
		}
	}

	// 7. Создание CI‑ветки от исходной
	if err := createRemoteBranch(branch, ciBranch); err != nil {
		return fmt.Errorf("не удалось создать CI‑ветку %s: %v", ciBranch, err)
	}

	// 8. Проверка: есть ли уже открытый MR для этих веток?
	hasMR, mrID, err := hasOpenMR(CI_COMMON_BRANCH, ciBranch)
	if err != nil {
		return fmt.Errorf("ошибка проверки существующих MR: %v", err)
	}
	if hasMR {
		// MR уже существует — не создаём новый, а используем существующий
		log.Printf("Уже есть открытый MR №%d для %s → %s", mrID, CI_COMMON_BRANCH, ciBranch)
	} else {
		// Создаём новый MR
		mrID, err = mergeBranchInto(CI_COMMON_BRANCH, ciBranch)
		if err != nil {
			return fmt.Errorf("не удалось создать MR: %v", err)
		}
	}

	// 9. Принятие MR (фактическое слияние)
	if err := acceptMergeRequest(mrID); err != nil {
		return fmt.Errorf("не удалось принять MR %d: %v", mrID, err)
	}

	// 10. Создание MR от CI‑ветки
	title := strings.TrimPrefix(ciBranch, PREFIX+CI_PREFIX)
	mrURL, err := createGitLabMR(ciBranch, title)
	if err != nil {
		return err
	}

	c.String(http.StatusOK, mrURL)
	return nil
}

// Логика для кнопки "Смержить"
func handleMergeAction(branch string, c *gin.Context) error {
	// 1. Проверка существования ветки в репозитории
	exists, err := branchExistsInRepo(branch)
	if err != nil {
		return fmt.Errorf("Ошибка проверки ветки в репозитории: %v", err)
	}
	if !exists {
		return fmt.Errorf("Ветка %s не найдена в репозитории", branch)
	}

	// 2. Проверка открытых MR
	hasMR, err := hasOpenMergeRequest(branch, TARGET_BRANCH)
	if err != nil {
		return fmt.Errorf("Ошибка при проверке MR: %v", err)
	}
	if hasMR {
		return fmt.Errorf("Для ветки %s уже есть открытый MR в %s", branch, TARGET_BRANCH)
	}

	return nil // Всё ок
}
