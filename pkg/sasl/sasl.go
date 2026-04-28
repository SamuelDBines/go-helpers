package sasl

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type SASLMechanism string

const (
	Anonymous   SASLMechanism = "ANONYMOUS"
	External    SASLMechanism = "EXTERNAL"
	Plain       SASLMechanism = "PLAIN"
	Login       SASLMechanism = "LOGIN"
	OAuthBearer SASLMechanism = "OAUTHBEARER"
)

var expectedChallenge = []byte("Password:")

var (
	ErrUnexpectedClientResponse  = errors.New("sasl: unexpected client response")
	ErrUnexpectedServerChallenge = errors.New("sasl: unexpected server challenge")
)

type anonymousClient struct {
	Trace string
}

type externalClient struct {
	Identity string
}

type loginClient struct {
	Username string
	Password string
}

type plainClient struct {
	Identity string
	Username string
	Password string
}

type OAuthBearerError struct {
	Status  string `json:"status"`
	Schemes string `json:"schemes"`
	Scope   string `json:"scope"`
}

type OAuthBearerOptions struct {
	Username string
	Token    string
	Host     string
	Port     int
}

type oauthBearerClient struct {
	OAuthBearerOptions
}

type AnonymousAuthenticator func(trace string) error
type ExternalAuthenticator func(identity string) error
type PlainAuthenticator func(identity, username, password string) error
type OAuthBearerAuthenticator func(opts OAuthBearerOptions) *OAuthBearerError

type Client interface {
	Start() (mech SASLMechanism, ir []byte, err error)
	Next(challeng []byte) (response []byte, err error)
}

type Server interface {
	Next(response []byte) (challenge []byte, done bool, err error)
}

// Start

func (c *anonymousClient) Start() (mech SASLMechanism, ir []byte, err error) {
	return Anonymous, []byte(c.Trace), nil
}

func (a *plainClient) Start() (mech SASLMechanism, ir []byte, err error) {
	return Plain, []byte(a.Identity + "\x00" + a.Username + "\x00" + a.Password), nil
}

func (a *loginClient) Start() (mech SASLMechanism, ir []byte, err error) {
	return Login, []byte(a.Username), nil
}

func (a *externalClient) Start() (mech SASLMechanism, ir []byte, err error) {
	return External, []byte(a.Identity), nil
}

func (a *oauthBearerClient) Start() (mech SASLMechanism, ir []byte, err error) {
	var authzid string
	if a.Username != "" {
		authzid = "a=" + a.Username
	}
	str := "n," + authzid + ","

	if a.Host != "" {
		str += "\x01host=" + a.Host
	}

	if a.Port != 0 {
		str += "\x01port=" + strconv.Itoa(a.Port)
	}
	str += "\x01auth=Bearer " + a.Token + "\x01\x01"
	ir = []byte(str)
	return OAuthBearer, ir, nil
}

// Client

// Next

func (c *anonymousClient) Next(challenge []byte) (response []byte, err error) {
	return nil, ErrUnexpectedServerChallenge
}

func (a *externalClient) Next(challenge []byte) (response []byte, err error) {
	return nil, ErrUnexpectedServerChallenge
}

func (a *loginClient) Next(challenge []byte) (response []byte, err error) {
	if bytes.Compare(challenge, expectedChallenge) != 0 {
		return nil, ErrUnexpectedServerChallenge
	}
	return []byte(a.Password), nil
}

func (err *OAuthBearerError) Error() string {
	return fmt.Sprintf("OAUTHBEARER authentication error (%v)", err.Status)
}

func (a *oauthBearerClient) Next(challenge []byte) ([]byte, error) {
	authBearerErr := &OAuthBearerError{}
	if err := json.Unmarshal(challenge, authBearerErr); err != nil {
		return nil, err
	}
	return nil, authBearerErr
}

func (a *plainClient) Next(challenge []byte) (response []byte, err error) {
	return nil, ErrUnexpectedServerChallenge
}

// New
func NewAnonymousClient(trace string) Client {
	return &anonymousClient{trace}
}

func NewExternalClient(identity string) Client {
	return &externalClient{identity}
}

func NewLoginClient(username, password string) Client {
	return &loginClient{username, password}
}

func NewOAuthBearerClient(opt *OAuthBearerOptions) Client {
	return &oauthBearerClient{*opt}
}

func NewPlainClient(identity, username, password string) Client {
	return &plainClient{identity, username, password}
}

// Server

type externalServer struct {
	done         bool
	authenticate ExternalAuthenticator
}

type anonymousServer struct {
	done         bool
	authenticate AnonymousAuthenticator
}

type plainServer struct {
	done         bool
	authenticate PlainAuthenticator
}
type oauthBearerServer struct {
	done         bool
	failErr      error
	authenticate OAuthBearerAuthenticator
}

// Next

func (s *anonymousServer) Next(response []byte) (challenge []byte, done bool, err error) {
	if s.done {
		err = ErrUnexpectedClientResponse
		return
	}

	if response == nil {
		challenge = []byte{}
		return
	}

	s.done = true

	err = s.authenticate(string(response))
	done = true
	return
}

func (a *externalServer) Next(response []byte) (challenge []byte, done bool, err error) {
	if a.done {
		err = ErrUnexpectedClientResponse
		return
	}

	if response == nil {
		challenge = []byte{}
		return
	}

	a.done = true

	if bytes.Contains(response, []byte("\x00")) {
		return nil, false, errors.New("sasl: identity contains a NUL character")
	}

	err = a.authenticate(string(response))
	done = true
	return
}

func (a *oauthBearerServer) fail(descr string) ([]byte, bool, error) {
	blob, err := json.Marshal(OAuthBearerError{
		Status:  "invalid_request",
		Schemes: "bearer",
	})
	if err != nil {
		panic(err) // wtf
	}
	a.failErr = errors.New("sasl: client error: " + descr)
	return blob, false, nil
}

func (a *oauthBearerServer) Next(response []byte) (challenge []byte, done bool, err error) {
	// Per RFC, we cannot just send an error, we need to return JSON-structured
	// value as a challenge and then after getting dummy response from the
	// client stop the exchange.
	if a.failErr != nil {
		// Server libraries (go-smtp, go-imap) will not call Next on
		// protocol-specific SASL cancel response ('*'). However, GS2 (and
		// indirectly OAUTHBEARER) defines a protocol-independent way to do so
		// using 0x01.
		if len(response) != 1 && response[0] != 0x01 {
			return nil, true, errors.New("sasl: invalid response")
		}
		return nil, true, a.failErr
	}

	if a.done {
		err = ErrUnexpectedClientResponse
		return
	}

	// Generate empty challenge.
	if response == nil {
		return []byte{}, false, nil
	}

	a.done = true

	// Cut n,a=username,\x01host=...\x01auth=...
	// into
	//   n
	//   a=username
	//   \x01host=...\x01auth=...\x01\x01
	parts := bytes.SplitN(response, []byte{','}, 3)
	if len(parts) != 3 {
		return a.fail("Invalid response")
	}
	flag := parts[0]
	authzid := parts[1]
	if !bytes.Equal(flag, []byte{'n'}) {
		return a.fail("Invalid response, missing 'n' in gs2-cb-flag")
	}
	opts := OAuthBearerOptions{}
	if len(authzid) > 0 {
		if !bytes.HasPrefix(authzid, []byte("a=")) {
			return a.fail("Invalid response, missing 'a=' in gs2-authzid")
		}
		opts.Username = string(bytes.TrimPrefix(authzid, []byte("a=")))
	}

	// Cut \x01host=...\x01auth=...\x01\x01
	// into
	//   *empty*
	//   host=...
	//   auth=...
	//   *empty*
	//
	// Note that this code does not do a lot of checks to make sure the input
	// follows the exact format specified by RFC.
	params := bytes.Split(parts[2], []byte{0x01})
	for _, p := range params {
		// Skip empty fields (one at start and end).
		if len(p) == 0 {
			continue
		}

		pParts := bytes.SplitN(p, []byte{'='}, 2)
		if len(pParts) != 2 {
			return a.fail("Invalid response, missing '='")
		}

		switch string(pParts[0]) {
		case "host":
			opts.Host = string(pParts[1])
		case "port":
			port, err := strconv.ParseUint(string(pParts[1]), 10, 16)
			if err != nil {
				return a.fail("Invalid response, malformed 'port' value")
			}
			opts.Port = int(port)
		case "auth":
			const prefix = "bearer "
			strValue := string(pParts[1])
			// Token type is case-insensitive.
			if !strings.HasPrefix(strings.ToLower(strValue), prefix) {
				return a.fail("Unsupported token type")
			}
			opts.Token = strValue[len(prefix):]
		default:
			return a.fail("Invalid response, unknown parameter: " + string(pParts[0]))
		}
	}

	authzErr := a.authenticate(opts)
	if authzErr != nil {
		blob, err := json.Marshal(authzErr)
		if err != nil {
			panic(err) // wtf
		}
		a.failErr = authzErr
		return blob, false, nil
	}

	return nil, true, nil
}

func (a *plainServer) Next(response []byte) (challenge []byte, done bool, err error) {
	if a.done {
		err = ErrUnexpectedClientResponse
		return
	}

	// No initial response, send an empty challenge
	if response == nil {
		return []byte{}, false, nil
	}

	a.done = true

	parts := bytes.Split(response, []byte("\x00"))
	if len(parts) != 3 {
		err = errors.New("sasl: invalid response")
		return
	}

	identity := string(parts[0])
	username := string(parts[1])
	password := string(parts[2])

	err = a.authenticate(identity, username, password)
	done = true
	return
}

// New
func NewAnonymousServer(authenticator AnonymousAuthenticator) Server {
	return &anonymousServer{authenticate: authenticator}
}

func NewExternalServer(authenticator ExternalAuthenticator) Server {
	return &externalServer{authenticate: authenticator}
}

func NewOAuthBearerServer(auth OAuthBearerAuthenticator) Server {
	return &oauthBearerServer{authenticate: auth}
}

func NewPlainServer(authenticator PlainAuthenticator) Server {
	return &plainServer{authenticate: authenticator}
}
