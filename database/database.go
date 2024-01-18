package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type File struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Size       int       `json:"size"`
	UploadDate time.Time `json:"uploaddate"`
}

var DB *sql.DB

func SetDB() {
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlPort := os.Getenv("MYSQL_PORT")

	var err error
	DB, err = sql.Open("mysql", "root:rootpass@tcp("+mysqlHost+":"+mysqlPort+")/mydatabase?parseTime=True")
	if err != nil {
		fmt.Printf("데이터베이스 연결 실패:%s\n", err)
		os.Exit(1)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS files (
		ID INTEGER PRIMARY KEY AUTO_INCREMENT,
		Name TEXT,
		Path TEXT,
		Size INTEGER,
		UploadDate DATETIME
	)`)
	if err != nil {
		fmt.Printf("데이터베이스 테이블 생성 실패:%s\n", err)
		os.Exit(1)
	}

	fmt.Printf("데이터베이스 연결 완료\n")
}

func GetFiles() ([]File, error) {
	rows, err := DB.Query("SELECT * FROM files")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var file File
		err := rows.Scan(&file.ID, &file.Name, &file.Path, &file.Size, &file.UploadDate)
		if err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	return files, nil
}

func GetFileByName(fileName string) (*File, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := "SELECT ID, Name, Path, Size, UploadDate FROM files WHERE NAME = ?"
	row := DB.QueryRow(query, fileName)

	var file File
	err := row.Scan(&file.ID, &file.Name, &file.Path, &file.Size, &file.UploadDate)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &file, nil
}

func SaveFileToDB(name string, path string) (File, error) {
	if DB == nil {
		return File{}, fmt.Errorf("database connection is nil")
	}

	stat, err := os.Stat(path)
	if err != nil {
		fmt.Printf("그만해그만~~~\n")
		return File{}, err
	}

	size := int(stat.Size())
	uploadDate := time.Now()

	query := "INSERT INTO files (Name, Path, Size, UploadDate) VALUES (?, ?, ?, ?)"
	_, err = DB.Exec(query, name, path, size, uploadDate)
	if err != nil {
		return File{}, fmt.Errorf("error inserting file into database: %v", err)
	}

	fmt.Printf("File '%s' saved to the database.\n", name)
	return File{
		Name:       name,
		Path:       path,
		Size:       size,
		UploadDate: uploadDate,
	}, nil
}
