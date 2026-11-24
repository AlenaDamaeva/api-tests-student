package addAdvertisement

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"

	"api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/auth/models"

	base "api-tests-template/tests"
)

const (
	testDescription = "Описание тестового объявления"
	testPrice       = 1000
	testQuantity    = 1

	testPhotoPath    = "test_photo.jpg"
	invalidPhotoPath = "test_photo.txt"
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

	s.loginData = auth.Login(
		s.T(),
		os.Getenv("TEST_LOGIN"),
		os.Getenv("TEST_PASSWORD"),
	)
}

func (s *TestSuite) TearDownSuite() {
	base.TearDownSuite()
}

// Проверяет, что сервер вернул описание ошибки с полями error и message.
// Если передан expectedSubstr, дополнительно проверяет его наличие в message.
func checkErrorResponse(t *testing.T, body string, expectedSubstr string) {
	require.NotEmpty(t, body, "Тело ответа с ошибкой пустое")

	errorText := gjson.Get(body, "error").String()
	messageText := gjson.Get(body, "message").String()

	require.NotEmpty(t, errorText, "В ответе отсутствует поле error (тело: %s)", body)
	require.NotEmpty(t, messageText, "В ответе отсутствует поле message (тело: %s)", body)

	if expectedSubstr != "" {
		require.Contains(
			t,
			strings.ToLower(messageText),
			strings.ToLower(expectedSubstr),
			"Поле message %q должно содержать %q (полный ответ: %s)",
			messageText,
			expectedSubstr,
			body,
		)
	}
}

// Позитивный сценарий: создание объявления, проверка чтения по ID, поиска и фото,
// а также удаление через DELETE в Cleanup.
func (s *TestSuite) TestCreateAdvertisementPositive() {
	title := fmt.Sprintf("Тестовое объявление %d", time.Now().UnixNano())

	var adID string

	s.Run("Создаем объявление со всеми полями и фотографией", func() {
		createResponse := advertisements.CreateAdvertisement(
			s.T(),
			s.loginData.Token,
			title,
			testDescription,
			testPrice,
			testQuantity,
			testPhotoPath,
			http.StatusCreated,
		)

		adID = gjson.Get(createResponse, "id").String()
		require.NotEmpty(s.T(), adID, "ID объявления пустой")
		require.Equal(s.T(), title, gjson.Get(createResponse, "title").String())
		require.Equal(s.T(), testDescription, gjson.Get(createResponse, "description").String())
		require.Equal(s.T(), int64(testPrice), gjson.Get(createResponse, "price").Int())
		require.Equal(s.T(), int64(testQuantity), gjson.Get(createResponse, "quantity").Int())
	})

	// Гарантированно удаляем созданное объявление после завершения проверок.
	s.T().Cleanup(func() {
		if adID != "" {
			advertisements.DeleteAdvertisement(
				s.T(),
				s.loginData.Token,
				adID,
				http.StatusNoContent,
			)
		}
	})

	s.Run("Получаем объявление по ID и проверяем, что данные сохранились корректно", func() {
		require.NotEmpty(s.T(), adID, "ID объявления должен быть получен на предыдущем шаге")

		getResponse := advertisements.GetAdvertisementByID(
			s.T(),
			s.loginData.Token,
			adID,
			http.StatusOK,
		)

		require.Equal(s.T(), adID, gjson.Get(getResponse, "id").String())
		require.Equal(s.T(), title, gjson.Get(getResponse, "title").String())
		require.Equal(s.T(), testDescription, gjson.Get(getResponse, "description").String())
		require.Equal(s.T(), int64(testPrice), gjson.Get(getResponse, "price").Int())
		require.Equal(s.T(), int64(testQuantity), gjson.Get(getResponse, "quantity").Int())
	})

	s.Run("Ищем объявление по полному названию и проверяем наличие в выдаче", func() {
		require.NotEmpty(s.T(), adID)

		searchResponse := advertisements.SearchAdvertisements(
			s.T(),
			s.loginData.Token,
			title,
			http.StatusOK,
		)

		items := gjson.Get(searchResponse, "items").Array()
		require.NotEmpty(s.T(), items, "Список items в поиске не должен быть пустым")

		found := false
		for _, item := range items {
			if item.Get("id").String() == adID {
				found = true
				require.Equal(s.T(), title, item.Get("title").String())
				break
			}
		}
		require.True(s.T(), found, "Созданное объявление с ID %s не найдено в items", adID)
	})

	s.Run("Проверяем доступность фотографий объявления", func() {
		require.NotEmpty(s.T(), adID)

		photoResponse := advertisements.GetAdvertisementPhotos(
			s.T(),
			s.loginData.Token,
			adID,
			http.StatusOK,
		)

		photos := gjson.Parse(photoResponse).Array()
		require.NotEmpty(s.T(), photos, "Массив фотографий пустой: %s", photoResponse)

		firstURL := photos[0].Get("url").String()
		require.NotEmpty(s.T(), firstURL, "У фотографии отсутствует url: %s", photoResponse)
		require.True(
			s.T(),
			strings.HasPrefix(firstURL, "http://") || strings.HasPrefix(firstURL, "https://"),
			"Некорректный URL фотографии: %s",
			firstURL,
		)
	})
}

// Негативный сценарий: пустой title.
func (s *TestSuite) TestCreateAdvertisementWithoutTitle() {
	body, status := advertisements.CreateAdvertisementRaw(
		s.T(),
		s.loginData.Token,
		"",
		testDescription,
		"1000",
		"1",
		[]string{testPhotoPath},
	)

	require.Equal(s.T(), http.StatusBadRequest, status)
	checkErrorResponse(s.T(), body, "title")
}

// Негативный сценарий: title длиннее 255 символов.
func (s *TestSuite) TestCreateAdvertisementWithLongTitle() {
	longTitle := strings.Repeat("A", 256)

	body, status := advertisements.CreateAdvertisementRaw(
		s.T(),
		s.loginData.Token,
		longTitle,
		testDescription,
		"1000",
		"1",
		[]string{testPhotoPath},
	)

	require.Equal(s.T(), http.StatusBadRequest, status)
	checkErrorResponse(s.T(), body, "title")
}

// Негативный сценарий: отрицательная цена.
func (s *TestSuite) TestCreateAdvertisementWithNegativePrice() {
	body, status := advertisements.CreateAdvertisementRaw(
		s.T(),
		s.loginData.Token,
		"Тестовое объявление",
		testDescription,
		"-100",
		"1",
		[]string{testPhotoPath},
	)

	require.Equal(s.T(), http.StatusBadRequest, status)
	checkErrorResponse(s.T(), body, "price")
}

// Негативный сценарий: отрицательное количество.
func (s *TestSuite) TestCreateAdvertisementWithNegativeQuantity() {
	body, status := advertisements.CreateAdvertisementRaw(
		s.T(),
		s.loginData.Token,
		"Тестовое объявление",
		testDescription,
		"1000",
		"-1",
		[]string{testPhotoPath},
	)

	require.Equal(s.T(), http.StatusBadRequest, status)
	checkErrorResponse(s.T(), body, "quantity")
}

// Негативный сценарий: цена передана строкой (не числом).
func (s *TestSuite) TestCreateAdvertisementWithStringPrice() {
	body, status := advertisements.CreateAdvertisementRaw(
		s.T(),
		s.loginData.Token,
		"Тестовое объявление",
		testDescription,
		"abc",
		"1",
		[]string{testPhotoPath},
	)

	require.Equal(s.T(), http.StatusBadRequest, status)
	checkErrorResponse(s.T(), body, "")
}

// Негативный сценарий: количество передано строкой (не числом).
func (s *TestSuite) TestCreateAdvertisementWithStringQuantity() {
	body, status := advertisements.CreateAdvertisementRaw(
		s.T(),
		s.loginData.Token,
		"Тестовое объявление",
		testDescription,
		"1000",
		"abc",
		[]string{testPhotoPath},
	)

	require.Equal(s.T(), http.StatusBadRequest, status)
	checkErrorResponse(s.T(), body, "")
}

// Негативный сценарий: количество передано не целым числом.
func (s *TestSuite) TestCreateAdvertisementWithFloatQuantity() {
	body, status := advertisements.CreateAdvertisementRaw(
		s.T(),
		s.loginData.Token,
		"Тестовое объявление",
		testDescription,
		"1000",
		"1.5",
		[]string{testPhotoPath},
	)

	require.Equal(s.T(), http.StatusBadRequest, status)
	checkErrorResponse(s.T(), body, "")
}

// Негативный сценарий: неподдерживаемый формат фотографии.
func (s *TestSuite) TestCreateAdvertisementWithInvalidPhotoFormat() {
	body, status := advertisements.CreateAdvertisementRaw(
		s.T(),
		s.loginData.Token,
		"Тестовое объявление",
		testDescription,
		"1000",
		"1",
		[]string{invalidPhotoPath},
	)

	require.Equal(s.T(), http.StatusBadRequest, status)
	checkErrorResponse(s.T(), body, "photo")
}

// Негативный сценарий: отправляем четыре фотографии вместо допустимых 1–3.
func (s *TestSuite) TestCreateAdvertisementWithTooManyPhotos() {
	photoPaths := []string{
		testPhotoPath,
		testPhotoPath,
		testPhotoPath,
		testPhotoPath,
	}

	body, status := advertisements.CreateAdvertisementRaw(
		s.T(),
		s.loginData.Token,
		"Тестовое объявление",
		testDescription,
		"1000",
		"1",
		photoPaths,
	)

	require.Equal(s.T(), http.StatusBadRequest, status)
	checkErrorResponse(s.T(), body, "photo")
}

// Негативный сценарий: создание объявления без токена авторизации.
func (s *TestSuite) TestCreateAdvertisementWithoutToken() {
	body, status := advertisements.CreateAdvertisementRaw(
		s.T(),
		"",
		"Тестовое объявление",
		testDescription,
		"1000",
		"1",
		[]string{testPhotoPath},
	)

	require.Equal(s.T(), http.StatusUnauthorized, status)
	checkErrorResponse(s.T(), body, "")
}

// Негативный сценарий: создание объявления с неверным токеном авторизации.
func (s *TestSuite) TestCreateAdvertisementWithInvalidToken() {
	body, status := advertisements.CreateAdvertisementRaw(
		s.T(),
		"invalid_token_12345",
		"Тестовое объявление",
		testDescription,
		"1000",
		"1",
		[]string{testPhotoPath},
	)

	require.Equal(s.T(), http.StatusUnauthorized, status)
	checkErrorResponse(s.T(), body, "")
}
