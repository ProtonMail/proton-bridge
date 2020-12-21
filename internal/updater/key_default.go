// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package updater

// DefaultPublicKey is the public key used to sign builds.
const DefaultPublicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQGNBF/ZKfIBDADbmJNFJvifih7rk6rMGtotwS6UTJh9Lo38dQ8gAetAqOqtdoDf
8CEcRG9LB0LRpKP0jJlBe1QhQ+3iGFMPb0mnBo1EGW5NTjUbZ0wWnLq80Z5Vat9Z
sFuPGxao8GTRLNNghG9UXlQirNKNAgJe3OOWKuYJ24mxZqFd53nG6AUmpXxc+bx9
4zc/OcXhnm8cNE5L0kzIdqD9i1KwRYa+8zqh0YT5zbH06Fl9sBOBBFb+uJm9ICA0
7HEpfHRYwJMiDXfX9qpHG+aqRj2wPmbkVTBHd3iLdCtaPG/OB4eglQM8ow+yZcRa
j6mP5yTnUELux8tDgtFRsYAsXFTsiYYnXr+7i76suhrYj9RZ9SlLmCOZRn0ulbX9
ZY3M1vUrWdjYVr0O11KV1llIXdUJd+ypsWmzye7AkzkjK1YO+zujkrxf3/kVPyBJ
7clLT90lvu+FsZZPTe4avq3QUyMT9hBthhnOE3slKYTGO+m9phCGVRA0ZMSaaova
snDf7AJOpErCjS8AEQEAAbQyUHJvdG9uTWFpbCBCcmlkZ2UgTGF1bmNoZXIgPGxl
cG9udEBwcm90b25tYWlsLmNvbT6JAc4EEwEIADgWIQTZQfj0vJ3zJbvXPgwWJ/pj
8zl8GQUCX9kp8gIbAwULCQgHAgYVCgkICwIEFgIDAQIeAQIXgAAKCRAWJ/pj8zl8
GUK1DACwnT6r8DiO2IHt4gJ9rVHpAGauRbN5BTxDrA1ywW0rr1TLGO4GBkrkpcXz
0RTTulDEt6YDgtGFiCnUJqeBSiY7ssT7XxEEodG8DU27cH1C922tFwM6H4DECPtR
NrGYbEHFNmZk64ncBs7ER4+sNv7fMbdUj/x25XEGQEzVuG4tiw0H00QfTpkRFJi5
EHT/AC6+MMKC6fuNrTeD4eE488EEEPImCEZ855k3bWQooScqyAKQr30aJuOiqKrY
/qR9lTldOyi5rMKRK29HmKgkmVG+WGnU+lBPbbeMXDx7Um07rUB84bPVJDmLKLWL
33IW/nfuKw0w0znDCKhyUYR1wlY7xwwXq0+INd/XgmWCsLGSSmYqDv3UYIdx4Uk2
EripTjLJ+/g0BvplHwzTWsJxrxpc3d5sObkxy3c61+mPu4jnRixl+duUWko71c4I
ClnkLb2E1GjHMTNjiX9hkft96+xqfPIaTjuadC81fSqF3/7BlLkDE60C3hYMC/EV
Xrq3+Oq5AY0EX9kp8gEMAJi+geuZlkVtQ4JQvR0qKqkIbFUv0Uwhcm8/Dk7dn4Am
oIxCOK9ZCFxSnEGqMSzNAjfS7JSjKfmzdHUcv9T1axTjFSjRg/rzX48DN9pUghdS
uleMcA6fJHPOqNDE15oKgvnfcN2jtQdwDsvek2iMxRw7koyd1Rd+twrSyCjvJk3Y
Zayfx0FPDZmfmToHcI4NAqDbRrzpDaePMKrcvaqonh/Fn9O/t0il9y7xH9orAbO6
AMDZgFtDMU38D0zZ9zD2MFj76BdFc4/GtBUB0NHI7PzLHo3+yumLOlEcnH/fQfDz
FNsiT8k+1ONmh23vAztXvOY2Wy5ZDs6mfLsTfUvda9kvLL3LOReFpNTFp/fVnUgr
al4jFodfTJyPsZ5wNqq5KtZAwq6t2BSguCsatrLqzVLNnYVgYAPLuSv5c7DW93zX
8k6kfb3NWkjIRvFKf2SOz/hpHjfl6CVRDHncNF4Le51ppbCl5DBx4MFV0WIirpj9
5B/aBE1exQTNWl8Q7faRlQARAQABiQG2BBgBCAAgFiEE2UH49Lyd8yW71z4MFif6
Y/M5fBkFAl/ZKfICGwwACgkQFif6Y/M5fBkHGgwAkGKmmjObUKVYM8lcHK+etCro
3OBX8Sxv6Yv7IQr3X6GpmNMJT1Ryk8PfFZv2mEA6NECHtWi6iytLTxcTgKeZqjuj
4r5WUwLabLkO2Pb8T372YiDbXHHhlBFdxUcAG4ERwO/QkZkVugOgotTSGXauEhn3
SNQTxV4vGbZq+0Aug5ibTuwvUQ5H147rJraQ+XwAgBs0AzE/iOxl3WSJEWyV3iJJ
bFL6ndkPRz46hIkEKfMqQKo7lgaHWKK3yo9OgIG4nnGLstMxCIwASKYnCDgdsPGf
xMm20Lc/U6loUxed+935OW7ig8+POETQQr0PKq0tfm4wuo3cqhJ+rQ4BFB9Z/Te9
/6PAHitFB2Mnlqx9FNZOmFXlVMt6xzsw8zsT+hjbPVQtl8FLz+tbm60sDL3EIS3D
zaN0U4LnOmSGRWNQo8DYNO+kzVsI0f1H3d62j1CfO2gIfjJ7qrPC8V4OlKcHjyjT
rV62sVvSmYuBLVLSQr6JQowPmwZ/urn/8LR4E851
=lAy0
-----END PGP PUBLIC KEY BLOCK-----`
