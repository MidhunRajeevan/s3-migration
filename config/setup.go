package config

import (
	"log"
)

func createDirectory() error {
	statement := `
		create table if not exists directory (
			id            bigserial primary key,
			did           text not null unique,
			total_files   int,
			status        text not null default 'pending',
			started_at    timestamptz,
			completed_at  timestamptz
		)`
	if _, err := DB.Exec(statement); err != nil {
		log.Println(err)
		panic("Create table DIRECTORY failed!")
	}

	if _, err := DB.Exec(statement); err != nil {
		log.Println(err)
		panic("Create index on directory failed!")
	}

	return nil
}

// Setup database
func Setup() {
	createDirectory()
}
