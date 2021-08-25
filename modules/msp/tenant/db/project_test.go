package db

import (
	"github.com/DATA-DOG/go-sqlmock"
	"reflect"
	"testing"

	"github.com/jinzhu/gorm"
)

func dbMockInit() (*MSPProjectDB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if nil != err {
		return nil, nil, nil
	}
	mydb, err := gorm.Open("mysql", db)
	mspProjectDB := &MSPProjectDB{mydb}
	return mspProjectDB, mock, err
}

func TestMSPProjectDB_Query(t *testing.T) {
	db, _, _ := sqlmock.New()
	mydb, _ := gorm.Open("mysql", db)
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *MSPProject
		wantErr bool
	}{
		{name: "case1", fields: fields{DB: mydb}, args: args{id: "1"}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &MSPProjectDB{
				DB: tt.fields.DB,
			}
			got, err := db.Query(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Query() got = %v, want %v", got, tt.want)
			}
		})
	}
}
