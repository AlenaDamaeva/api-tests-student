package addAdvertisement

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"

	"api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/auth/models"

	base "api-tests-template/tests"
)

type TestSuite struct {
	suite.Suite
	loginData models.LoginOkResponse
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func (s *TestSuite) SetupSuite() {
	base.SetupSuite()
	s.loginData = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
}

func (s *TestSuite) TearDownSuite() {
	base.TearDownSuite()
}

func (s *TestSuite) TestCreateAdvertisementPositive() {
	title := "Тестовое объявление"
	description := "Описание тестового объявления"
	price := 1000
	quantity := 1
	photoPath := "test_photo.jpg"

	var createResponse string
	var adID string

	s.Run("Создаем объявление со всеми полями и фото", func() {
		createResponse = advertisements.CreateAdvertisement(
			s.T(), s.loginData.Token,
			title, description, price, quantity, photoPath,
			http.StatusCreated,
		)
		adID = gjson.Get(createResponse, "id").String()
		require.NotEmpty(s.T(), adID, "ID объявления пустой")
	})

	s.Run("Проверяем, что объявление доступно по ID и возвращает те же параметры", func() {
		getResponse := advertisements.GetAdvertisementByID(s.T(), s.loginData.Token, adID, http.StatusOK)
		require.Equal(s.T(), title, gjson.Get(getResponse, "title").String(), "Title не совпадает")
		require.Equal(s.T(), description, gjson.Get(getResponse, "description").String(), "Description не совпадает")
		require.Equal(s.T(), price, int(gjson.Get(getResponse, "price").Int()), "Price не совпадает")
	})

	s.Run("Проверяем, что объявление появляется в поиске по полному названию", func() {
		searchResponse := advertisements.SearchAdvertisements(s.T(), s.loginData.Token, title, http.StatusOK)
		require.Contains(s.T(), searchResponse, adID, "Объявление не найдено в поиске")
	})

	s.Run("Проверяем, что фотография доступна по ручке /advertisements/{id}/photos", func() {
		photoResponse := advertisements.GetAdvertisementPhotos(s.T(), s.loginData.Token, adID, http.StatusOK)
		require.NotEmpty(s.T(), photoResponse, "Ответ с фото пустой")
		require.Contains(s.T(), photoResponse, "url", "В ответе нет ссылки на фото")
	})

	s.T().Logf("Все проверки для объявления %s пройдены!", adID)
}
func (s *TestSuite) TestCreateAdvertisementWithoutTitle() {
	advertisements.CreateAdvertisement(
		s.T(), s.loginData.Token,
		"", "Описание", 1000, 1, "test_photo.jpg",
		http.StatusBadRequest,
	)
}

func (s *TestSuite) TestCreateAdvertisementWithNegativePrice() {
	advertisements.CreateAdvertisement(
		s.T(), s.loginData.Token,
		"Тестовое объявление", "Описание", -100, 1, "test_photo.jpg",
		http.StatusBadRequest,
	)
}

func (s *TestSuite) TestCreateAdvertisementWithoutToken() {
	advertisements.CreateAdvertisement(
		s.T(), "",
		"Тестовое объявление", "Описание", 1000, 1, "test_photo.jpg",
		http.StatusUnauthorized,
	)
}
