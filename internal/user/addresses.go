package user

import "gitlab.protontech.ch/go/liteapi"

type addrList struct {
	apiAddrs ordMap[string, string, liteapi.Address]
}

func newAddrList(apiAddrs []liteapi.Address) *addrList {
	return &addrList{
		apiAddrs: newOrdMap(
			func(addr liteapi.Address) string { return addr.ID },
			func(addr liteapi.Address) string { return addr.Email },
			func(a, b liteapi.Address) bool { return a.Order < b.Order },
			apiAddrs...,
		),
	}
}

func (list *addrList) insert(address liteapi.Address) {
	list.apiAddrs.insert(address)
}

func (list *addrList) delete(addrID string) string {
	return list.apiAddrs.delete(addrID)
}

func (list *addrList) primary() string {
	return list.apiAddrs.keys()[0]
}

func (list *addrList) addrIDs() []string {
	return list.apiAddrs.keys()
}

func (list *addrList) addrID(email string) (string, bool) {
	return list.apiAddrs.getKey(email)
}

func (list *addrList) emails() []string {
	return list.apiAddrs.values()
}

func (list *addrList) email(addrID string) (string, bool) {
	return list.apiAddrs.getVal(addrID)
}

func (list *addrList) addrMap() map[string]string {
	return list.apiAddrs.toMap()
}
