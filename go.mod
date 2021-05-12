module github.com/ProtonMail/proton-bridge

go 1.13

// These dependencies are `replace`d below, so the version numbers should be ignored.
// They are in a separate require block to highlight this.
require (
	github.com/docker/docker-credential-helpers v0.6.3
	github.com/emersion/go-imap v1.0.6
	github.com/jameskeane/bcrypt v0.0.0-20170924085257-7509ea014998
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
)

require (
	github.com/0xAX/notificator v0.0.0-20191016112426-3962a5ea8da1
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/ProtonMail/go-autostart v0.0.0-20181114175602-c5272053443a
	github.com/ProtonMail/go-imap-id v0.0.0-20190926060100-f94a56b9ecde
	github.com/ProtonMail/go-rfc5322 v0.5.0
	github.com/ProtonMail/go-vcard v0.0.0-20180326232728-33aaa0a0c8a5
	github.com/ProtonMail/gopenpgp/v2 v2.1.3
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/abiosoft/ishell v2.0.0+incompatible
	github.com/abiosoft/readline v0.0.0-20180607040430-155bce2042db // indirect
	github.com/allan-simon/go-singleinstance v0.0.0-20160830203053-79edcfdc2dfc
	github.com/chzyer/logex v1.1.10 // indirect
	github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1 // indirect
	github.com/cucumber/godog v0.8.1
	github.com/emersion/go-imap-appendlimit v0.0.0-20190308131241-25671c986a6a
	github.com/emersion/go-imap-move v0.0.0-20190710073258-6e5a51a5b342
	github.com/emersion/go-imap-quota v0.0.0-20210203125329-619074823f3c
	github.com/emersion/go-imap-unselect v0.0.0-20171113212723-b985794e5f26
	github.com/emersion/go-mbox v1.0.2
	github.com/emersion/go-message v0.12.1-0.20201221184100-40c3f864532b
	github.com/emersion/go-sasl v0.0.0-20200509203442-7bfe0ed36a21
	github.com/emersion/go-smtp v0.14.0
	github.com/emersion/go-textwrapper v0.0.0-20200911093747-65d896831594
	github.com/emersion/go-vcard v0.0.0-20190105225839-8856043f13c5 // indirect
	github.com/fatih/color v1.9.0
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/getsentry/sentry-go v0.8.0
	github.com/go-resty/resty/v2 v2.6.0
	github.com/golang/mock v1.4.4
	github.com/google/go-cmp v0.5.1
	github.com/google/uuid v1.1.1
	github.com/gopherjs/gopherjs v0.0.0-20190430165422-3e4dfb77656c // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/jaytaylor/html2text v0.0.0-20200412013138-3577fbdbcff7
	github.com/keybase/go-keychain v0.0.0-20200502122510-cda31fe0c86d
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/miekg/dns v1.1.41
	github.com/nsf/jsondiff v0.0.0-20200515183724-f29ed568f4ce
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/stretchr/testify v1.6.1
	github.com/therecipe/qt v0.0.0-20200701200531-7f61353ee73e
	github.com/urfave/cli/v2 v2.2.0
	github.com/vmihailenco/msgpack/v5 v5.1.3
	go.etcd.io/bbolt v1.3.5
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	golang.org/x/text v0.3.5-0.20201125200606-c27b9fd57aec
)

replace (
	github.com/docker/docker-credential-helpers => github.com/ProtonMail/docker-credential-helpers v1.1.0
	github.com/emersion/go-imap => github.com/ProtonMail/go-imap v0.0.0-20201228133358-4db68cea0cac
	github.com/jameskeane/bcrypt => github.com/ProtonMail/bcrypt v0.0.0-20170924085257-7509ea014998
	golang.org/x/crypto => github.com/ProtonMail/crypto v0.0.0-20201112115411-41db4ea0dd1c
)
