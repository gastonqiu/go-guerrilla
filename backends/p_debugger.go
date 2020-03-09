package backends

import (
	"github.com/sirupsen/logrus"

	"strings"
	"time"

	"github.com/flashmob/go-guerrilla/mail"
)

// ----------------------------------------------------------------------------------
// Processor Name: debugger
// ----------------------------------------------------------------------------------
// Description   : Log received emails
// ----------------------------------------------------------------------------------
// Config Options: log_received_mails bool - log if true
// --------------:-------------------------------------------------------------------
// Input         : e.MailFrom, e.RcptTo, e.Header
// ----------------------------------------------------------------------------------
// Output        : none (only output to the log if enabled)
// ----------------------------------------------------------------------------------
func init() {
	processors[strings.ToLower(defaultProcessor)] = func() Decorator {
		return Debugger()
	}
}

type debuggerConfig struct {
	LogReceivedMails bool `json:"log_received_mails"`
	SleepSec         int  `json:"sleep_seconds,omitempty"`
}

func Debugger() Decorator {
	var config *debuggerConfig
	initFunc := InitializeWith(func(backendConfig BackendConfig) error {
		configType := BaseConfig(&debuggerConfig{})
		bcfg, err := Svc.ExtractConfig(backendConfig, configType)
		if err != nil {
			return err
		}
		config = bcfg.(*debuggerConfig)
		return nil
	})
	Svc.AddInitializer(initFunc)
	return func(p Processor) Processor {
		return ProcessWith(func(e *mail.Envelope, task SelectTask) (Result, error) {
			if task == TaskSaveMail {
				if config.LogReceivedMails {

					// Convert []Address to []string
					var mailTo []string
					for _, to := range e.RcptTo {
						mailTo = append(mailTo, to.String())
					}

					Log().WithFields(logrus.Fields{
						"from":     e.MailFrom.String(),
						"to":       strings.Join(mailTo, ", "),
						"tls":      e.TLS,
						"hello":    e.Helo,
						"username": e.Account.Username,
						"password": e.Account.Password,
						"subject":  e.Header.Get("Subject"),
						"date":     e.Header.Get("Date"),
					}).Info("email log")
				}

				if config.SleepSec > 0 {
					Log().Infof("sleeping for %d", config.SleepSec)
					time.Sleep(time.Second * time.Duration(config.SleepSec))
					Log().Infof("woke up")

					if config.SleepSec == 1 {
						panic("panic on purpose")
					}

				}

				// continue to the next Processor in the decorator stack
				return p.Process(e, task)
			} else {
				return p.Process(e, task)
			}
		})
	}
}
