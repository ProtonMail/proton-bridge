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

lexer grammar RFC5322Lexer;

U_00:             '\u0000';
U_01_08:          '\u0001'..'\u0008';
TAB:              '\t';     // \u0009
LF:               '\n';     // \u000A
U_0B:             '\u000B';
U_0C:             '\u000C';
CR:               '\r';     // \u000D
U_0E_1F:          '\u000E'..'\u001F';

// Printable (0x20-0x7E)
SP:               ' ';      // \u0020
Exclamation:      '!';      // \u0021
DQuote:           '"';      // \u0022
Hash:             '#';      // \u0023
Dollar:           '$';      // \u0024
Percent:          '%';      // \u0025
Ampersand:        '&';      // \u0026
SQuote:           '\'';     // \u0027
LParens:          '(';      // \u0028
RParens:          ')';      // \u0029
Asterisk:         '*';      // \u002A
Plus:             '+';      // \u002B
Comma:            ',';      // \u002C
Minus:            '-';      // \u002D
Period:           '.';      // \u002E
Slash:            '/';      // \u002F
Digit:            [0-9];    // \u0030 -- \u0039
Colon:            ':';      // \u003A
Semicolon:        ';';      // \u003B
Less:             '<';      // \u003C
Equal:            '=';      // \u003D
Greater:          '>';      // \u003E
Question:         '?';      // \u003F
At:               '@';      // \u0040
// alphaUpper
LBracket:         '[';      // \u005B
Backslash:        '\\';     // \u005C
RBracket:         ']';      // \u005D
Caret:            '^';      // \u005E
Underscore:       '_';      // \u005F
Backtick:         '`';      // \u0060
// alphaLower
LCurly:           '{';      // \u007B
Pipe:             '|';      // \u007C
RCurly:           '}';      // \u007D
Tilde:            '~';      // \u007E

// Other
Delete: '\u007F';

// RFC6532 Extension
UTF8NonAscii: '\u0080'..'\uFFFF';

A: 'A'|'a';
B: 'B'|'b';
C: 'C'|'c';
D: 'D'|'d';
E: 'E'|'e';
F: 'F'|'f';
G: 'G'|'g';
H: 'H'|'h';
I: 'I'|'i';
J: 'J'|'j';
K: 'K'|'k';
L: 'L'|'l';
M: 'M'|'m';
N: 'N'|'n';
O: 'O'|'o';
P: 'P'|'p';
Q: 'Q'|'q';
R: 'R'|'r';
S: 'S'|'s';
T: 'T'|'t';
U: 'U'|'u';
V: 'V'|'v';
W: 'W'|'w';
X: 'X'|'x';
Y: 'Y'|'y';
Z: 'Z'|'z';
