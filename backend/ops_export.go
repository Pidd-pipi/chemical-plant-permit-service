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

// BuildPermitRows 分页读取全部许可并生成台账行。
func BuildPermitRows(service *OpsService, pageSize int) ([]ExportRow, error) {
	rows := make([]ExportRow, 0, 64)
	page := 1
	for {
		pageResult, err := service.Search(context.Background(), OpsQuery{Page: page, PageSize: pageSize})
		if err != nil {
			return nil, err
		}
		for _, item := range pageResult.Items {
			rows = append(rows, ExportRow{ID: item.ID, Status: string(item.Status), Priority: string(item.Priority)})
		}
		if !pageResult.HasNext {
			break
		}
		page++
	}
	return rows, nil
}

// WritePermitCSV 把台账行写成 CSV；flush 失败也必须返回错误。
func WritePermitCSV(w io.Writer, rows []ExportRow) (err error) {
	bw := bufio.NewWriter(w)
	defer func() {
		if flushErr := bw.Flush(); flushErr != nil && err == nil {
			err = flushErr
		}
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
