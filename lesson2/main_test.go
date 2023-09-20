package lesson2_test

import (
	"fmt"
	"lesson2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInitialBodyState(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/get", nil)
	respRec := httptest.NewRecorder()

	lesson2.GetHandler(respRec, req)

	if respRec.Body.String() != "" {
		t.Errorf("Bad initial body value")
	}
}

func TestReplaceSavesBody(t *testing.T) {
	body := "{'lol': 'kek'}"
	replaceReq := httptest.NewRequest(http.MethodPost, "/replace", strings.NewReader(body))
	respRec := httptest.NewRecorder()
	lesson2.ReplaceHandler(respRec, replaceReq)

	if respRec.Result().StatusCode != http.StatusOK {
		t.Errorf("/replace went wrong")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/get", nil)
	lesson2.GetHandler(respRec, getReq)

	if respRec.Result().StatusCode != http.StatusOK {
		t.Errorf("/get went wrong")
	}

	if actual := respRec.Body.String(); actual != body {
		t.Errorf("/get expected %s got %s", body, actual)
	}

}

func TestMultipleReplace(t *testing.T) {

	TestReplaceSavesBody(t)

	for cnt := 0; cnt < 10; cnt++ {

		body, i := fmt.Sprintf("body %d", cnt), cnt

		t.Run("go /replace", func(t *testing.T) {
			t.Parallel()

			goRec := httptest.NewRecorder()
			goReqReplace := httptest.NewRequest(http.MethodPost, "/replace", strings.NewReader(body))
			goReqGet := httptest.NewRequest(http.MethodGet, "/get", nil)

			lesson2.GetHandler(goRec, goReqReplace)

			if err := goRec.Result().StatusCode; err != http.StatusOK {
				t.Errorf("go /replace went wrong on %d, got %d", i, err)
			}

			lesson2.GetHandler(goRec, goReqGet)

			if goRec.Result().StatusCode != http.StatusOK {
				t.Errorf("go /get went wrong on %d", i)
			}

			if actual := goRec.Body.String(); actual != body {
				t.Errorf("go /get got %s, expected %s on %d", actual, body, i)
			}
		})
	}
}
