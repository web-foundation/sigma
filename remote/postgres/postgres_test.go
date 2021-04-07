package postgres_test

import (
	"database/sql"
	"database/sql/driver"
	"log"
	"reflect"
	"sigma-production/remote/postgres"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"sigma-production/interpreter"
)

func TestPostgres_AddModel(t *testing.T) {
	db, mock := newDBMock()
	defer db.Close()
	pg := postgres.New(db)

	primaryType := interpreter.Model{
		Name: "User",
		Fields: []interpreter.Field{
			{
				Name:     "username",
				Type:     "String",
				Nullable: true,
			},
			{
				Name:     "email",
				Type:     "String",
				Nullable: false,
			},
			{
				Name:     "password",
				Type:     "String",
				Nullable: false,
			},
			{
				Name:     "settings",
				Type:     "Settings",
				Nullable: false,
			},
		},
	}
	relationType := interpreter.Model{
		Name: "Settings",
		Fields: []interpreter.Field{
			{
				Name:     "theme",
				Type:     "String",
				Nullable: false,
			},
			{
				Name:     "subUser",
				Type:     "User",
				Nullable: false,
			},
		},
	}

	dne1 := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	dne2 := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	dne3 := sqlmock.NewRows([]string{"exists"}).AddRow(true)

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(dne1)
	mock.ExpectExec("CREATE TABLE").WillReturnResult(driver.ResultNoRows)
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(dne2)
	mock.ExpectExec("CREATE TABLE").WillReturnResult(driver.ResultNoRows)
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(dne3)
	mock.ExpectExec("ALTER TABLE").WillReturnResult(driver.ResultNoRows)
	mock.ExpectExec("ALTER TABLE").WillReturnResult(driver.ResultNoRows)
	mock.ExpectExec("ALTER TABLE").WillReturnResult(driver.ResultNoRows)
	mock.ExpectExec("ALTER TABLE").WillReturnResult(driver.ResultNoRows)

	res, err := pg.AddModel(primaryType, []interpreter.Model{primaryType, relationType})
	if err != nil {
		t.Fatalf("Error ::: %s", err.Error())
	}

	expected := []string{
		`CREATE TABLE "User" (id SERIAL PRIMARY KEY, username TEXT, email TEXT NOT NULL, password TEXT NOT NULL)`,
		`CREATE TABLE "Settings" (id SERIAL PRIMARY KEY, theme TEXT NOT NULL)`,
		`ALTER TABLE "Settings" ADD COLUMN "subUserId" INT NOT NULL`,
		`ALTER TABLE "Settings" ADD CONSTRAINT "fk_subUser_id" FOREIGN KEY (id) REFERENCES "User" (id) ON UPDATE CASCADE ON DELETE CASCADE`,
		`ALTER TABLE "User" ADD COLUMN "settingsId" INT NOT NULL`,
		`ALTER TABLE "User" ADD CONSTRAINT "fk_settings_id" FOREIGN KEY (id) REFERENCES "Settings" (id) ON UPDATE CASCADE ON DELETE CASCADE`,
	}

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("\ngot\n--- %s \n\nexpected\n--- %s", fmtSQLStmts(res), fmtSQLStmts(expected))
	}
}

func fmtSQLStmts(stmts []string) string {
	newFStmt := "\n" + (stmts[0])
	stmts[0] = newFStmt
	return strings.Join(stmts, "\n")
}

func newDBMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return db, mock
}