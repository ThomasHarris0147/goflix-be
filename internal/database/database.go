package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

const (
	host        = "localhost"
	pg_port     = 15432
	user        = "root"
	pg_password = "password"
	dbname      = "videos"
)

func ConnectToDB() (*sql.DB, error) {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, pg_port, user, pg_password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected!")
	return db, err
}

func InitVideosDB() error {
	db, err := ConnectToDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	queryString := ""
	_, queryErr := db.Query(queryString)
	if queryErr != nil {
		return queryErr
	}
	return nil
}

func InsertInto(tableName string, columns []string, values []string) error {
	db, err := ConnectToDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var withSingleQuotes []string
	for _, value := range values {
		withSingleQuotes = append(withSingleQuotes, "'"+value+"'")
	}
	queryString := "INSERT INTO " + tableName + " (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(withSingleQuotes, ", ") + ");"
	fmt.Println(queryString)
	rows, queryErr := db.Query(queryString)
	if queryErr != nil {
		return queryErr
	}
	defer rows.Close()
	result := make([]string, 0)
	for rows.Next() {
		var row string
		if err := rows.Scan(&row); err != nil {
			return err
		}
		result = append(result, row)
	}
	if len(result) > 1 {
		panic("too many results: " + strings.Join(result, ", "))
	}
	fmt.Println("Inserted Successfully")
	return nil
}

func getResultsFromQuery(rows *sql.Rows) ([]map[string]interface{}, error) {
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Prepare a slice to hold results
	var result []map[string]interface{}

	for rows.Next() {
		// Create a slice to hold values for each column
		columnValues := make([]interface{}, len(columns))
		// Create a slice of pointers to populate with row values
		columnPointers := make([]interface{}, len(columns))

		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		// Scan the current row into the pointers
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		// Create a map for the current row
		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			rowMap[colName] = columnValues[i]
		}

		// Append the map to the result
		result = append(result, rowMap)
	}
	return result, nil
}

type VideoResponse struct {
	Name        string
	Description string
	Path        string
}

func GetVideoBasedOnNameAndQuality(name string, quality string) ([]map[string]interface{}, error) {
	db, dbErr := ConnectToDB()
	if dbErr != nil {
		return nil, dbErr
	}
	queryString := "SELECT * FROM videos WHERE name = '" + name + "' AND quality = '" + quality + "';"
	log.Println("Query: " + queryString)
	rows, queryErr := db.Query(queryString)
	if queryErr != nil {
		return nil, queryErr
	}
	result, err := getResultsFromQuery(rows)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetAllFromTable(tableName string) ([]map[string]interface{}, error) {
	db, connectToDBErr := ConnectToDB()
	if connectToDBErr != nil {
		return nil, connectToDBErr
	}
	defer db.Close()

	// Query all rows from the table
	rows, queryErr := db.Query("SELECT * FROM " + tableName)
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()

	result, err := getResultsFromQuery(rows)

	if err != nil {
		return nil, err
	}

	return result, nil
}
