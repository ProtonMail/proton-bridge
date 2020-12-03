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

// +build build_qa

package updater

// DefaultPublicKey is the public key used to sign builds.
const DefaultPublicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQENBF9Q55wBCADiwBHGCyJiO2ZSDh9ZPecFKnf+JEryzqGYu3jImEoV2X5Bx/Kl
5n3hHvao9jekEDFr1AjvSKfG9Zz/1GdionUUEdw76mkc7y09GKdXENOyCQYs7CV7
WbWDSGSmp6DVBcRzRzMKm4zuB208a6Wwd2aYqIJ9Oo0l3ypQnox0BQCbbqewYSYN
Dmj+WJkO+e2ovJQWrQgtpnj/QBX18KBjP4FiLSPHAyy7aC2t6JlTIz8UVAw2VZFn
GBUUqnn0iy3W0nJNgv1ouo0rCa+eYBpz3n+GKTFWFDTIPQfZbh15nFJJgBSuiwyM
sHjWCNJYu5PQmwNlGJJjtKw/9xgTFLC9yaNPABEBAAG0BkJyaWRnZYkBTgQTAQgA
OBYhBH3hU445a9yHH+QknbtAQ7nyijPUBQJfUOecAhsDBQsJCAcCBhUKCQgLAgQW
AgMBAh4BAheAAAoJELtAQ7nyijPUpisH/iznWGoma1PXpaQlD2241k9zSzg3Nczn
yfm2mYtXlGVvjGLr29neErWpLy0Kb2ihKTTsgMkwSwcasBap8HYTtENNl1nUzQL7
UhaASTzZ2jYw4Dypps+DYpoLm9RUWKHuUOE5Ov8QPjTBC/BswA0Lv1Z9u9t5qsdp
UgB+YVYgRC+zSHMIzWSMx0dCSPgRilkPvIa5wB77J1+ZE7y1n/uQXOYrKitWrf+w
tXcRYoPqYQ4KXIQ/PMCTwSEDDbsPD7F09AzYQPv6D20d7dyEf0/hlfpj+cvGyBG0
GdGLjwjjKNA99ra1IXjgBUIEv/XpijfKK2D0FDiOdZi3JnVr8OYBCeW5AQ0EX1Dn
nAEIAMtD5sLJ3hXE/bKRQaINx+7hzYhFOxzdGdOTlzlzEjsWYLmy2cWb2fjazIhf
37g8HlSlMaHtHkdJIn1hS9+N76GxEChH31tF6Cuyz+k6TRqroNHsIxzOIjv3+qkM
7xWPRhq8msB8ulWKBQtWpwVVC3sa/qTh9k29wuEiwQY0IxLV0a6BkE1TqK5/7A6Q
o8SMCvQW6wAxPZMhPM/FwxMYxrKUT3UUDmRYS5RvSlMGUwK2HucQVU/qwsOPkJs4
wq6RI+5NDtyGxMxUKod/GYpPaICUI/VNgIZXX6NNzS7JYEYBjtI/JOEOc0yQSh1u
jEGl1k+4OLogUiV02mpGCrHutm0AEQEAAYkBNgQYAQgAIBYhBH3hU445a9yHH+Qk
nbtAQ7nyijPUBQJfUOecAhsMAAoJELtAQ7nyijPU/wUIAKibg4GFxHFSiEjtzdlO
2cIIr3yCsFmGFYVLF3JkOtVvQk7QDZTNsx5ZqC+Mtlf3Z04btG5M/FpHQ097orfl
IH+bZVXMrYtzd4J7ujKGEJU2hY6a9j50odsiwl6CSrXdppS7RGdkhui0RCke/y9Z
wJU5oyiWmcsQfhnET7DEpI7twqEwg43VBGOnaRxKFecyYsQVASlrWMENEpoaup8B
oIS2nDvMVSSK77tmkNcLt8911VqZPtOYmxzM5rc+gm7Pn9kSZUXoGy4p5sFDu/mj
zT1w+Qev2GlSVwFdKPasefLmb3lBEbNeZAkfFl48WEzwtK3VJM60Xl8RPFk0IKLe
tXw=
=aaxG
-----END PGP PUBLIC KEY BLOCK-----`
