package lesson3_test

import (
	"fmt"
	"lesson3"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInitialBodyState(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/get", nil)
	respRec := httptest.NewRecorder()

	lesson3.GetHandler(respRec, req)

	if respRec.Body.String() != "" {
		t.Errorf("Bad initial body value")
	}
}

func TestReplaceSavesBody(t *testing.T) {
	body := "{'lol': 'kek'}"
	replaceReq := httptest.NewRequest(http.MethodPost, "/replace", strings.NewReader(body))
	respRec := httptest.NewRecorder()
	lesson3.ReplaceHandler(respRec, replaceReq)

	if respRec.Result().StatusCode != http.StatusOK {
		t.Errorf("/replace went wrong")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/get", nil)
	lesson3.GetHandler(respRec, getReq)

	if respRec.Result().StatusCode != http.StatusOK {
		t.Errorf("/get went wrong")
	}

	if actual := respRec.Body.String(); actual != body {
		t.Errorf("/get \n expected: %s \n actual: %s", body, actual)
	}

}

func TestMultipleReplace(t *testing.T) {

	TestReplaceSavesBody(t)

	for cnt := 0; cnt < 10; cnt++ {

		body := fmt.Sprintf("body %d", cnt)

		t.Run("go /replace", func(t *testing.T) {
			t.Parallel()

			goRec := httptest.NewRecorder()
			goReqReplace := httptest.NewRequest(http.MethodPost, "/replace", strings.NewReader(body))
			lesson3.GetHandler(goRec, goReqReplace)

			if err := goRec.Result().StatusCode; err != http.StatusOK {
				t.Errorf("/replace \n error: %d", err)
			}

			goReqGet := httptest.NewRequest(http.MethodGet, "/get", nil)
			lesson3.GetHandler(goRec, goReqGet)

			if err := goRec.Result().StatusCode; err != http.StatusOK {
				t.Errorf("/get \n error: %d", err)
			}

			if actual := goRec.Body.String(); actual != body {
				t.Errorf("/get \n expected: '%s'\n actual: '%s'", body, actual)
			}
		})
	}
}
