// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.
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


#include "Cert.h"


// The following test TLS certificate and key are self-signed, and valid until October 2042.

QString const testTLSCert = R"(-----BEGIN CERTIFICATE-----
MIIDkzCCAnugAwIBAgIRANAfjBkSRx8LY96XwVNGT4MwDQYJKoZIhvcNAQELBQAw
SzELMAkGA1UEBhMCQ0gxEjAQBgNVBAoTCVByb3RvbiBBRzEUMBIGA1UECxMLUHJv
dG9uIE1haWwxEjAQBgNVBAMTCTEyNy4wLjAuMTAeFw0yMjEwMTAwNzI1MjFaFw00
MjEwMDUwNzI1MjFaMEsxCzAJBgNVBAYTAkNIMRIwEAYDVQQKEwlQcm90b24gQUcx
FDASBgNVBAsTC1Byb3RvbiBNYWlsMRIwEAYDVQQDEwkxMjcuMC4wLjEwggEiMA0G
CSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC3jJiMtwQMzsASB51LwWTd9TIXqh1U
F0rz+wugBPMRzoc4NfOjHpP/w6fey7Kc3wdM9wfhCQ7WgJaI+u3w5muL+ypLD1wB
KdBSiTp8pBRTvpJpaoGbl86gxsB6Rpimh+u/+1Dgh0A/b6cfvoa0gYk3PHR79oqS
Wy5EG0jCqqC9lGNBlCurAIbmY+pEF+9WKWhl+SfCOXJL+Z+rOdmhWlnWArONoKQ7
qhOSwfJj+5xudE0s5tZL7C7XKEUCNufyXt5WhMfggEzHxFdoihpkF6VOR3d60aVW
OoTbmGYCqYFlW4omWMfstj/y5FchphwnZhZoJd8bSHNtKVc8OQNhauzFAgMBAAGj
cjBwMA4GA1UdDwEB/wQEAwICpDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUH
AwIwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUI2canlIJDLw770sb2YlveYlX
SyIwDwYDVR0RBAgwBocEfwAAATANBgkqhkiG9w0BAQsFAAOCAQEAsdYfNSOLoJRf
GhQ6iS6Fue8IKwdskgTjqkBFR6xeEiGVhMCouEyQR/d+j7TwSfriWFCIsYhx8tFj
fLAawQfiU1FaLClhkdWXDndg/09mngS5r7xl6ZJZJccJ9NOo2Cq0q52iJQr1kIkm
79D6l02fvt9TbttQU5lpU9kcpRg/1YS9v1Tpi+5HUtRK/aN0C7A0zHD7uARJwjip
urUmty4xKsVtYNtJX0OUbmJkafZe9Kfb33C6ih0DwfUKuB5B32kzIVugk+DplK8j
02NiWsxh0wrw57N5iO1yLxHwNrMSvy2UI8In8QcPzGulyJDBVw/RpA4NQP4UOzoV
AZku+LvYcw==
-----END CERTIFICATE-----)";


QString const testTLSKey = R"(-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAt4yYjLcEDM7AEgedS8Fk3fUyF6odVBdK8/sLoATzEc6HODXz
ox6T/8On3suynN8HTPcH4QkO1oCWiPrt8OZri/sqSw9cASnQUok6fKQUU76SaWqB
m5fOoMbAekaYpofrv/tQ4IdAP2+nH76GtIGJNzx0e/aKklsuRBtIwqqgvZRjQZQr
qwCG5mPqRBfvViloZfknwjlyS/mfqznZoVpZ1gKzjaCkO6oTksHyY/ucbnRNLObW
S+wu1yhFAjbn8l7eVoTH4IBMx8RXaIoaZBelTkd3etGlVjqE25hmAqmBZVuKJljH
7LY/8uRXIaYcJ2YWaCXfG0hzbSlXPDkDYWrsxQIDAQABAoIBAEL7P8A6GXRDDryF
otU+YfzNudYA8mr5hRS8DGX86GcbIyVUKvDf+8peMCiR1UCB8zwW+f0ZPRzyF/0s
9R/wNlcC9VAm7sBN7gPwqDNL/U8CQJPPljSdlX3+iccVdCdxeoq4v67wLHX53Ncs
xCOjEdviZ+/E7JS0SZH5EvhXJAmKOsKRlz0XU55Ra7N2SToz9j8WhX/do6RcufOY
Pxmdg9A2sm48Ps5dTl2RZvy/yDTt41U2YeXM+EAuYCmBmRyw/dWMB9q3h0a5+Gx7
pkb0SvmUCfaN/WK3wlxJ+rfOp+xJSM5aA9zeSjQpAIif3poEsafENEG9sdpWDVtl
UrDwHQECgYEA4l79b5krz5c0bpRHYN+GnLkBrDOR+ZMiNBR5UVYDUJ769PvdMmbK
fhFBsU5diG/+6GZRbYPrXP1jcXHRHeJMwQFxbwOXU5YCJfoSLxSUyk6u5wGiMC5H
p2+WCd6w6mIe23PzgHlCbccAFDVZzC+R0pYsRj/jMsfuLNtGLz1tyLECgYEAz5LG
abRrrfUsjM3OyywJbAig8bSOV/IrOvIkTxJNQi9sw1rYEjG9wEqnz44H3hSSgY7c
aJuoS8ehHR/FdMpiwP4jWHK7T0Y2Jjrqo07iW8U/YedwPVvJiRKlwic13Px1h/Op
Hp2wdYh4L2Gcbi0krYd/RSFATmOjIJndp6w+alUCgYEAroSnBEtlEES1Am9EXDXX
lKm41WZoqq05GEeUhBU4twXp2cb28C14/RoWuDf/OfmF3utK6ZBjeqxK5yHlIxHd
NIsFRZ3SI3mprFePf0Zxs0pX4vZKcLStPzNyy6coY3pD6dIJr0lM4k8iC3JaCWW/
GUf3WC1W3kZuo5xlDnRgV/ECgYEAnvZLZrYZ5I2nAWm3XVarHIX7Iz9f5y/5NVos
vjVI30/MXksav8xCAZnqq4OcuNFOZVOPrbjPCMGnu9MR91/qgtvdG6Y5lfsyCtMB
z/DgXuFOqd6A0SySyZtzP52hnUvlgijyshSXB1tslvSMxL9joFTs/Xb6dU3Opm/P
FNJOtkUCgYEAjxBLJt2cbE7LjAM2GA5txILUFJBh6Q2U7M922rzr+C7p9oUlx00Z
ZD7atctAQGMqhfBTL1Lxu2gJ3rr5NGVCFDUv1SaYC68nuRkbRD6mHbB8IGpUcmCi
YBADKClADQt7+Nn+ufSsypTpAo9X2gMGQqplSQlmPgoP8wCGbycC5iQ=
-----END RSA PRIVATE KEY-----)";
