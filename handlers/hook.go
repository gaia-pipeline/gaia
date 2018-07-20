package handlers

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gaia-pipeline/gaia"

	"github.com/labstack/echo"
)

// Hook represent a github based webhook context.
type Hook struct {
	Signature string
	Event     string
	ID        string
	Payload   []byte
}

const secret = "superawesomesecretgithubpassword"

func signBody(secret, body []byte) []byte {
	computed := hmac.New(sha1.New, secret)
	computed.Write(body)
	return []byte(computed.Sum(nil))
}

func verifySignature(secret []byte, signature string, body []byte) bool {

	signaturePrefix := "sha1="
	signatureLength := 45

	if len(signature) != signatureLength || !strings.HasPrefix(signature, signaturePrefix) {
		return false
	}

	actual := make([]byte, 20)
	hex.Decode(actual, []byte(signature[5:]))
	expected := signBody(secret, body)
	return hmac.Equal(expected, actual)
}

func parse(secret []byte, req *http.Request) (Hook, error) {
	h := Hook{}

	if h.Signature = req.Header.Get("x-hub-signature"); len(h.Signature) == 0 {
		return Hook{}, errors.New("no signature")
	}

	if h.Event = req.Header.Get("x-github-event"); len(h.Event) == 0 {
		return Hook{}, errors.New("no event")
	}

	if h.ID = req.Header.Get("x-github-delivery"); len(h.ID) == 0 {
		return Hook{}, errors.New("no event id")
	}

	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		return Hook{}, err
	}

	if !verifySignature(secret, h.Signature, body) {
		return Hook{}, errors.New("Invalid signature")
	}

	h.Payload = body
	return h, err
}

// GitWebHook handles callbacks from GitHub's webhook system.
func GitWebHook(c echo.Context) error {
	h, err := parse([]byte(secret), c.Request())
	c.Request().Header.Set("Content-type", "application/json")

	if err != nil {
		e := struct {
			Message string
		}{
			Message: err.Error(),
		}
		return c.JSON(400, e)
	}

	gaia.Cfg.Logger.Info("received: ", h.Event)
	return c.JSON(http.StatusOK, struct{}{})
}
