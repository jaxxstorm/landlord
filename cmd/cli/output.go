package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jaxxstorm/landlord/internal/api/models"
)

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#04B575"))
	errorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5F5F"))
	labelStyle   = lipgloss.NewStyle().Bold(true)
)

func renderTenantList(tenants []models.TenantResponse) string {
	headers := []string{"ID", "Name", "Status", "Workflow", "Retries"}
	rows := make([][]string, 0, len(tenants))

	for _, t := range tenants {
		workflow := ""
		if t.WorkflowSubState != nil {
			workflow = *t.WorkflowSubState
		}
		retries := ""
		if t.WorkflowRetryCount != nil {
			retries = fmt.Sprintf("%d", *t.WorkflowRetryCount)
		}
		rows = append(rows, []string{t.ID, t.Name, formatStatus(t.Status), workflow, retries})
	}

	widths := columnWidths(headers, rows)
	var lines []string
	lines = append(lines, headerStyle.Render(formatRow(headers, widths)))
	for _, row := range rows {
		lines = append(lines, formatRow(row, widths))
	}

	return strings.Join(lines, "\n")
}

func renderTenantDetails(tenant models.TenantResponse) string {
	lines := []string{
		fmt.Sprintf("%s %s", labelStyle.Render("ID:"), tenant.ID),
		fmt.Sprintf("%s %s", labelStyle.Render("Name:"), tenant.Name),
		fmt.Sprintf("%s %s", labelStyle.Render("Status:"), formatStatus(tenant.Status)),
	}

	if tenant.StatusMessage != "" {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Status Message:"), tenant.StatusMessage))
	}

	if tenant.WorkflowExecutionID != nil && *tenant.WorkflowExecutionID != "" {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Workflow Execution ID:"), *tenant.WorkflowExecutionID))
	}

	if tenant.WorkflowSubState != nil && *tenant.WorkflowSubState != "" {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Workflow Sub-State:"), *tenant.WorkflowSubState))
	}

	if tenant.WorkflowRetryCount != nil {
		lines = append(lines, fmt.Sprintf("%s %d", labelStyle.Render("Workflow Retry Count:"), *tenant.WorkflowRetryCount))
	}

	if tenant.WorkflowErrorMessage != nil && *tenant.WorkflowErrorMessage != "" {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Workflow Error:"), *tenant.WorkflowErrorMessage))
	}

	if len(tenant.DesiredConfig) > 0 {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Config:"), formatMap(tenant.DesiredConfig)))
	}

	if len(tenant.ComputeConfig) > 0 {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Compute:"), formatMap(tenant.ComputeConfig)))
	}

	if len(tenant.ObservedConfig) > 0 {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Observed Config:"), formatMap(tenant.ObservedConfig)))
	}

	if len(tenant.ObservedResourceIDs) > 0 {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Observed Resources:"), formatMap(tenant.ObservedResourceIDs)))
	}

	if len(tenant.Labels) > 0 {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Labels:"), formatMap(tenant.Labels)))
	}

	if len(tenant.Annotations) > 0 {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Annotations:"), formatMap(tenant.Annotations)))
	}

	if !tenant.CreatedAt.IsZero() {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Created At:"), tenant.CreatedAt.Format(time.RFC3339)))
	}

	if !tenant.UpdatedAt.IsZero() {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Updated At:"), tenant.UpdatedAt.Format(time.RFC3339)))
	}

	if tenant.Version != 0 {
		lines = append(lines, fmt.Sprintf("%s %d", labelStyle.Render("Version:"), tenant.Version))
	}

	return strings.Join(lines, "\n")
}

func renderComputeConfigDiscovery(resp models.ComputeConfigDiscoveryResponse) string {
	lines := []string{
		fmt.Sprintf("%s %s", labelStyle.Render("Provider:"), resp.Provider),
	}

	if len(resp.Schema) > 0 {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Schema:"), formatJSONRaw(resp.Schema)))
	}

	if len(resp.Defaults) > 0 {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Defaults:"), formatJSONRaw(resp.Defaults)))
	}

	return strings.Join(lines, "\n")
}

func formatStatus(status string) string {
	switch status {
	case "ready":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Render(status)
	case "failed":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F5F")).Render(status)
	case "deleting":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F5A623")).Render(status)
	case "archiving":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F5A623")).Render(status)
	default:
		return status
	}
}

func formatMap(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}

func formatJSONRaw(value json.RawMessage) string {
	if len(value) == 0 {
		return ""
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, value, "", "  "); err == nil {
		return buf.String()
	}
	return string(value)
}

func columnWidths(headers []string, rows [][]string) []int {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	return widths
}

func formatRow(cells []string, widths []int) string {
	parts := make([]string, 0, len(cells))
	for i, cell := range cells {
		parts = append(parts, padRight(cell, widths[i]+2))
	}
	return strings.TrimRight(strings.Join(parts, ""), " ")
}

func padRight(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return fmt.Sprintf("%-*s", width, value)
}
