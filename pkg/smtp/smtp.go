package smtp

import (
	"io"
	"time"

	"github.com/SamuelDBines/go-helpers/pkg/sasl"
)

type BodyType string

const (
	Body7Bit       BodyType = "7BIT"
	Body8BitMIME   BodyType = "8BITMIME"
	BodyBinaryMIME BodyType = "BINARYMIME"
)

type DSNReturn string

const (
	DSNReturnFull    DSNReturn = "FULL"
	DSNReturnHeaders DSNReturn = "HDRS"
)

type DSNNotify string

const (
	DSNNotifyNever   DSNNotify = "NEVER"
	DSNNotifyDelayed DSNNotify = "DELAY"
	DSNNotifyFailure DSNNotify = "FAILURE"
	DSNNotifySuccess DSNNotify = "SUCCESS"
)

type DSNAddressType string

const (
	DSNAddressTypeRFC822 DSNAddressType = "RFC822"
	DSNAddressTypeUTF8   DSNAddressType = "UTF-8"
)

type DeliverByMode string

const (
	DeliverByNotify DeliverByMode = "N"
	DeliverByReturn DeliverByMode = "R"
)

type PriorityProfile string

const (
	PriorityUnspecified PriorityProfile = ""
	PriorityMIXER       PriorityProfile = "MIXER"
	PrioritySTANAG4406  PriorityProfile = "STANAG4406"
	PriorityNSEP        PriorityProfile = "NSEP"
)

type MailOptions struct {
	Body       BodyType
	Size       int64
	RequireTLS bool
	UTF8       bool
	Return     DSNReturn
	EnvelopeID string
	Auth       *string
}

type DeliverByOptions struct {
	Time  time.Duration
	Mode  DeliverByMode
	Trace bool
}

type RcptOptions struct {
	Notify                     []DSNNotify
	OriginalRecipientType      DSNAddressType
	OriginalRecipient          string
	RequireRecipientValidSince time.Time
	DeliverBy                  *DeliverByOptions
	MTPriority                 *int
}

var (
	ErrAuthFailed = &SMTPError{
		Code:         535,
		EnhancedCode: EnhancedCode{5, 7, 8},
		Message:      "Authentication failed",
	}
	ErrAuthRequired = &SMTPError{
		Code:         502,
		EnhancedCode: EnhancedCode{5, 7, 0},
		Message:      "Please authenticate first",
	}
	ErrAuthUnsupported = &SMTPError{
		Code:         502,
		EnhancedCode: EnhancedCode{5, 7, 0},
		Message:      "Authentication not supported",
	}
	ErrAuthUnknownMechanism = &SMTPError{
		Code:         504,
		EnhancedCode: EnhancedCode{5, 7, 4},
		Message:      "Unsupported authentication mechanism",
	}
)

type Backend interface {
	NewSession(c *Conn) (Session, error)
}

type BackendFunc func(c *Conn) (Session, error)

var _ Backend = (BackendFunc)(nil)

func (f BackendFunc) NewSession(c *Conn) (Session, error) {
	return f(c)
}

type Session interface {
	Reset()
	Logout() error
	Mail(from string, opts *MailOptions) error
	Rcpt(to string, opts *RcptOptions) error
	Data(r io.Reader) error
}

type LMTPSession interface {
	Session
	LMTPData(r io.Reader, status StatusCollector) error
}

type StatusCollector interface {
	SetStatus(rcptTo string, err error)
}

type AuthSession interface {
	Session

	AuthMechanisms() []string
	Auth(mech string) (sasl.Server, error)
}
