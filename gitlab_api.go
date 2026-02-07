package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	GITLAB_API_URL = ""
	PROJECT_ID     = ""
	ACCESS_TOKEN   = ""
)

// Проверка существования ветки в репозитории через API
func branchExistsInRepo(branch string) (bool, error) {
	escapedBranch := url.PathEscape(branch)
	branchesUrl := fmt.Sprintf(
		"%s/projects/%s/repository/branches/%s",
		GITLAB_API_URL, PROJECT_ID, escapedBranch)

	req, err := http.NewRequest("GET", branchesUrl, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Private-Token", ACCESS_TOKEN)

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf(
			"API ошибка: %d %s", resp.StatusCode, string(body))
	}
}

// Проверка открытых MR
func hasOpenMergeRequest(sourceBranch string, targetBranch string) (bool, error) {
	url := fmt.Sprintf(
		"%s/projects/%s/merge_requests?source_branch=%s&target_branch=%s&state=opened",
		GITLAB_API_URL, PROJECT_ID, sourceBranch, targetBranch)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Private-Token", ACCESS_TOKEN)

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	return len(body) > 2, nil // [] — пустой массив
}

// Проверяет, есть ли открытый MR от CI‑ветки в целевую ветку
func hasOpenMergeRequestForCIBranch(ciBranch string) (bool, error) {
	url := fmt.Sprintf(
		"%s/projects/%s/merge_requests?source_branch=%s&target_branch=%s&state=opened",
		GITLAB_API_URL, PROJECT_ID, url.PathEscape(ciBranch), url.PathEscape(TARGET_BRANCH))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Private-Token", ACCESS_TOKEN)

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("API ошибка: %d %s", resp.StatusCode, string(body))
	}

	var mrList []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&mrList); err != nil {
		return false, err
	}

	return len(mrList) > 0, nil
}

func hasOpenMR(source, target string) (bool, int, error) {
	url := fmt.Sprintf(
		"%s/projects/%s/merge_requests?source_branch=%s&target_branch=%s&state=opened",
		GITLAB_API_URL, PROJECT_ID, source, target)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, 0, err
	}

	req.Header.Set("Private-Token", ACCESS_TOKEN)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return false, 0, fmt.Errorf("API ошибка проверки MR: %d %s", resp.StatusCode, string(body))
	}

	var mrs []struct {
		ID int `json:"iid"`
	}
	if err := json.Unmarshal(body, &mrs); err != nil {
		return false, 0, err
	}

	if len(mrs) > 0 {
		return true, mrs[0].ID, nil // Возвращаем ID первого найденного MR
	}

	return false, 0, nil
}

// Создание MR
func createGitLabMR(sourceBranch, title string) (string, error) {
	url := fmt.Sprintf("%s/projects/%s/merge_requests", GITLAB_API_URL, PROJECT_ID)

	data := map[string]interface{}{
		"source_branch":        sourceBranch,
		"target_branch":        TARGET_BRANCH,
		"title":                title,
		"assignee_ids":         []int{ASSIGNEE_ID},
		"remove_source_branch": true,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Private-Token", ACCESS_TOKEN)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API ошибка: %d %s", resp.StatusCode, string(body))
	}

	var mrResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&mrResponse); err != nil {
		return "", err
	}

	mrURL, ok := mrResponse["web_url"].(string)
	if !ok {
		return "", fmt.Errorf("не удалось извлечь URL MR")
	}

	return mrURL, nil
}

// Создание ветки в удалённом репозитории
func createRemoteBranch(sourceBranch, newBranch string) error {
	url := fmt.Sprintf("%s/projects/%s/repository/branches",
		GITLAB_API_URL, PROJECT_ID)

	data := map[string]interface{}{
		"branch": newBranch,
		"ref":    sourceBranch, // исходная ветка/коммит
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Private-Token", ACCESS_TOKEN)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Не удалось создать ветку %s: %d %s", newBranch, resp.StatusCode, string(body))
	}

	return nil
}

// Удаление ветки в удалённом репозитории
func deleteRemoteBranch(branch string) error {
	url := fmt.Sprintf("%s/projects/%s/repository/branches/%s",
		GITLAB_API_URL, PROJECT_ID, url.PathEscape(branch))

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Private-Token", ACCESS_TOKEN)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Не удалось удалить ветку %s: %d %s", branch, resp.StatusCode, string(body))
	}

	return nil
}

// Сливает указанную ветку (source) в целевую (targetBranch) через GitLab API
func mergeBranchInto(source, targetBranch string) (int, error) {
	url := fmt.Sprintf(
		"%s/projects/%s/merge_requests", GITLAB_API_URL, PROJECT_ID)

	data := map[string]interface{}{
		"source_branch": source,
		"target_branch": targetBranch,
		"title":         fmt.Sprintf("Merge %s into %s", source, targetBranch),
		"description":   "Автоматическое слияние CI-зависимостей",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Private-Token", ACCESS_TOKEN)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API ошибка слияния: %d %s", resp.StatusCode, string(body))
	}

	// Декодируем ответ, чтобы получить mrID
	var mrResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&mrResponse); err != nil {
		return 0, err
	}

	mrID, ok := mrResponse["iid"].(float64)
	if !ok {
		return 0, fmt.Errorf("не удалось извлечь ID MR из ответа")
	}

	return int(mrID), nil
}

func acceptMergeRequest(mrID int) error {
	url := fmt.Sprintf(
		"%s/projects/%s/merge_requests/%d/merge", GITLAB_API_URL, PROJECT_ID, mrID)

	// Тело запроса (даже если параметры не нужны)
	data := map[string]interface{}{
		"merge_when_pipeline_succeeds": false,
		"should_remove_source_branch":  false,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Private-Token", ACCESS_TOKEN)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("API ошибка принятия MR: %d %s", resp.StatusCode, string(body))
	}

	return nil
}
