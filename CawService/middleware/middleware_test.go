package middleware

import (
	"net/http"
	"net/http/httptest"
	"time"

	"Caw/UserService/utils"
	"testing"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
)

type WrappedHandler struct {
	WasCalled bool
}

func NewWrappedHander() *WrappedHandler {
	return &WrappedHandler{WasCalled: false}
}

func (wh *WrappedHandler) Handler(w http.ResponseWriter, r *http.Request) {
	wh.WasCalled = true
}

func TestMustAuth(t *testing.T) {
	validToken, err := utils.NewUserToken("anycmon", time.Now().Add(time.Duration(1*time.Minute)).Unix())
	if err != nil {
		t.Error(err)
	}
	invalidToken, err := utils.NewUserToken("anycmon", time.Now().Add(time.Duration(-1*time.Minute)).Unix())
	if err != nil {
		t.Error(err)
	}
	testCases := []struct {
		Name                                 string
		AuthorizationHeader                  string
		ExpectedIsProtectedHandlerCalled     bool
		ExpectedIsUserClaimsPresentInContext bool
		ExpectedStatusCode                   int
	}{
		{
			"SuccessfullyAuthenticationTest",
			"Bearer " + validToken,
			true,
			true,
			http.StatusOK,
		},
		{
			"ExpiredTokenTest",
			"Bearer " + invalidToken,
			false,
			false,
			http.StatusUnauthorized,
		},
		{
			"UnsupportedAuthorizationMethodTest",
			"UnknowMethod " + validToken,
			false,
			false,
			http.StatusUnauthorized,
		},
		{
			"InvalidTokenSignatureTest",
			"Bearer " + validToken[:len(validToken)-1],
			false,
			false,
			http.StatusUnauthorized,
		},
		{
			"MissingAuthorizationMethodTest",
			validToken,
			false,
			false,
			http.StatusUnauthorized,
		},
	}
	log := logrus.New()
	isProtectedHandlerCalled := false
	isUserClaimsPresentInContext := false

	mustAuth := MustAuth(log)(func(w http.ResponseWriter, r *http.Request) {
		isProtectedHandlerCalled = true
		if r.Context().Value("userClaims") != nil {
			isUserClaimsPresentInContext = true
		}
	})

	ts := httptest.NewServer(mustAuth)
	defer ts.Close()

	clinet := http.Client{}
	for _, testCase := range testCases {
		t.Log(testCase.Name)

		isProtectedHandlerCalled = false
		isUserClaimsPresentInContext = false

		req, err := http.NewRequest("GET", ts.URL+"/testUrl", nil)
		if err != nil {
			t.Error(err)
		}

		req.Header.Set("Authorization", testCase.AuthorizationHeader)
		res, err := clinet.Do(req)
		if err != nil {
			t.Error(err)
		}
		if res.StatusCode != testCase.ExpectedStatusCode {
			t.Errorf("Wrong status code: expected %v given %v", testCase.ExpectedStatusCode, res.StatusCode)
		}
		if isProtectedHandlerCalled != testCase.ExpectedIsProtectedHandlerCalled {
			t.Errorf("IsProtectedHandlerCalled is different than expected. Expected %v given %v ",
				testCase.ExpectedIsProtectedHandlerCalled, isProtectedHandlerCalled)
		}
		if isUserClaimsPresentInContext != testCase.ExpectedIsUserClaimsPresentInContext {
			t.Errorf("isUserClaimsPresentInContext is different than expected. Expected %v given %v ",
				testCase.ExpectedIsUserClaimsPresentInContext, isUserClaimsPresentInContext)
		}

	}
}

func TestProduce(t *testing.T) {
	supportedAccept := "application/json"
	testCases := []struct {
		Name                           string
		AcceptHeader                   string
		ExceptedIsWrappedHandlerCalled bool
		ExpectedStatusCode             int
	}{
		{
			"SupportedAcceptHeaderValueTest",
			supportedAccept,
			true,
			http.StatusOK,
		},
		{
			"UnsupportedAcceptHeaderValueTest",
			"application/xml",
			false,
			http.StatusNotAcceptable,
		},
		{
			"SupportedAcceptHeaderValueInListOfValuesTest",
			"application/xml, application/json",
			true,
			http.StatusOK,
		},
		{
			"SupportedAcceptHeaderValueDoesNotPresentOnListOfValuesTest",
			"application/xml, text/plain",
			false,
			http.StatusNotAcceptable,
		},
		{
			"MissingAcceptHeaderValueTest",
			"",
			true,
			http.StatusOK,
		},
		{
			"AllMediaTypeAcceptHeaderValueTest",
			"*/*",
			true,
			http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)

		wrappedHander := NewWrappedHander()
		produce := Produce(supportedAccept)(wrappedHander.Handler)
		ts := httptest.NewServer(produce)
		defer ts.Close()
		clinet := http.Client{}

		req, err := http.NewRequest("GET", ts.URL+"/testUrl", nil)
		if err != nil {
			t.Error(err)
		}

		req.Header.Set("Accept", testCase.AcceptHeader)
		res, err := clinet.Do(req)
		if err != nil {
			t.Error(err)
		}

		if res.StatusCode != testCase.ExpectedStatusCode {
			t.Errorf("Wrong status code: expected %v given %v", testCase.ExpectedStatusCode, res.StatusCode)
		}

		if wrappedHander.WasCalled != testCase.ExceptedIsWrappedHandlerCalled {
			t.Errorf("WasCalled is different than expected. Expected %v given %v ",
				testCase.ExceptedIsWrappedHandlerCalled, wrappedHander.WasCalled)
		}
	}
}

func TestConsume(t *testing.T) {
	supportedContentType := "application/json"
	testCases := []struct {
		Name                           string
		ContentTypeHeader              string
		ExceptedIsWrappedHandlerCalled bool
		ExpectedStatusCode             int
	}{
		{
			"SupportedContentTypeHeaderTest",
			supportedContentType,
			true,
			http.StatusOK,
		},
		{
			"UnsupportedContentTypeHeaderTest",
			"text/plain",
			false,
			http.StatusUnsupportedMediaType,
		},
		{
			"EmptyContentTypeHeaderTest",
			"",
			false,
			http.StatusUnsupportedMediaType,
		},
		{
			"AllMediaTypeContentTypeHeaderTest",
			"*/*",
			false,
			http.StatusUnsupportedMediaType,
		},
		{
			"SupportedContentTypeHeaderOnValueListTest",
			"text/plain, application/json",
			true,
			http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)

		wrappedHandler := NewWrappedHander()
		consume := Consume(supportedContentType)(wrappedHandler.Handler)
		ts := httptest.NewServer(consume)
		defer ts.Close()

		client := http.Client{}
		req, err := http.NewRequest("POST", ts.URL+"/testUrl", nil)
		if err != nil {
			t.Error(err)
		}

		req.Header.Add("Content-Type", testCase.ContentTypeHeader)
		res, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}

		if res.StatusCode != testCase.ExpectedStatusCode {
			t.Errorf("Wrong status code. Expected %v given %v",
				testCase.ExpectedStatusCode, res.StatusCode)
		}

		if wrappedHandler.WasCalled != testCase.ExceptedIsWrappedHandlerCalled {
			t.Errorf("WasCalled is different than expected. Expected %v given %v ",
				testCase.ExceptedIsWrappedHandlerCalled, wrappedHandler.WasCalled)
		}
	}
}

type ThrottleMock struct {
	OnAllow bool
}

func (tm *ThrottleMock) Limit() rate.Limit {
	return 0.0
}
func (tm *ThrottleMock) Wait(ctx context.Context) error {
	return nil
}

func (tm *ThrottleMock) Allow() bool {
	return tm.OnAllow
}

func TestThrottle(t *testing.T) {
	testCases := []struct {
		Name                  string
		AllowResult           bool
		ShouldCallWrappedFunc bool
		ExpectedStatusCode    int
	}{
		{"AllowedToPerformRequestTest", true, true, http.StatusOK},
		{"NotAllowedToPerformRequestTest", false, false, http.StatusForbidden},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)

		wrappedHandler := NewWrappedHander()
		throttleHandler := Throttle(&ThrottleMock{OnAllow: testCase.AllowResult})(wrappedHandler.Handler)
		ts := httptest.NewServer(throttleHandler)
		defer ts.Close()

		client := http.Client{}
		req, err := http.NewRequest("POST", ts.URL+"/testUrl", nil)
		if err != nil {
			t.Error(err)
		}

		res, err := client.Do(req)
		if err != nil {
			t.Error("Cannot to perform request")
		}

		if res.StatusCode != testCase.ExpectedStatusCode {
			t.Errorf("Response status code does not equest to expected status code: %v, %v", res.StatusCode, testCase.ExpectedStatusCode)
		}

		if wrappedHandler.WasCalled != testCase.ShouldCallWrappedFunc {
			t.Error("WasCalled does not equal to expected value: %v, %v", wrappedHandler.WasCalled, testCase.ShouldCallWrappedFunc)
		}
	}
}

func TestPer(t *testing.T) {
	testCase := []struct {
		events         int
		duration       time.Duration
		expectedResult float64
	}{
		{1, time.Second, 1},
		{1, 2 * time.Second, 0.5},
		{1, 1 * time.Minute, 1.0 / 60.0},
		{3, 1 * time.Minute, 3.0 / 60.0},
	}

	for _, testCase := range testCase {
		if result := Per(testCase.events, testCase.duration); result != testCase.expectedResult {
			t.Errorf("Per returns %v for events %v per duration %v. Expected %v", result, testCase.events, testCase.duration, testCase.expectedResult)
		}
	}

}

func request(t testing.TB, url string) *http.Request {
	validToken, err := utils.NewUserToken("anycmon", time.Now().Add(time.Duration(10*time.Minute)).Unix())
	if err != nil {
		t.Error(err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", "Bearer "+validToken)
	return req
}
