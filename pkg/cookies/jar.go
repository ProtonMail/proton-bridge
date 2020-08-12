package cookies

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"

	"github.com/sirupsen/logrus"
)

type Jar struct {
	jar       *cookiejar.Jar
	persister *Persister
	locker    sync.Locker
}

func New(persister *Persister) (*Jar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	cookies, err := persister.Load()
	if err != nil {
		return nil, err
	}

	for rawURL, cookies := range cookies {
		url, err := url.Parse(rawURL)
		if err != nil {
			continue
		}

		jar.SetCookies(url, cookies)
	}

	return &Jar{
		jar:       jar,
		persister: persister,
		locker:    &sync.Mutex{},
	}, nil
}

func (j *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.locker.Lock()
	defer j.locker.Unlock()

	j.jar.SetCookies(u, cookies)

	if err := j.persister.Persist(u.String(), cookies); err != nil {
		logrus.WithError(err).Warn("Failed to persist cookie")
	}
}

func (j *Jar) Cookies(u *url.URL) []*http.Cookie {
	j.locker.Lock()
	defer j.locker.Unlock()

	return j.jar.Cookies(u)
}
