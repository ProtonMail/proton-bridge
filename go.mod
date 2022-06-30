module github.com/ProtonMail/proton-bridge/v2

go 1.15

// These dependencies are `replace`d below, so the version numbers should be ignored.
// They are in a separate require block to highlight this.
require (
	github.com/docker/docker-credential-helpers v0.6.3
	github.com/emersion/go-imap v1.0.6
)

require (
	github.com/0xAX/notificator v0.0.0-20191016112426-3962a5ea8da1
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/ProtonMail/bcrypt v0.0.0-20211005172633-e235017c1baf // indirect
	github.com/ProtonMail/go-autostart v0.0.0-20181114175602-c5272053443a
	github.com/ProtonMail/go-crypto v0.0.0-20220623141421-5afb4c282135
	github.com/ProtonMail/go-imap-id v0.0.0-20190926060100-f94a56b9ecde
	github.com/ProtonMail/go-rfc5322 v0.8.0
	github.com/ProtonMail/go-srp v0.0.5
	github.com/ProtonMail/go-vcard v0.0.0-20180326232728-33aaa0a0c8a5
	github.com/ProtonMail/gopenpgp/v2 v2.4.7
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/abiosoft/ishell v2.0.0+incompatible
	github.com/abiosoft/readline v0.0.0-20180607040430-155bce2042db // indirect
	github.com/allan-simon/go-singleinstance v0.0.0-20160830203053-79edcfdc2dfc
	github.com/chzyer/logex v1.1.10 // indirect
	github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1 // indirect
	github.com/cucumber/godog v0.12.1
	github.com/cucumber/messages-go/v16 v16.0.1
	github.com/elastic/go-sysinfo v1.7.1
	github.com/elastic/go-windows v1.0.1 // indirect
	github.com/emersion/go-imap-appendlimit v0.0.0-20190308131241-25671c986a6a
	github.com/emersion/go-imap-move v0.0.0-20190710073258-6e5a51a5b342
	github.com/emersion/go-imap-quota v0.0.0-20210203125329-619074823f3c
	github.com/emersion/go-imap-unselect v0.0.0-20171113212723-b985794e5f26
	github.com/emersion/go-message v0.12.1-0.20201221184100-40c3f864532b
	github.com/emersion/go-sasl v0.0.0-20200509203442-7bfe0ed36a21
	github.com/emersion/go-smtp v0.14.0
	github.com/emersion/go-textwrapper v0.0.0-20200911093747-65d896831594
	github.com/emersion/go-vcard v0.0.0-20190105225839-8856043f13c5 // indirect
	github.com/fatih/color v1.9.0
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/getsentry/sentry-go v0.12.0
	github.com/go-resty/resty/v2 v2.6.0
	github.com/godbus/dbus v4.1.0+incompatible
	github.com/golang/mock v1.4.4
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.1.1
	github.com/hashicorp/go-multierror v1.1.0
	github.com/jameskeane/bcrypt v0.0.0-20120420032655-c3cd44c1e20f // indirect
	github.com/jaytaylor/html2text v0.0.0-20200412013138-3577fbdbcff7
	github.com/keybase/go-keychain v0.0.0
	github.com/kr/text v0.2.0 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/miekg/dns v1.1.41
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/nsf/jsondiff v0.0.0-20200515183724-f29ed568f4ce
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/ricochet2200/go-disk-usage/du v0.0.0-20210707232629-ac9918953285
	github.com/sirupsen/logrus v1.7.0
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/stretchr/testify v1.7.0
	github.com/therecipe/qt v0.0.0-20200701200531-7f61353ee73e
	github.com/therecipe/qt/internal/binding/files/docs/5.12.0 v0.0.0-20200904063919-c0c124a5770d // indirect
	github.com/therecipe/qt/internal/binding/files/docs/5.13.0 v0.0.0-20200904063919-c0c124a5770d // indirect
	github.com/urfave/cli/v2 v2.2.0
	github.com/vmihailenco/msgpack/v5 v5.1.3
	go.etcd.io/bbolt v1.3.6
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4
	golang.org/x/sys v0.0.0-20220627191245-f75cf1eec38b
	golang.org/x/text v0.3.7
	howett.net/plist v1.0.0
)

replace (
	github.com/docker/docker-credential-helpers => github.com/ProtonMail/docker-credential-helpers v1.1.0
	github.com/emersion/go-imap => github.com/ProtonMail/go-imap v0.0.0-20201228133358-4db68cea0cac
	github.com/emersion/go-message => github.com/ProtonMail/go-message v0.0.0-20210611055058-fabeff2ec753
	github.com/keybase/go-keychain => github.com/cuthix/go-keychain v0.0.0-20220405075754-31e7cee908fe
)
