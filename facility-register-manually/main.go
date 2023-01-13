package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// https://bmcgeriatr.biomedcentral.com/articles/10.1186/s12877-019-1160-9
type Facility struct {
	FacilityId      string `json:"facilityid"`
	FacilityName    string `json:"facilityname"`
	ZipCode         string `json:"zipcode"`
	PrefName        string `json:"prefname"`
	FacilityAddr    string `json:"facilityaddr"`
	FacilityTel     string `json:"facilitytel"`
	Latitude        string `json:"latitude"`
	Longitude       string `json:"longitude"`
	SubmitDate      string `json:"submitdate"`
	LocalGovCode    string `json:"localgovcode"`
	CityName        string `json:"cityname"`
	FacilityCode    string `json:"facilitycode"`
	FacilityType    string `json:"facilitytype"`
	AnsType         string `json:"anstype"`
	Hospitalization string `json:"hospitalization"`
	Outpatient      string `json:"outpatient"`
	Emergency       string `json:"emergency"`
}

var (
	ErrNon200Response = errors.New("non 200 Response found")
	ErrNoJsonResponse = errors.New("no JsonResponse in HTTP response")
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// var cmd = exec.Command("ls", "-la")
	// var result, _ = cmd.Output()
	// fmt.Println(os.Getwd())
	// fmt.Println(string(result))
	// fmt.Println(request)
	// fmt.Println(request.Body)

	var FacilityList []Facility
	//https://budougumi0617.github.io/2019/02/24/go-print-detail-of-json-syntax-error/
	if err := json.Unmarshal([]byte(request.Body), &FacilityList); err != nil {
		// if err, ok := err.(*json.SyntaxError); ok {
		// 	fmt.Println(string([]byte(request.Body)[err.Offset-15 : err.Offset+15]))
		// }
		fmt.Println(err)
		log.Fatal(err)
	}

	FacilitiyInfoMap := make(map[string]Facility)
	for _, el := range FacilityList {

		//入院・外来・救急の重複した医療機関情報を削除し、共通情報をリストで保持する
		if _, ok := FacilitiyInfoMap[el.FacilityId]; !ok {
			FacilitiyInfoMap[el.FacilityId] = el
		}

		//医療機関情報に入院・外来・救急の各回答状況を追加する
		if _, ok := FacilitiyInfoMap[el.FacilityId]; ok {
			val := FacilitiyInfoMap[el.FacilityId]
			if el.FacilityType == "入院" && FacilitiyInfoMap[el.FacilityId].FacilityId == el.FacilityId {
				val.Hospitalization = el.AnsType
				FacilitiyInfoMap[el.FacilityId] = val
			} else if el.FacilityType == "外来" && FacilitiyInfoMap[el.FacilityId].FacilityId == el.FacilityId {
				val.Outpatient = el.AnsType
				FacilitiyInfoMap[el.FacilityId] = val
			} else if el.FacilityType == "救急" && FacilitiyInfoMap[el.FacilityId].FacilityId == el.FacilityId {
				val.Emergency = el.AnsType
				FacilitiyInfoMap[el.FacilityId] = val
			} else {
				log.Fatalf("no mapping facility type")
			}
		}
	}

	db, err := sql.Open("mysql", "")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	insert, err := db.Prepare(fmt.Sprintf("INSERT INTO %s (facility_id, facility_name, zipcode, pref_name, facility_addr, facility_tel, latitude, longitude, submit_date, local_gov_code, city_name, facility_code, hospitalization, outpatient, emergency) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", "facility"))
	if err != nil {
		log.Fatal(err)
	}
	defer insert.Close()

	for _, val := range FacilitiyInfoMap {
		_, err = insert.Exec(val.FacilityId, val.FacilityName, val.ZipCode, val.PrefName, val.FacilityAddr, val.FacilityTel, val.Latitude, val.Longitude, val.SubmitDate, val.LocalGovCode, val.CityName, val.FacilityCode, val.Hospitalization, val.Outpatient, val.Emergency)
		if err != nil {
			log.Fatal(err)
		}
	}

	// fmt.Printf("FacilitiyInfo %#v\n", FacilitiyInfoMap)
	return events.APIGatewayProxyResponse{
		Body:       "test",
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
