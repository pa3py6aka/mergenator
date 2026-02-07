package main

import (
	"fmt"
	"strings"
)

const (
	REQUIRED_PREFIX = "r2533-andagar/VRPROC-"
	CI_PREFIX       = "CI-"
	PREFIX          = "r2533-andagar/"
)

// Проверка префикса ветки
func validateBranchPrefix(branch string) error {
	if !strings.HasPrefix(branch, REQUIRED_PREFIX) {
		return fmt.Errorf(
			"Ветка должна начинаться с префикса '%s'", REQUIRED_PREFIX)
	}
	return nil
}

// Формирование CI-ветки
func makeCIBranchName(branch string) string {
	suffix := strings.TrimPrefix(branch, PREFIX)
	return PREFIX + CI_PREFIX + suffix
}
