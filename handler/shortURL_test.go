package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"url-shortner/handler"
	"url-shortner/log"

	"url-shortner/model"
	"url-shortner/repository"
)

type Response struct {
	message string
	code    int
}

type URLSuite struct {
	suite.Suite
	mock      sqlmock.Sqlmock
	db        *gorm.DB
	linkStore *repository.Link
	redisDB   *redis.Client
	redisMock redismock.ClientMock
}

func (suite *URLSuite) SetupSuite() {
	log.InitLogger()
	log.SetOutput(os.Stdout)
	log.SetFormat(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel("debug")
}

func (suite *URLSuite) Test_SaveURL_Fail() {
	require := suite.Require()
	testCases := map[interface{}]Response{
		model.Link{URL: "twitter"}:    {code: http.StatusBadRequest, message: "This is not a url at all"},
		model.Link{URL: "ww.twitter"}: {code: http.StatusBadRequest, message: "This is not a url at all"},
		model.Link{URL: ""}:           {code: http.StatusBadRequest, message: "This is not a url at all"},
		"https://echo.labstack.com":   {code: http.StatusBadRequest, message: "can not decode the body as json"},
		model.Link{URL: "https://echo.lb.com"}: {code: http.StatusInternalServerError,
			message: "can not insert to the database"},
		model.Link{URL: "https://echolb.com"}: {code: http.StatusInternalServerError,
			message: "some error in database occur"},
	}

	for link, response := range testCases {
		if response.message == "some error in database occur" {
			suite.mock.ExpectBegin()
			suite.mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `links` (`url`) VALUES (?)")).
				WithArgs(link.(model.Link).URL).
				WillReturnResult(sqlmock.NewResult(-2, 1))
			suite.mock.ExpectCommit()
		}

		body, err := json.Marshal(link)
		reader := bytes.NewReader(body)
		request := httptest.NewRequest(http.MethodPost, "/new", reader)
		context, _ := newEchoContext(request)
		httpResponse := handler.SaveURL(suite.linkStore, suite.redisDB)(context)
		httpError, ok := httpResponse.(*echo.HTTPError)

		require.NoError(err)
		require.Equal(true, ok)
		require.Error(httpError)
		require.Equal(httpError.Code, response.code)
		require.Equal(httpError.Message, response.message)
	}
}

func (suite *URLSuite) Test_SaveURL_Success() {
	require := suite.Require()
	url := "https://github.com/"
	shortURL := "ZZZZZZZc"
	var link *model.Link

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `links` (`url`) VALUES (?)")).
		WithArgs(url).
		WillReturnResult(sqlmock.NewResult(2, 1))
	suite.mock.ExpectCommit()
	suite.redisMock.ExpectHSet(shortURL, url, 1*time.Hour).SetErr(errors.New("test"))
	body, _ := json.Marshal(model.Link{URL: url})
	reader := bytes.NewReader(body)
	request := httptest.NewRequest(http.MethodPost, "/new", reader)
	context, response := newEchoContext(request)
	httpError := handler.SaveURL(suite.linkStore, suite.redisDB)(context)
	err := json.Unmarshal(response.Body.Bytes(), &link)

	require.NoError(httpError)
	require.Equal(http.StatusOK, response.Code)
	require.NoError(err)
	require.Equal(shortURL, link.ShortURL)

}

func (suite *URLSuite) Test_Redirect_Fail() {
	require := suite.Require()
	testCases := map[string]Response{
		"z":        {code: http.StatusBadRequest, message: "the short url is not valid"},
		"":         {code: http.StatusBadRequest, message: "the short url is not valid"},
		"ZZZZZZZZ": {code: http.StatusNotFound, message: "the short url is not found"}}

	for shortURL, response := range testCases {
		request := httptest.NewRequest(http.MethodGet, "/{shortURL}"+shortURL, nil)
		context, _ := newEchoContext(request)
		context.SetParamNames("shortURL")
		context.SetParamValues(shortURL)
		httpResponse := handler.Redirect(suite.linkStore, suite.redisDB)(context)
		httpError, ok := httpResponse.(*echo.HTTPError)

		require.Equal(true, ok)
		require.Error(httpError)
		require.Equal(response.code, httpError.Code)
		require.Equal(httpError.Message, response.message)
	}
}

func (suite *URLSuite) Test_Redirect_Success() {
	require := suite.Require()
	shortURL := "ZZZZZZZc"
	rows := suite.mock.NewRows([]string{"id", "url"}).AddRow(2, "jnkjhkjh")

	for i := 0; i < 2; i++ {
		if i == 0 {
			suite.redisMock.ExpectGet(shortURL).SetVal("https://github.com/")
		} else {
			suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `links` WHERE id = ? LIMIT 1")).
				WithArgs(2).WillReturnRows(rows)
			suite.redisMock.ExpectGet(shortURL).RedisNil()
		}

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		context, response := newEchoContext(request)
		context.SetParamNames("shortURL")
		context.SetParamValues(shortURL)
		httpError := handler.Redirect(suite.linkStore, suite.redisDB)(context)

		require.NoError(httpError)
		require.Equal(http.StatusFound, response.Code)
	}
}

func newEchoContext(request *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	response := httptest.NewRecorder()
	e := echo.New()

	return e.NewContext(request, response), response
}

func TestURLSuite(t *testing.T) {
	suite.Run(t, new(URLSuite))
}

func (suite *URLSuite) BeforeTest(_, _ string) {
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
	suite.db, err = gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})

	suite.linkStore = &repository.Link{
		DB: suite.db,
	}
	suite.redisDB, suite.redisMock = redismock.NewClientMock()
	if suite.redisDB == nil {
		log.Errorf("mock redis is null")
	}

	if suite.redisMock == nil {
		log.Errorf("redisMock  is null")
	}
}
