package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/goccy/go-json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// https://bmcgeriatr.biomedcentral.com/articles/10.1186/s12877-019-1160-9
type Facility struct {
	Id              string `json:"id" db:"id"`
	FacilityId      string `json:"facilityid" db:"facility_id"`
	FacilityCode    string `json:"facilitycode" db:"facility_code"`
	FacilityName    string `json:"facilityname" db:"facility_name"`
	FacilityAddr    string `json:"facilityaddr" db:"facility_addr"`
	FacilityTel     string `json:"facilitytel" db:"facility_tel"`
	LocalGovCode    string `json:"localgovcode" db:"local_gov_code"`
	ZipCode         string `json:"zipcode" db:"zipcode"`
	PrefName        string `json:"prefname" db:"pref_name"`
	CityName        string `json:"cityname" db:"city_name"`
	Latitude        string `json:"latitude" db:"latitude"`
	Longitude       string `json:"longitude" db:"longitude"`
	SubmitDate      string `json:"submitdate" db:"submit_date"`
	Hospitalization string `json:"hospitalization" db:"hospitalization"`
	Outpatient      string `json:"outpatient" db:"outpatient"`
	Emergency       string `json:"emergency" db:"emergency"`
	CreatedAt       string `json:"createdat" db:"created_at"`
	UpdatedAt       string `json:"updatedat" db:"updated_at"`
}

var (
	ErrNon200Response = errors.New("non 200 Response found")
	ErrNoJsonResponse = errors.New("no JsonResponse in HTTP response")
	Url               = "https://opendata.corona.go.jp/api/covid19DailySurvey"
	ENV               = os.Getenv("ENV")
	DBHOST            = os.Getenv("DBHOST")
	DBNAME            = os.Getenv("DBNAME")
	DBUSER            = os.Getenv("DBUSER")
	DBPASS            = os.Getenv("DBPASS")
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	db, err := sql.Open("mysql", "")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	var whereClauses []string
	queryPrefName := request.MultiValueQueryStringParameters["prefName"]
	for _, val := range queryPrefName {
		whereClauses = append(whereClauses, fmt.Sprintf("pref_name = '%s'", val))
	}
	queryCityName := request.MultiValueQueryStringParameters["cityName"]
	for _, val := range queryCityName {
		whereClauses = append(whereClauses, fmt.Sprintf("city_name = '%s'", val))
	}

	var whereClause string
	if len(whereClauses) > 0 {
		whereClause = " WHERE " + strings.Join(whereClauses, " OR ")
	}

	rows, err := db.Query("SELECT * FROM facility" + whereClause)
	if err != nil {
		panic(err)
	}

	facilityList := make([]Facility, 0)
	for rows.Next() {
		var facility Facility
		err = rows.Scan(
			&facility.Id,
			&facility.FacilityId,
			&facility.FacilityName,
			&facility.ZipCode,
			&facility.PrefName,
			&facility.FacilityAddr,
			&facility.FacilityTel,
			&facility.Latitude,
			&facility.Longitude,
			&facility.SubmitDate,
			&facility.LocalGovCode,
			&facility.CityName,
			&facility.FacilityCode,
			&facility.Hospitalization,
			&facility.Outpatient,
			&facility.Emergency,
			&facility.CreatedAt,
			&facility.UpdatedAt)
		if err != nil {
			log.Fatal(err)
		}
		facilityList = append(facilityList, facility)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}

	jsonBytes, err := json.Marshal(facilityList)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Body Size : %d Byte \n", len(jsonBytes))

	return events.APIGatewayProxyResponse{
		Body:       string(jsonBytes),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
