package vrm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"

	_ "github.com/rs/zerolog"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type vrmSession struct {
	token  string
	Client HTTPClient
	UserID int
}

func newVRMSession() *vrmSession {
	return &vrmSession{
		Client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (s *vrmSession) request(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if len(s.token) > 0 {
		req.Header.Add("X-Authorization", "Bearer "+s.token)
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request at %s: %w", url, err)
	}

	if !(res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices) {
		return nil, fmt.Errorf("http error: %d", res.StatusCode)
	}

	return res, nil
}

func (s *vrmSession) getAndLoad(url string, resData interface{}) error {
	res, err := s.request(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(resData); err != nil {
		return fmt.Errorf("could not decode response: %w\n\n%+v", err, res.Body)
	}

	return nil
}

func (s *vrmSession) postAndLoad(reqData interface{}, url string, resData interface{}) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(reqData); err != nil {
		return err
	}

	res, err := s.request(http.MethodPost, url, buf)
	if err != nil {
		return errors.Wrapf(err, "failed to create POST request")
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(resData); err != nil {
		return err
	}

	return nil
}

type LoginOption func(*LoginRequest)

func WithSMSToken(smsToken string) LoginOption {
	return func(r *LoginRequest) {
		r.SMSToken = smsToken
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	SMSToken string `json:"sms_token,omitempty"`
}

func Login(username, password string, opts ...LoginOption) (*vrmSession, error) {
	url, err := formatURL(loginURL, URLParams{}, nil)
	if err != nil {
		return nil, err
	}

	req := &LoginRequest{
		Username: username,
		Password: password,
	}
	for _, opt := range opts {
		opt(req)
	}

	response := struct {
		Token string `json:"token"`
		UserID int `json:"idUser"`
	}{}

	s := newVRMSession()
	if err := s.postAndLoad(req, url, &response); err != nil {
		return nil, err
	}
	s.token = response.Token
	s.UserID = response.UserID

	return s, nil
}

func LoginAsDemo() (*vrmSession, error) {
	url, err := formatURL(loginAsDemoURL, URLParams{}, nil)
	if err != nil {
		return nil, err
	}

	response := struct {
		Token string `json:"token"`
		UserID string `json:"idUser"`
	}{}

	s := newVRMSession()
	if err := s.getAndLoad(url, &response); err != nil {
		return nil, err
	}

	s.token = response.Token
	s.UserID = DemoUserID

	return s, nil
}

func (s *vrmSession) Logout() error {
	url, err := formatURL(logoutURL, URLParams{}, nil)
	if err != nil {
		return err
	}

	if err := s.postAndLoad(struct{}{}, url, nil); err != nil {
		return err
	}

	return nil
}
