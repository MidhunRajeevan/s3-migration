package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/MidhunRajeevan/s3-migration/config"
	"github.com/minio/minio-go/v7"
)

var (
	stopChan         = make(chan bool)
	isRunning        = false
	isPaused         = false
	logFilePath      = "migration.log"
	faildLogFilePath = "failed_files.log"
	logFile          *os.File
	logFileMutex     sync.Mutex
)

func StartMigrationHandler(w http.ResponseWriter, r *http.Request) {
	if isRunning {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Migration is already running"))
		return
	}

	err := openLogFile()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to open log file"))
		return
	}
	defer closeLogFile()

	go startMigration()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Migration started"))
}

func StopMigrationHandler(w http.ResponseWriter, r *http.Request) {
	if !isRunning {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Migration is not running"))
		return
	}

	stopChan <- true
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Migration stopping..."))
}

func startMigration() {
	isRunning = true
	isPaused = false
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		select {
		case <-stopChan:
			cancel()
			isRunning = false
			writeLog("Migration stopped")
		}
	}()

	err := migrateDirectories(ctx)
	if err != nil {
		writeLog(fmt.Sprintf("Migration failed: %v", err))
	} else {
		writeLog("Migration completed successfully")
	}

	isRunning = false
}

func migrateDirectories(ctx context.Context) error {
	directories, err := SelectDirectories()
	if err != nil {
		log.Println("Select directories Error:", err.Error())
		return err
	}

	for _, dir := range directories {
		select {
		case <-ctx.Done():
			log.Println("Migration stopped by context cancellation")
			return ctx.Err()
		default:
			if DirectoryMigrated(dir.Did) {
				log.Printf("Directory %s already migrated, skipping...", dir.Did)
				continue
			}
			err := MarkDirectoryAsStarted(dir)
			if err != nil {
				log.Printf("Failed to update directory start time for %s: %v", dir.Did, err)
				continue
			}
			log.Printf("Migrating directory: %s", dir.Did)

			err = migrateFilesInDirectory(ctx, dir.Did)
			if err != nil {
				log.Printf("Migration failed for directory %s: %v", dir.Did, err)
				continue
			}
			err = MarkDirectoryAsCompleted(dir.Did)
			if err != nil {
				log.Printf("Failed to update directory completion time for %s: %v", dir.Did, err)
			}

			log.Printf("Successfully migrated directory: %s", dir.Did)
		}
	}

	return nil
}

func migrateFilesInDirectory(ctx context.Context, directory string) error {
	sourceClient := config.SourceClient
	doneCh := make(chan struct{})
	defer close(doneCh)

	objectCh := sourceClient.ListObjects(ctx, config.Source.Bucket, minio.ListObjectsOptions{
		Prefix:    directory,
		Recursive: true,
	})

	var wg sync.WaitGroup
	for {
		select {
		case object, ok := <-objectCh:
			if !ok {
				wg.Wait() // Wait for all ongoing migrations to finish
				return nil
			}
			if object.Err != nil {
				return object.Err
			}

			wg.Add(1)
			go func(object minio.ObjectInfo) {
				defer wg.Done()
				select {
				case <-ctx.Done():
					return
				default:
					log.Printf("Migrating file: %s", object.Key)
					err := migrateObject(object.Key)
					if err != nil {
						log.Printf("Failed to migrate file %s: %v", object.Key, err)
						logFailedFile(object.Key)
						return
					}

					markFileAsMigrated(object.Key)
				}
			}(object)
		case <-ctx.Done():
			wg.Wait() // Wait for all ongoing migrations to finish
			return ctx.Err()
		}
	}
}

func migrateObject(objectKey string) error {

	sourceClient := config.SourceClient
	targetClient := config.TargetClient
	ctx := context.Background()

	// Retrieve the object from Nuba S3
	object, err := sourceClient.GetObject(ctx, config.Source.Bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object from Nuba S3: %v", err)
	}
	defer object.Close()

	// Get object info
	objInfo, err := object.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat object: %v", err)
	}

	// Put object to AWS S3
	_, err = targetClient.PutObject(ctx, config.Target.Bucket, objectKey, object, objInfo.Size, minio.PutObjectOptions{ContentType: objInfo.ContentType})
	if err != nil {
		return fmt.Errorf("failed to put object to AWS S3: %v", err)
	}

	return nil
}

func logFailedFile(objectKey string) {
	file, err := os.OpenFile(faildLogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open failed files log: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(fmt.Sprintf("Failed to migrate file: %s\n", objectKey)); err != nil {
		log.Fatalf("Failed to write to failed files log: %v", err)
	}
}

func markFileAsMigrated(objectKey string) {
	writeLog(fmt.Sprintf("File %s migrated", objectKey))
}

func openLogFile() error {
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	return err
}

func closeLogFile() {
	if logFile != nil {
		logFile.Close()
	}
}

func writeLog(message string) {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	if logFile != nil {
		logFile.WriteString(message + "\n")
		logFile.Sync() // Ensure data is written to disk
	}
}
