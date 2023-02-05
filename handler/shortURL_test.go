package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
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

	"url-shortner/config"
	"url-shortner/handler"
	"url-shortner/log"
	"url-shortner/model"
	"url-shortner/repository"
)

type URLTestSuite struct {
	suite.Suite
	mock      sqlmock.Sqlmock
	db        *gorm.DB
	linkStore *repository.Link
	redisDB   *redis.Client
	redisMock redismock.ClientMock
	buffer    bytes.Buffer
}

func (suite *URLTestSuite) SetupSuite() {
	log.InitLogger()
	suite.buffer = bytes.Buffer{}
	log.SetOutput(&suite.buffer)
	log.SetFormat(&logrus.TextFormatter{DisableTimestamp: true, DisableQuote: true})
	log.SetLevel("debug")
	_, err := config.Init()
	if err != nil {
		log.Errorf("Failed to read from config ,got err: %s", err)
	}
}

func (suite *URLTestSuite) SetupTest() {
	suite.buffer.Reset()
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

func (suite *URLTestSuite) TearDownTest() {
	sqlDB, err := suite.db.DB()
	if err != nil {
		log.Errorf("Failed to close mock sql db, got error: %v", err)
	}

	err = sqlDB.Close()
	if err != nil {
		log.Errorf("Failed to close mock sql db, got error: %v", err)
	}

	err = suite.redisDB.Close()
	if err != nil {
		log.Errorf("Failed to close mock sql db, got error: %v", err)
	}
}

func (suite *URLTestSuite) Test_SaveURL_Success() {
	require := suite.Require()
	url := "https://github.com/"
	expectedShortURL := "ZZZZZZZc"
	expectedLogMessage := fmt.Sprintf("level=info msg=insert in the redis successfully.the value is: [%s]\n", expectedShortURL)
	var link *model.Link

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `links` (`url`) VALUES (?)")).
		WithArgs(url).
		WillReturnResult(sqlmock.NewResult(2, 1))
	suite.mock.ExpectCommit()
	suite.redisMock.ExpectSet(expectedShortURL, url, 1*time.Hour).SetVal("")
	body, _ := json.Marshal(model.Link{URL: url})
	reader := bytes.NewReader(body)
	request := httptest.NewRequest(http.MethodPost, "/new", reader)
	context, response := newEchoContext(request)
	httpError := handler.SaveURL(suite.linkStore, suite.redisDB)(context)
	err := json.Unmarshal(response.Body.Bytes(), &link)

	require.NoError(httpError)
	require.NoError(err)
	require.Equal(http.StatusOK, response.Code)
	require.Equal(expectedShortURL, link.ShortURL)
	require.Equal(suite.buffer.String(), expectedLogMessage)
}

func (suite *URLTestSuite) Test_Redirect_Success() {
	require := suite.Require()
	shortURL := "ZZZZZZZc"
	rows := suite.mock.NewRows([]string{"id", "url"}).AddRow(2, "jnkjhkjh")

	for i := 0; i < 2; i++ {
		if i == 0 {
			suite.redisMock.ExpectGet(shortURL).SetVal("https://github.com/")
		} else {
			suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `links` WHERE id = ?")).
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

func (suite *URLTestSuite) Test_SaveURL_URLValidation_Fail() {
	require := suite.Require()
	testCases := map[model.Link]echo.HTTPError{
		model.Link{URL: "twitter"}:    {Code: http.StatusBadRequest, Message: "This is not a url at all"},
		model.Link{URL: "ww.twitter"}: {Code: http.StatusBadRequest, Message: "This is not a url at all"},
		model.Link{URL: ""}:           {Code: http.StatusBadRequest, Message: "This is not a url at all"},
		model.Link{URL: "https://echo.lb.com"}: {Code: http.StatusInternalServerError,
			Message: "can not insert to the database"},
		model.Link{URL: "https://echolb.com"}: {Code: http.StatusInternalServerError,
			Message: "A database error has occurred"},
	}

	for link, response := range testCases {
		if response.Message == "A database error has occurred" {
			suite.mock.ExpectBegin()
			suite.mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `links` (`url`) VALUES (?)")).
				WithArgs(link.URL).
				WillReturnResult(sqlmock.NewResult(-2, 1))
			suite.mock.ExpectCommit()
		}

		body, _ := json.Marshal(link)
		reader := bytes.NewReader(body)
		request := httptest.NewRequest(http.MethodPost, "/new", reader)
		context, _ := newEchoContext(request)
		httpResponse := handler.SaveURL(suite.linkStore, suite.redisDB)(context)
		httpError, ok := httpResponse.(*echo.HTTPError)

		require.Equal(true, ok)
		require.Error(httpError)
		require.Equal(httpError.Code, response.Code)
		require.Equal(httpError.Message, response.Message)
	}
}

func (suite *URLTestSuite) Test_SaveURL_RequestBody_Fail() {
	require := suite.Require()
	URL := "https://echo.labstack.com"
	expectedStatusCode := http.StatusBadRequest
	expectedMessage := "can not decode the body as json"

	body, _ := json.Marshal(URL)
	reader := bytes.NewReader(body)
	request := httptest.NewRequest(http.MethodPost, "/new", reader)
	context, _ := newEchoContext(request)
	httpResponse := handler.SaveURL(suite.linkStore, suite.redisDB)(context)
	httpError, ok := httpResponse.(*echo.HTTPError)

	require.Equal(true, ok)
	require.Error(httpError)
	require.Equal(httpError.Code, expectedStatusCode)
	require.Equal(httpError.Message, expectedMessage)
}

func (suite *URLTestSuite) Test_SaveURL_Redis_Fail() {
	require := suite.Require()
	url := "https://echo.labstack.com"
	shortURL := "ZZZZZZZb"
	err := errors.New("bomb")
	expectedLogMessage := fmt.Sprintf("level=error msg=can not insert in redis the err is : [%s]\n", err)

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `links` (`url`) VALUES (?)")).
		WithArgs(url).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.mock.ExpectCommit()
	suite.redisMock.ExpectSet(shortURL, url, 1*time.Hour).SetErr(err)
	body, _ := json.Marshal(model.Link{URL: url})
	reader := bytes.NewReader(body)
	request := httptest.NewRequest(http.MethodPost, "/new", reader)
	context, _ := newEchoContext(request)
	_ = handler.SaveURL(suite.linkStore, suite.redisDB)(context)

	require.Equal(suite.buffer.String(), expectedLogMessage)
}

func (suite *URLTestSuite) Test_SaveURL_Mysql_Fail() {
	require := suite.Require()
	url := "https://echo.labstack.com"
	expectedCode := http.StatusInternalServerError
	expectedMessage := "can not insert to the database"

	body, _ := json.Marshal(model.Link{URL: url})
	reader := bytes.NewReader(body)
	request := httptest.NewRequest(http.MethodPost, "/new", reader)
	context, _ := newEchoContext(request)
	httpResponse := handler.SaveURL(suite.linkStore, suite.redisDB)(context)
	httpError, ok := httpResponse.(*echo.HTTPError)

	require.Equal(true, ok)
	require.Error(httpError)
	require.Equal(httpError.Code, expectedCode)
	require.Equal(httpError.Message, expectedMessage)
}

func (suite *URLTestSuite) Test_Redirect_ShortURLValidation_Fail() {
	require := suite.Require()
	testCases := map[string]echo.HTTPError{
		"z":        {Code: http.StatusBadRequest, Message: "the short url is not valid"},
		"#ZZZZZZ#": {Code: http.StatusBadRequest, Message: "the short url is not valid"}}

	for shortURL, response := range testCases {
		request := httptest.NewRequest(http.MethodGet, "/{shortURL}"+shortURL, nil)
		context, _ := newEchoContext(request)
		context.SetParamNames("shortURL")
		context.SetParamValues(shortURL)
		httpResponse := handler.Redirect(suite.linkStore, suite.redisDB)(context)
		httpError, ok := httpResponse.(*echo.HTTPError)

		require.Equal(true, ok)
		require.Error(httpError)
		require.Equal(response.Code, httpError.Code)
		require.Equal(httpError.Message, response.Message)
	}
}

func (suite *URLTestSuite) Test_Redirect_Redis_Fail() {
	require := suite.Require()
	shortURL := "ZZZZZZZc"
	err := errors.New("bomb")
	expectedLogMessages := [2]string{"level=info msg=the short url is not in redis : [redis: nil]",
		"level=error msg=A redis error has occurred: [bomb]"}

	for i := 0; i < 2; i++ {
		if i == 0 {
			suite.redisMock.ExpectGet(shortURL).RedisNil()
		} else {
			suite.redisMock.ExpectGet(shortURL).SetErr(err)
		}

		request := httptest.NewRequest(http.MethodGet, "/{shortURL}"+shortURL, nil)
		context, _ := newEchoContext(request)
		context.SetParamNames("shortURL")
		context.SetParamValues(shortURL)
		_ = handler.Redirect(suite.linkStore, suite.redisDB)(context)

		require.Equal(strings.Split(suite.buffer.String(), "\n")[0], expectedLogMessages[i])
		suite.buffer.Reset()
	}
}

func newEchoContext(request *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	response := httptest.NewRecorder()
	e := echo.New()

	return e.NewContext(request, response), response
}

func TestURLTestSuite(t *testing.T) {
	suite.Run(t, new(URLTestSuite))
}
