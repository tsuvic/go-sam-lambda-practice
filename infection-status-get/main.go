package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
)

// https://qiita.com/dondoko-susumu/items/7285eab65a9dfa9e73e8
type InfectionStatus struct {
	Date                        time.Time `json:"date"`
	Prefecture                  string    `json:"prefecture"`
	InfectionNumberDaily        int       `json:"infectionNumberDaily"`
	InfectionNumberCumulatively int       `json:"infectionNumberCumulatively"`
}

func openDB() (*sql.DB, error) {
	var (
		dbhost = os.Getenv("DBHOST")
		dbname = os.Getenv("DBNAME")
		dbuser = os.Getenv("DBUSER")
		dbpass = os.Getenv("DBPASS")
	)
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", dbuser, dbpass, dbhost, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// 複数日付の検索条件はOR、複数都道府県の検索条件はOR、日付と都道府県の検索条件はAND
func createWhereClause(req events.APIGatewayProxyRequest) (string, []interface{}, error) {
	var dataClause string
	qDate := req.MultiValueQueryStringParameters["date"]
	if len(qDate) > 0 {
		dataClauses := make([]string, len(qDate))
		for i, _ := range qDate {
			dataClauses[i] = "date = ?"
		}
		switch len(qDate) {
		case 1:
			dataClause = "(" + dataClauses[0] + ")"
		default:
			dataClause = "(" + strings.Join(dataClauses, " OR ") + ")"
		}
	}

	var prefectureClause string
	qPrefecture := req.MultiValueQueryStringParameters["prefecture"]
	if len(qPrefecture) > 0 {
		prefectureClauses := make([]string, len(qPrefecture))
		for i, _ := range qPrefecture {
			prefectureClauses[i] = "prefecture = ?"
		}
		switch len(qPrefecture) {
		case 1:
			prefectureClause = "(" + prefectureClauses[0] + ")"
		default:
			prefectureClause = "(" + strings.Join(prefectureClauses, " OR ") + ")"
		}
	}

	var whereClause string
	if len(qDate) > 0 && len(qPrefecture) > 0 {
		whereClause = " WHERE " + dataClause + " AND " + prefectureClause
	} else if len(qDate) > 0 && len(qPrefecture) == 0 {
		whereClause = " WHERE " + dataClause
	} else if len(qDate) == 0 && len(qPrefecture) > 0 {
		whereClause = " WHERE " + prefectureClause
	}

	query := make([]interface{}, 0, len(qDate)+len(qPrefecture))
	for _, val := range qDate {
		query = append(query, val)
	}
	for _, val := range qPrefecture {
		query = append(query, val)
	}

	return whereClause, query, nil
}

func getStatusDaily(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	db, err := openDB()
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	whereClause, query, err := createWhereClause(req)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	stmt, err := db.Prepare(fmt.Sprintf("SELECT date, prefecture, infection_number_daily, infection_number_cumulatively FROM infection_status %s", whereClause))
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	rows, err := stmt.Query(query...)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	infectionStatusList := make([]InfectionStatus, 0)
	for rows.Next() {
		//https://github.com/mattn/go-sqlite3/issues/190

		var infectionStatus InfectionStatus
		if err = rows.Scan(
			&infectionStatus.Date,
			&infectionStatus.Prefecture,
			&infectionStatus.InfectionNumberDaily,
			&infectionStatus.InfectionNumberCumulatively,
		); err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		infectionStatusList = append(infectionStatusList, infectionStatus)
	}
	if err = rows.Err(); err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	bytes, err := json.Marshal(infectionStatusList)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	fmt.Printf("Body Size : %d Byte \n", len(bytes))

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(bytes),
	}, nil
}

func getStatusCumulatively(events.APIGatewayProxyRequest) {}
func getStatusComparison(events.APIGatewayProxyRequest)   {}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var res events.APIGatewayProxyResponse
	var err error
	switch req.PathParameters["type"] {
	case "daily":
		res, err = getStatusDaily(req)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
	case "cumuratively":
		getStatusCumulatively(req)
	case "comparison":
		getStatusComparison(req)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: res.StatusCode,
		Body:       res.Body,
	}, nil
}

func main() {
	lambda.Start(handler)
}
