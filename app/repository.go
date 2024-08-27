package app

import (
	"log"

	"github.com/MidhunRajeevan/s3-migration/config"
)

func SelectDirectories() ([]DirectoryRecord, error) {
	w := make([]DirectoryRecord, 0)
	statement := `
	select id, did, total_files, status, started_at, completed_at 
	from roles where status='pending'`
	rows, err := config.DB.Query(statement)
	if err != nil {
		log.Println("Database Select Error:", err.Error())
	}
	for rows.Next() {
		r := DirectoryRecord{}
		err = rows.Scan(&r.ID, &r.Did, &r.Totalfiles, &r.StartedAt, &r.CompletedAt, &r.Status)
		if err != nil {
			return nil, err
		}
		w = append(w, r)
	}
	return w, err
}

func MarkDirectoryAsStarted(dir DirectoryRecord) error {
	_, err := config.DB.Exec(`
		UPDATE directory
		SET status = 'in_progress', started_at = now()
		WHERE did = $1
	`, dir.Did)
	return err
}

func MarkDirectoryAsCompleted(did string) error {
	_, err := config.DB.Exec(`
		UPDATE directory
		SET status = 'completed', completed_at = now()
		WHERE did = $1
	`, did)
	return err
}

func DirectoryMigrated(did string) bool {

	// Prepare the SQL statement to get directory status
	statement := `
		SELECT status
		FROM directory
		WHERE did = $1`

	// Execute the query
	var status string
	err := config.DB.QueryRow(statement, did).Scan(&status)
	if err != nil {
		return false
	}
	return status == "completed"
}
