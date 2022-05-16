// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package grpc

//goland:noinspection SpellCheckingInspection
const (
	serverCert = `-----BEGIN CERTIFICATE-----
MIIC5TCCAc2gAwIBAgIJAMUQK0VGexMsMA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNV
BAMMCWxvY2FsaG9zdDAeFw0yMjA2MTQxNjUyNTVaFw0yMjA3MTQxNjUyNTVaMBQx
EjAQBgNVBAMMCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
ggEBAL6T1JQ0jptq512PBLASpCLFB0px7KIzEml0oMUCkVgUF+2cayrvdBXJZnaO
SG+/JPnHDcQ/ecgqkh2Ii6a2x2kWA5KqWiV+bSHp0drXyUGJfM85muLsnrhYwJ83
HHtweoUVebRZvHn66KjaH8nBJ+YVWyYbSUhJezcg6nBSEtkW+I/XUHu4S2C7FUc5
DXPO3yWWZuZ22OZz70DY3uYE/9COuilotuKdj7XgeKDyKIvRXjPFyqGxwnnp6bXC
vWvrQdcxy0wM+vZxew3QtA/Ag9uKJU9owP6noauXw95l49lEVIA5KXVNtdaldVht
MO/QoelLZC7h79PK22zbii3x930CAwEAAaM6MDgwFAYDVR0RBA0wC4IJbG9jYWxo
b3N0MAsGA1UdDwQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDATANBgkqhkiG9w0B
AQsFAAOCAQEAW/9PE8dcAN+0C3K96Xd6Y3qOOtQhRw+WlZXhtiqMtlJfTjvuGKs9
58xuKcTvU5oobxLv+i5+4gpqLjUZZ9FBnYXZIACNVzq4PEXf+YdzcA+y6RS/rqT4
dUjsuYrScAmdXK03Duw3HWYrTp8gsJzIaYGTltUrOn0E4k/TsZb/tZ6z+oH7Fi+p
wdsI6Ut6Zwm3Z7WLn5DDk8KvFjHjZkdsCb82SFSAUVrzWo5EtbLIY/7y3A5rGp9D
t0AVpuGPo5Vn+MW1WA9HT8lhjz0v5wKGMOBi3VYW+Yx8FWHDpacvbZwVM0MjMSAd
M7SXYbNDiLF4LwPLsunoLsW133Ky7s99MA==
-----END CERTIFICATE-----`

	serverKey = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC+k9SUNI6baudd
jwSwEqQixQdKceyiMxJpdKDFApFYFBftnGsq73QVyWZ2jkhvvyT5xw3EP3nIKpId
iIumtsdpFgOSqlolfm0h6dHa18lBiXzPOZri7J64WMCfNxx7cHqFFXm0Wbx5+uio
2h/JwSfmFVsmG0lISXs3IOpwUhLZFviP11B7uEtguxVHOQ1zzt8llmbmdtjmc+9A
2N7mBP/QjropaLbinY+14Hig8iiL0V4zxcqhscJ56em1wr1r60HXMctMDPr2cXsN
0LQPwIPbiiVPaMD+p6Grl8PeZePZRFSAOSl1TbXWpXVYbTDv0KHpS2Qu4e/Tytts
24ot8fd9AgMBAAECggEBAJFkGpOOnRU4s5YO3BavwgS8p9lFnLAJooxNa7GhSd0W
R0MBSEkTMU7FvaPI3L5T5xOfpoMHohLxV1Osrk3bt7oWD1e/GtLr5routejtIx8a
kttNKTriJhyhqSJOWy5ZGz+YqKbMpxuwLftTnVjAQX4o4MbrnjbFyHjAZdqW4sY2
jLulfEdOave6nxaEocmIkoXEjuX90LB+yNG6ncSYM3GV+IyCVw7DsoU4dLd/IRDa
4iJVF7tVdAsZqN6/EVYXpGqG0t1HI8ddacHa1qWgCG3kBB+3faxXZcDJdlRrXLUQ
4jLH8oEfXOb5YgCwyYzW2EynXEpG5vjsPmsCWJY/mIECgYEA52av81+lui97KLg+
T07XtR8zJPMkHnBNfc6ooWku/+0NuQPpUq14vqzRVut9jBHUDP3xSvrPnXsp15ZA
/mipLQLNKssTYtk90cyGqLUkrd/NPLFZLXToBfWBlfazdcJQQRIxZ2dTy5MH+HIU
Oio3LZi+iDIbdzzSlmL8PaLit20CgYEA0tYsswhq6OaWx25iu4hBMRlt6hr9qGVW
jlzCFjBhlh3YtoBti2w2fsJdU+hUpeXU327fhFmdCQFXtf+Om5CSHihmJ+mHj9O1
5Jd6zn4o8szdg5je9T4gt7KG6QdXaFJ2aMuq+SxZl1NIE+9qnf/qom4GHHZ/Nj41
vwlQu+zS5lECgYAOzSK0DoorPp5CHIbfy8tAap563pKQ394VDgL7UB8Rf7hA/V8P
SslOaP9679U4AGvv6M5mXWSqThZ/E71UiJ1Jo8Q72IGE8SBjKxHx+KQ/+vDF0RJD
NhchSnLfhMg14BgCEYfXdWSGwQDhg2qHzet5nyuQyqO3HMzbkblQt/qIgQKBgHLv
nPiQmy+SHRplO9+93MQ2d6wKwMNfUztSp9/OyjQ62xxKkO1TtbWOobAPVK4Hx+9y
EtmkvK3fFIC763M08eMM5PvXHDa1FFCkn6cYMZyDQDLwUINjNhTOdytr/CN76N8i
QHeLzN9o4D814mp1y+R2lFBJ7PmWGlilbGS2KxaxAoGAFMsb1MER+eTOUO3z05Di
lts4VRWQhq2frd/on6AcTv4idQox1RcOrKWQbRVgeQVY1SkkHhg8lN0jX3W3EfuQ
aOfyky04GbLiwO8NRHZMlORWLxlCkrUrb6Va+LQlT0JvpQbqdbu6Ix8NomG9K697
aScKmY7bGC0ki2IIdt2YZ5I=
-----END PRIVATE KEY-----`
)
