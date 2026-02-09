package main

import (
	"fmt"
	"strings"
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
	suffix := strings.TrimPrefix(branch, Prefix)
	return Prefix + CIPrefix + suffix
}
