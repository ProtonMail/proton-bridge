package cookies

import (
	"encoding/json"
	"net/http"

	"github.com/ProtonMail/proton-bridge/internal/preferences"
)

type Persister struct {
	prefs GetterSetter
}

type GetterSetter interface {
	Get(string) string
	Set(string, string)
}

func NewPersister(prefs GetterSetter) *Persister {
	return &Persister{prefs: prefs}
}

func (p *Persister) Persist(url string, cookies []*http.Cookie) error {
	b, err := json.Marshal(cookies)
	if err != nil {
		return err
	}

	val, err := p.load()
	if err != nil {
		return err
	}

	val[url] = string(b)

	return p.save(val)
}

func (p *Persister) Load() (map[string][]*http.Cookie, error) {
	res := make(map[string][]*http.Cookie)

	val, err := p.load()
	if err != nil {
		return nil, err
	}

	for url, rawCookies := range val {
		var cookies []*http.Cookie

		if err := json.Unmarshal([]byte(rawCookies), &cookies); err != nil {
			return nil, err
		}

		res[url] = cookies
	}

	return res, nil
}

type dataStructure map[string]string

func (p *Persister) load() (dataStructure, error) {
	b := p.prefs.Get(preferences.CookiesKey)

	if b == "" {
		if err := p.save(make(dataStructure)); err != nil {
			return nil, err
		}

		return p.load()
	}

	var val dataStructure

	if err := json.Unmarshal([]byte(b), &val); err != nil {
		return nil, err
	}

	return val, nil
}

func (p *Persister) save(val dataStructure) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}

	p.prefs.Set(preferences.CookiesKey, string(b))

	return nil
}
