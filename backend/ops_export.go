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
// 行只写入本次调用返回的切片，不保留任何包级状态，避免多次导出之间数据串段。
func BuildPermitRows(service *OpsService, pageSize int) ([]ExportRow, error) {
	rows := make([]ExportRow, 0)
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

// WritePermitCSV 把台账行写成 CSV。任何底层写入错误都会原样返回，
// 不会因为 bufio 缓冲而被吞掉。
func WritePermitCSV(w io.Writer, rows []ExportRow) error {
	bw := bufio.NewWriter(w)
	if _, err := bw.WriteString("id,status,priority\n"); err != nil {
		return err
	}
	for _, row := range rows {
		line := strings.Join([]string{row.ID, row.Status, row.Priority}, ",") + "\n"
		if _, err := bw.WriteString(line); err != nil {
			return err
		}
	}
	// Flush 是最后一步写操作，必须把它的错误回报给调用方。
	return bw.Flush()
}
