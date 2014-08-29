package fetch

import (
	"fmt"

	"github.com/jprobinson/eazye"
	"github.com/jprobinson/newshound"
)

func GetMail(cfg *newshound.Config) {
	mail := make(chan eazye.Response)
	go eazye.GenerateUnread(cfg.MailboxInfo, cfg.MarkRead, false, mail)

	for email := range mail {
		fmt.Printf("MAAAIL:\n%+v\n", email.Email)
	}
}
