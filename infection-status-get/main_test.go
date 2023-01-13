package main

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	_ "github.com/go-sql-driver/mysql"
)

func Test_createWhereClause(t *testing.T) {
	type args struct {
		req events.APIGatewayProxyRequest
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   []interface{}
		wantErr bool
	}{
		{
			name: "no query",
			args: args{
				events.APIGatewayProxyRequest{
					MultiValueQueryStringParameters: map[string][]string{},
				},
			},
			want:    "",
			want1:   make([]interface{}, 0),
			wantErr: false,
		},
		{
			name: "date: 1query",
			args: args{
				events.APIGatewayProxyRequest{
					MultiValueQueryStringParameters: map[string][]string{"date": {"20230101"}},
				},
			},
			want:    " WHERE (date = ?)",
			want1:   []interface{}{"20230101"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := createWhereClause(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("createWhereClause() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("createWhereClause() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("createWhereClause() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_openDB(t *testing.T) {
	tests := []struct {
		name    string
		want    *sql.DB
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := openDB()
			if (err != nil) != tt.wantErr {
				t.Errorf("openDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("openDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getStatusDaily(t *testing.T) {
	type args struct {
		req events.APIGatewayProxyRequest
	}
	tests := []struct {
		name    string
		args    args
		want    events.APIGatewayProxyResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getStatusDaily(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getStatusDaily() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getStatusDaily() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getStatusCumulatively(t *testing.T) {
	type args struct {
		in0 events.APIGatewayProxyRequest
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getStatusCumulatively(tt.args.in0)
		})
	}
}

func Test_getStatusComparison(t *testing.T) {
	type args struct {
		in0 events.APIGatewayProxyRequest
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getStatusComparison(tt.args.in0)
		})
	}
}

func Test_handler(t *testing.T) {
	type args struct {
		req events.APIGatewayProxyRequest
	}
	tests := []struct {
		name    string
		args    args
		want    events.APIGatewayProxyResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handler(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("handler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}
