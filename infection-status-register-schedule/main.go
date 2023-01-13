package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	errNon200Response = errors.New("non 200 Response found")
	errNoJsonResponse = errors.New("no JsonResponse in HTTP response")
	endpoint          = "https://opendata.corona.go.jp/api/Covid19JapanAll"
)

type InfectionStatusTmp struct {
	ErrorInfo struct {
		ErrorFlag    string `json:"errorFlag"`
		ErrorCode    string `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	} `json:"errorInfo"`
	ItemList []struct {
		Date      string `json:"date"`
		NameJp    string `json:"name_jp"`
		NPatients string `json:"npatients"`
	} `json:"itemList"`
}

type InfectionStatus struct {
	Date                        time.Time `json:"date"`
	Prefecture                  string    `json:"prefecture"`
	InfectionNumberDaily        int       `json:"infectionNumberDaily"`
	InfectionNumberCumulatively int       `json:"infectionNumberCumulatively"`
}

type Key struct {
	Date       time.Time
	Prefecture string
}

func openDB() (*sql.DB, error) {
	var (
		dbhost = os.Getenv("DBHOST")
		dbname = os.Getenv("DBNAME")
		dbuser = os.Getenv("DBUSER")
		dbpass = os.Getenv("DBPASS")
	)

	//https://github.com/go-sql-driver/mysql/issues/9#issuecomment-51552649
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

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//クエリパラメータ
	base, err := url.Parse(endpoint)
	if err != nil {
		return events.APIGatewayProxyResponse{}, nil
	}
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	param := yesterday.Format("20060102")
	query := url.Values{}
	query.Add("date", param)
	base.RawQuery = query.Encode()

	// //パスパラメータ
	// endpoint, err := url.JoinPath(endpoint, param)
	// if err != nil {
	// 	return events.APIGatewayProxyResponse{}, nil
	// }

	res, err := http.Get(base.String())
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	if res.StatusCode != 200 {
		return events.APIGatewayProxyResponse{}, errNon200Response
	}

	JsonResponse, err := io.ReadAll(res.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	if len(JsonResponse) == 0 {
		return events.APIGatewayProxyResponse{}, errNoJsonResponse
	}

	//https://budougumi0617.github.io/2019/02/24/go-print-detail-of-json-syntax-error/
	var infectionStatusTmp InfectionStatusTmp
	if err = json.Unmarshal(JsonResponse, &infectionStatusTmp); err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	db, err := openDB()
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	var infectionStatusList []InfectionStatus
	for _, val := range infectionStatusTmp.ItemList {
		var infectionStatus InfectionStatus

		//都道府県
		infectionStatus.Prefecture = val.NameJp

		//日付
		layout := "2006-01-02"
		t, err := time.Parse(layout, val.Date)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
		infectionStatus.Date = t

		//累積 感染者数
		cumulative, err := strconv.Atoi(val.NPatients)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
		infectionStatus.InfectionNumberCumulatively = cumulative

		//日次 感染者数
		q := "SELECT infection_number_cumulatively FROM infection_status WHERE date = ? AND prefecture = ? LIMIT 1"
		var i int
		dayBefore := infectionStatus.Date.AddDate(0, 0, -1)
		if err := db.QueryRow(q, dayBefore, infectionStatus.Prefecture).Scan(&i); err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
		infectionStatus.InfectionNumberDaily = infectionStatus.InfectionNumberCumulatively - i

		infectionStatusList = append(infectionStatusList, infectionStatus)
	}

	insert, err := db.Prepare(fmt.Sprintf("INSERT INTO %s (date, prefecture, infection_number_daily, infection_number_cumulatively) VALUES (?, ?, ?, ?)", "infection_status"))
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	var rowsAffectedSum int64
	for _, val := range infectionStatusList {
		res, err := insert.Exec(val.Date, val.Prefecture, val.InfectionNumberDaily, val.InfectionNumberCumulatively)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
		rowsAffectedSum += rowsAffected
	}

	body := fmt.Sprintf("Inserted %d rows.\n", rowsAffectedSum)

	bytes, err := json.Marshal(infectionStatusList)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	fmt.Printf("Body Size : %d Byte \n", len(bytes))

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       body,
	}, nil
}

func main() {
	lambda.Start(handler)
}
