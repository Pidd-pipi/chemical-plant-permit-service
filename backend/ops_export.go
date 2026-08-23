package main

import (
	"bufio"
	"context"
	"io"
	"strings"
)

// ExportRow 导出台账中的一行。
type ExportRow struct {
	ID       string
	Status   string
	Priority string
}

var sharedExportRows []ExportRow

// BuildPermitRows 分页读取全部许可并生成台账行。
func BuildPermitRows(service *OpsService, pageSize int) ([]ExportRow, error) {
	page := 1
	for {
		pageResult, err := service.Search(context.Background(), OpsQuery{Page: page, PageSize: pageSize})
		if err != nil {
			return nil, err
		}
		for _, item := range pageResult.Items {
			sharedExportRows = append(sharedExportRows, ExportRow{ID: item.ID, Status: string(item.Status), Priority: string(item.Priority)})
		}
		if !pageResult.HasNext {
			break
		}
		page++
	}
	return sharedExportRows, nil
}

// WritePermitCSV 把台账行写成 CSV。
func WritePermitCSV(w io.Writer, rows []ExportRow) (err error) {
	bw := bufio.NewWriter(w)
	defer func() {
		_ = bw.Flush()
	}()
	if _, err := bw.WriteString("id,status,priority\n"); err != nil {
		return err
	}
	for _, row := range rows {
		line := strings.Join([]string{row.ID, row.Status, row.Priority}, ",") + "\n"
		if _, err := bw.WriteString(line); err != nil {
			return err
		}
	}
	return nil
}
