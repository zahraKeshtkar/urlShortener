package model_test

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"url-shortner/log"
	"url-shortner/model"
)

type URLSuite struct {
	suite.Suite
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

func (suite *URLSuite) Test_URLValidate_Fail() {
	require := suite.Require()
	testCases := [3]model.Link{{ShortURL: ""}, {ShortURL: "ww.google"}, {ShortURL: "wwe.google.com"}}

	for _, link := range testCases {
		require.Equal(false, link.URLValidate())
	}
}

func (suite *URLSuite) Test_URLValidate_Success() {
	require := suite.Require()
	testCases := [2]model.Link{{URL: "www.google.com"}, {URL: "https://linuxhint.com/"}}

	for _, link := range testCases {
		require.Equal(true, link.URLValidate())
	}
}

func (suite *URLSuite) Test_ShortURLValidate_Fail() {
	require := suite.Require()
	testCases := [3]model.Link{{ShortURL: ""}, {ShortURL: "ZZZZZZZZZv"}, {ShortURL: "gg"}}

	for _, link := range testCases {
		require.Equal(false, link.ShortURLValidate())
	}
}

func (suite *URLSuite) Test_ShortURLValidate_Success() {
	require := suite.Require()
	shortURL := model.Link{ShortURL: "ZZZZZZZv"}

	require.Equal(true, shortURL.ShortURLValidate())
}

func (suite *URLSuite) Test_MakeShortURL_Fail() {
	require := suite.Require()
	testCases := [2]model.Link{{ID: 0}, {ID: -1}}

	for _, link := range testCases {
		err := link.MakeShortURL()
		require.Error(err)
	}
}

func (suite *URLSuite) Test_MakeShortURL_Success() {
	require := suite.Require()
	testCases := map[model.Link]string{model.Link{ID: 3}: "ZZZZZZZd", model.Link{ID: 27}: "ZZZZZZZB"}

	for link, shortURL := range testCases {
		err := link.MakeShortURL()

		require.NoError(err)
		require.Equal(shortURL, link.ShortURL)
	}
}

func (suite *URLSuite) Test_ShortURLToID_Fail() {
	require := suite.Require()
	testCases := [3]model.Link{{ShortURL: "ZZZZZZZZSS"}, {ShortURL: "hZ"}, {ShortURL: "ZZZZZZZZ"}}

	for _, link := range testCases {
		_, err := link.ShortURLToID()
		require.Error(err)
	}
}

func (suite *URLSuite) Test_ShortURLToID_Success() {
	require := suite.Require()
	testCases := map[model.Link]int{model.Link{ShortURL: "ZZZZZZZd"}: 3, model.Link{ShortURL: "ZZZZZZZB"}: 27}

	for link, expectedID := range testCases {
		id, err := link.ShortURLToID()

		require.NoError(err)
		require.Equal(expectedID, id)
	}
}

func TestURLSuite(t *testing.T) {
	suite.Run(t, new(URLSuite))
}
