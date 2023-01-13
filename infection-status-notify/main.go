package main

import (
	"bytes"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"sort"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/freetype/truetype"
	"github.com/slack-go/slack"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/wcharczuk/go-chart/v2"
)

type InfectionStatus struct {
	Date                        time.Time `json:"date"`
	Prefecture                  string    `json:"prefecture"`
	InfectionNumberDaily        int       `json:"infectionNumberDaily"`
	InfectionNumberCumulatively int       `json:"infectionNumberCumulatively"`
}

type Key struct {
	Date       time.Time `json:"date"`
	Prefecture string    `json:"prefecture"`
}

//go:embed Koruri-Bold.ttf
var fontBytes []byte

var regionDict = map[string][]string{
	"1. 北海道・東北地方": {"北海道", "青森県", "岩手県", "宮城県", "秋田県", "山形県", "福島県"},
	"2. 関東地方":     {"茨城県", "栃木県", "群馬県", "埼玉県", "千葉県", "東京都", "神奈川県"},
	"3. 中部地方":     {"新潟県", "富山県", "石川県", "福井県", "山梨県", "長野県", "岐阜県", "静岡県", "愛知県", "三重県"},
	"4. 近畿地方":     {"滋賀県", "京都府", "大阪府", "兵庫県", "奈良県", "和歌山県"},
	"5. 中国・四国地方":  {"鳥取県", "島根県", "岡山県", "広島県", "山口県", "徳島県", "香川県", "愛媛県", "高知県"},
	"6. 九州・沖縄地方":  {"福岡県", "佐賀県", "長崎県", "熊本県", "大分県", "宮崎県", "鹿児島県", "沖縄県"},
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

// チャート日付 昇順ソート
func compareByDate(s1, s2 *InfectionStatus) bool {
	return s1.Date.Before(s2.Date)
}

// 地方 昇順ソート
func compareByRegion(s1, s2 *chart.Chart) bool {
	return s1.Title < s2.Title
}

// 都道府県 昇順ソート
func compareByPrefecture(s1, s2 *chart.TimeSeries) bool {
	return s1.Name < s2.Name
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//font
	face, err := truetype.Parse(fontBytes)
	if err != nil {
		panic(err)
	}

	//x UTC
	x := []time.Time{
		time.Now().AddDate(0, 0, -8),
		time.Now().AddDate(0, 0, -7),
		time.Now().AddDate(0, 0, -6),
		time.Now().AddDate(0, 0, -5),
		time.Now().AddDate(0, 0, -4),
		time.Now().AddDate(0, 0, -3),
		time.Now().AddDate(0, 0, -2),
		time.Now().AddDate(0, 0, -1),
	}

	//y
	db, err := openDB()
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	from := time.Now().AddDate(0, 0, -8).Format("2006-01-02")
	to := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	q := fmt.Sprintf("SELECT date, prefecture, infection_number_daily FROM infection_status WHERE date >= '%s' AND date <= '%s'", from, to)
	rows, err := db.Query(q)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	//DB取得データのMap化
	infectionStatusMap := make(map[string][]InfectionStatus)
	for rows.Next() {
		var infectionStatus InfectionStatus
		if err := rows.Scan(
			&infectionStatus.Date,
			&infectionStatus.Prefecture,
			&infectionStatus.InfectionNumberDaily,
		); err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		if val, ok := infectionStatusMap[infectionStatus.Prefecture]; ok {
			val = append(val, infectionStatus)
			infectionStatusMap[infectionStatus.Prefecture] = val
		} else {
			infectionStatusList := []InfectionStatus{infectionStatus}
			infectionStatusMap[infectionStatus.Prefecture] = infectionStatusList
		}
	}

	//都道府県チャートの作成
	var prefectureChartList []chart.TimeSeries
	for prefectureName, infectionStatusList := range infectionStatusMap {
		prefectureChart := chart.TimeSeries{
			Name:    prefectureName,
			XValues: x,
		}

		//y軸の日付順をソート・y軸の作成
		sort.Slice(infectionStatusList, func(i, j int) bool { return compareByDate(&infectionStatusList[i], &infectionStatusList[j]) })
		for _, infectionStatus := range infectionStatusList {
			prefectureChart.YValues = append(prefectureChart.YValues, float64(infectionStatus.InfectionNumberDaily))
		}
		prefectureChartList = append(prefectureChartList, prefectureChart)
	}
	sort.Slice(prefectureChartList, func(i, j int) bool { return compareByPrefecture(&prefectureChartList[i], &prefectureChartList[j]) })

	//地方チャートの作成
	regionChartList := make([]chart.Chart, 0)
	for regionName, prefectures := range regionDict {
		regionChart := chart.Chart{
			Title: regionName,
			Font:  face,
			Background: chart.Style{
				Padding: chart.Box{
					Top:  20,
					Left: 260,
				},
			},
		}
		//都道府県チャートの挿入
		for _, prefecture := range prefectures {
			for _, prefectureChart := range prefectureChartList {
				if prefecture == prefectureChart.Name {
					regionChart.Series = append(regionChart.Series, prefectureChart)
				}
			}
		}
		regionChartList = append(regionChartList, regionChart)
	}
	sort.Slice(regionChartList, func(i, j int) bool { return compareByRegion(&regionChartList[i], &regionChartList[j]) })

	for _, graph := range regionChartList {
		graph.Elements = []chart.Renderable{
			chart.LegendLeft(&graph),
		}

		buf := bytes.NewBuffer([]byte{})
		graph.Render(chart.PNG, buf)

		//画像送信
		token := os.Getenv("TOKEN")
		api := slack.New(token)

		_, err = api.UploadFile(
			slack.FileUploadParameters{
				Reader:   buf,
				Filename: graph.Title + ".png",
				Channels: []string{"go-academy"},
			})
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
	}

	// //テキスト送信
	// data := []byte("{'text': 'test'}")
	// body := bytes.NewReader(data)
	// res, err := http.Post(os.Getenv("WEBHOOK"), "application/json", body)
	// if err != nil {
	// 	return events.APIGatewayProxyResponse{}, err
	// }
	// defer res.Body.Close()

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "ok",
	}, nil
}

func main() {
	lambda.Start(handler)
}
