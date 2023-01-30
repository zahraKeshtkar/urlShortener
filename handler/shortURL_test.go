package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"url-shortner/handler"
	"url-shortner/log"

	"url-shortner/model"
	"url-shortner/repository"
)

type URLSuite struct {
	suite.Suite
	mock       sqlmock.Sqlmock
	engine     *echo.Echo
	repository *repository.Link
	db         *gorm.DB
}

func (suite *URLSuite) SetupSuite() {
	var (
		db  *sql.DB
		err error
	)

	db, suite.mock, err = sqlmock.New()
	if err != nil {
		log.Errorf("Failed to open mock sql db, got error: %v", err)
	}

	if db == nil {
		log.Errorf("mock db is null")
	}

	if suite.mock == nil {
		log.Errorf("sqlmock is null")
	}
	log.InitLogger()
	log.SetOutput(os.Stdout)
	log.SetFormat(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel("ddd")

	suite.db, err = gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})

	suite.engine = echo.New()
	linkStore := &repository.Link{
		DB: suite.db,
	}
	redis, _ := redismock.NewClientMock()

	suite.engine.POST("/new", handler.SaveURL(linkStore, redis))
	suite.engine.GET("/:shortURL", handler.Redirect(linkStore, redis))
}

func (suite *URLSuite) TestSuccessRequest() {
	require := suite.Require()

	b, err := json.Marshal(model.Link{URL: "https://github.com/go-redis/redismock"})
	require.NoError(err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/new", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("^INSERT *").WithArgs("https://github.com/DATA-DOG/go-sqlmock").WillReturnResult(sqlmock.NewResult(1, 1))

	suite.mock.ExpectCommit()
	suite.engine.ServeHTTP(w, req)
	require.Equal(http.StatusOK, w.Code)
}

// func (suite *URLSuite) TestBody() {
// 	require := suite.Require()
//
// 	b, err := json.Marshal(9087)
// 	require.NoError(err)
//
// 	w := httptest.NewRecorder()
// 	req := httptest.NewRequest(http.MethodPost, "/new", bytes.NewReader(b))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
//
// 	suite.engine.ServeHTTP(w, req)
// 	require.Equal(http.StatusBadRequest, w.Code)
// }
//
// func (suite *URLSuite) TestBadRequest() {
// 	require := suite.Require()
//
// 	b, err := json.Marshal(model.Link{URL: "invalid url"})
// 	require.NoError(err)
//
// 	w := httptest.NewRecorder()
// 	req := httptest.NewRequest(http.MethodPost, "/new", bytes.NewReader(b))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	suite.engine.ServeHTTP(w, req)
// 	require.Equal(http.StatusBadRequest, w.Code)
//
// 	b, err = json.Marshal(model.Link{URL: "https://uibakery.io/regex-library/url"})
// 	require.NoError(err)
//
// 	w = httptest.NewRecorder()
// 	req = httptest.NewRequest(http.MethodPost, "/new", bytes.NewReader(b))
// 	suite.engine.ServeHTTP(w, req)
// 	require.Equal(http.StatusBadRequest, w.Code)
// }

// func (suite *URLSuite) TestRedirectSuccess() {
// 	require := suite.Require()
// 	w := httptest.NewRecorder()
// 	req := httptest.NewRequest(http.MethodGet, "/ZZZZZZZb", nil)
//
// 	suite.mock.ExpectBegin()
// 	suite.mock.ExpectQuery(regexp.QuoteMeta(
// 		`SELECT * FROM "links" WHERE (id = $1)`)).
// 		WithArgs(2).
// 		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}).
// 			AddRow(2, "url"))
// 	suite.mock.ExpectCommit()
// 	suite.engine.ServeHTTP(w, req)
// 	require.Equal(http.StatusFound, w.Code)
// }

// func (suite *URLSuite) TestRedirectBadRequest() {
// 	require := suite.Require()
// 	w := httptest.NewRecorder()
// 	req := httptest.NewRequest(http.MethodGet, "/b", nil)
// 	suite.engine.ServeHTTP(w, req)
// 	require.Equal(http.StatusBadRequest, w.Code)
//
// 	req = httptest.NewRequest(http.MethodGet, "/ZZZZZZZZ", nil)
// 	suite.engine.ServeHTTP(w, req)
// 	require.Equal(http.StatusBadRequest, w.Code)
// }

func TestURLSuite(t *testing.T) {
	suite.Run(t, new(URLSuite))
}
