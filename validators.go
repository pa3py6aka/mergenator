package main

import (
	"fmt"
	"strings"
)

const (
	RequiredPrefix = "r2533-andagar/VRPROC-"
	CIPrefix       = "CI-"
	PREFIX         = "r2533-andagar/"
)

// Проверка префикса ветки
func validateBranchPrefix(branch string) error {
	if !strings.HasPrefix(branch, RequiredPrefix) {
		return fmt.Errorf(
			"Ветка должна начинаться с префикса '%s'", RequiredPrefix)
	}
	return nil
}

// Формирование CI-ветки
func makeCIBranchName(branch string) string {
	suffix := strings.TrimPrefix(branch, PREFIX)
	return PREFIX + CIPrefix + suffix
}
