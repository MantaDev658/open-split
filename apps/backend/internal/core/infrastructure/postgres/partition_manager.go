package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// run a daily background job to manage audit log lifecycle
func StartPartitionManager(ctx context.Context, db *sql.DB, retentionMonths int) {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				managePartitions(db, retentionMonths)
			}
		}
	}()
	// run once on startup
	managePartitions(db, retentionMonths)
}

func managePartitions(db *sql.DB, retentionMonths int) {
	now := time.Now().UTC()

	// create partition for NEXT month
	nextMonth := now.AddDate(0, 1, 0)
	startOfNext := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfNext := startOfNext.AddDate(0, 1, 0)

	tableName := fmt.Sprintf("audit_logs_y%dm%02d", startOfNext.Year(), startOfNext.Month())

	// #nosec G201 -- DDL statements cannot use parameters; inputs are strictly controlled internal dates
	createSql := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s PARTITION OF audit_logs 
		FOR VALUES FROM ('%s') TO ('%s');
	`, tableName, startOfNext.Format("2006-01-02"), endOfNext.Format("2006-01-02"))

	if _, err := db.Exec(createSql); err != nil {
		log.Printf("Failed to create audit partition %s: %v", tableName, err)
	}

	// drop partition from 6 months ago
	expiredDate := now.AddDate(0, -retentionMonths, 0)
	expiredTableName := fmt.Sprintf("audit_logs_y%dm%02d", expiredDate.Year(), expiredDate.Month())

	// #nosec G201 -- DDL statements cannot use parameters
	dropSql := fmt.Sprintf("DROP TABLE IF EXISTS %s;", expiredTableName)
	if _, err := db.Exec(dropSql); err != nil {
		log.Printf("Failed to drop expired partition %s: %v", expiredTableName, err)
	}
}
