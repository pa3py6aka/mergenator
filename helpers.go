package main

import "fmt"

type Link struct {
	Href string
	Text string
}

func getLink(link Link) string {
	text := link.Text
	if text == "" {
		text = link.Href
	}
	return fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s</a>", link.Href, text)
}

func getProjectNameByID(projectId string) string {
	switch projectId {
	case "141":
		return "FrontEnd"
	case "140":
		return "BackEnd"
	}

	return "UndefinedProject"
}

func getStandBranchByProjectID(projectId string) string {
	switch projectId {
	case "140":
		return BackendStandBranch
	case "141":
		return FrontendStandBranch
	}

	return "UndefinedProject"
}
