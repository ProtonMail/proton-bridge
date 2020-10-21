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

parser grammar RFC5322Parser;

options { tokenVocab=RFC5322Lexer; }


// -------------------
// 3.2. Lexical tokens
// -------------------

quotedChar: vchar | wsp;

quotedPair
	: Backslash quotedChar
	| obsQP
	;

fws 
	: (wsp* crlf)? wsp+
	| obsFWS
	;

ctext
	: alpha
	| Exclamation
	| DQuote
	| Hash
	| Dollar
	| Percent
	| Ampersand
	| SQuote
	| Asterisk
	| Plus
	| Comma
	| Minus
	| Period
	| Slash
	| Digit
	| Colon
	| Semicolon
	| Less
	| Equal
	| Greater
	| Question
	| At
	| LBracket
	| RBracket
	| Caret
	| Underscore
	| Backtick
	| LCurly
	| Pipe
	| RCurly
	| Tilde
	| obsCtext
	| UTF8NonAscii
	;

ccontent
	: ctext
	| quotedPair
	| comment
	;

comment: LParens (fws? ccontent)* fws? RParens;

cfws
	: (fws? comment)+ fws?
	| fws
	;

atext
	: alpha
	| Digit
	| Exclamation 
	| Hash
	| Dollar 
	| Percent
	| Ampersand 
	| SQuote
	| Asterisk 
	| Plus
	| Minus 
	| Slash
	| Equal 
	| Question
	| Caret 
	| Underscore
	| Backtick 
	| LCurly
	| Pipe 
	| RCurly
	| Tilde
	| UTF8NonAscii
	;

atom: atext+;

// Allow dotAtom to have a trailing period; some messages in the wild look like this.
dotAtom: atext+ (Period atext+)* Period?;

qtext
	: alpha
	| Exclamation
	| Hash
	| Dollar
	| Percent
	| Ampersand
	| SQuote
	| LParens
	| RParens
	| Asterisk
	| Plus
	| Comma
	| Minus
	| Period
	| Slash
	| Digit
	| Colon
	| Semicolon
	| Less
	| Equal
	| Greater
	| Question
	| At
	| LBracket
	| RBracket
	| Caret
	| Underscore
	| Backtick
	| LCurly
	| Pipe
	| RCurly
	| Tilde
	| obsQtext
	| UTF8NonAscii
	;

quotedContent
	: qtext
	| quotedPair
	;

quotedValue: (fws? quotedContent)*;

quotedString: DQuote quotedValue fws? DQuote;

// Allow word to consist of the @ token.
word
	: cfws? encodedWord cfws?
	| cfws? atom cfws?
	| cfws? quotedString cfws?
	| At
	;


// --------------------------------
// 3.3. Date and Time Specification
// --------------------------------

dateTime: (dayOfweek Comma)? day month year hour Colon minute (Colon second)? zone? cfws? EOF;

dayOfweek
	: fws? dayName 
	| cfws? dayName cfws?
	;

dayName
	: M O N 
	| T U E 
	| W E D 
	| T H U 
	| F R I 
	| S A T 
	| S U N
	;

day
	: fws? Digit Digit? fws 
	| cfws? Digit Digit? cfws?
	;

month
	: J A N 
	| F E B 
	| M A R 
	| A P R 
	| M A Y 
	| J U N 
	| J U L 
	| A U G 
	| S E P 
	| O C T 
	| N O V 
	| D E C
	;

year
	: fws Digit Digit Digit Digit fws 
	| cfws? Digit Digit cfws?
	;

// NOTE: RFC5322 requires two digits for the hour, but we 
// relax that requirement a bit, allowing single digits.
hour
	: Digit? Digit 
	| cfws? Digit? Digit cfws?
	;

minute
	: Digit Digit 
	| cfws? Digit Digit cfws?
	;

second
	: Digit Digit 
	| cfws? Digit Digit cfws?
	;

offset: (Plus | Minus)? Digit Digit Digit Digit;

zone
	: fws offset
	| obsZone
	;


// --------------------------
// 3.4. Address Specification
// --------------------------

address
	: mailbox
	| group
	;

mailbox
	: nameAddr
	| addrSpec
	;

nameAddr: displayName? angleAddr;

angleAddr
	: cfws? Less addrSpec? Greater cfws?
	| obsAngleAddr
	;

group: displayName Colon groupList? Semicolon cfws?;

displayName
	: word+
	| word (word | Period | cfws)*
	;

mailboxList
	: mailbox (Comma mailbox)*
	| obsMboxList
	;

addressList
	: address (Comma address)* EOF
	| obsAddrList EOF
	;

groupList
	: mailboxList
	| cfws
	| obsGroupList
	;

// Allow addrSpec contain a port.
addrSpec: localPart At domain (Colon port)?;

localPart
	: cfws? dotAtom cfws?
	| cfws? quotedString cfws?
	| obsLocalPart
	;

port: Digit+;

domain
	: cfws? dotAtom cfws? 
	| cfws? domainLiteral cfws?
	| cfws? obsDomain cfws?
	;

domainLiteral: LBracket (fws? dtext)* fws? RBracket;

dtext
	: alpha
	| Exclamation
	| DQuote
	| Hash
	| Dollar
	| Percent
	| Ampersand
	| SQuote
	| LParens
	| RParens
	| Asterisk
	| Plus
	| Comma
	| Minus
	| Period
	| Slash
	| Digit
	| Colon
	| Semicolon
	| Less
	| Equal
	| Greater
	| Question
	| At
	| Caret
	| Underscore
	| Backtick
	| LCurly
	| Pipe
	| RCurly
	| Tilde
//| obsDtext
	| UTF8NonAscii
	;


// ----------------------------------
// 4.1. Miscellaneous Obsolete Tokens
// ----------------------------------

obsNoWSCTL
	: U_01_08
	| U_0B
	| U_0C
	| U_0E_1F
	| Delete
	;

obsCtext: obsNoWSCTL;

obsQtext: obsNoWSCTL;

obsQP: Backslash (U_00 | obsNoWSCTL | LF | CR);


// ---------------------------------
// 4.2. Obsolete Folding White Space
// ---------------------------------

obsFWS: wsp+ (crlf wsp+);


// ---------------------------
// 4.3. Obsolete Date and Time
// ---------------------------

obsZone
	: U T 
	| U T C
	| G M T
	| E S T
	| E D T
	| C S T
	| C D T
	| M S T
	| M D T
	| P S T
	| P D T
//| obsZoneMilitary
	;


// ------------------------
// 4.4. Obsolete Addressing
// ------------------------

obsAngleAddr: cfws? Less obsRoute addrSpec Greater cfws?;

obsRoute: obsDomainList Colon;

obsDomainList: (cfws | Comma)* At domain (Comma cfws? (At domain)?)*;

obsMboxList: (cfws? Comma)* mailbox (Comma (mailbox | cfws)?)*;

obsAddrList: (cfws? Comma)* address (Comma (address | cfws)?)*;

obsGroupList: (cfws? Comma)+ cfws?;

obsLocalPart: word (Period word)*;

obsDomain: atom (Period atom)*;


// ------------------------------------
// 2. Syntax of encoded-words (RFC2047)
// ------------------------------------

encodedWord: Equal Question charset Question encoding Question encodedText Question Equal;

charset: token;

encoding: token;

token: tokenChar+;

tokenChar
	: alpha
	| Exclamation
	| Hash
	| Dollar
	| Percent
	| Ampersand
	| SQuote
	| Asterisk
	| Plus
	| Minus
	| Digit
	| Backslash
	| Caret
	| Underscore
	| Backtick
	| LCurly
	| Pipe
	| RCurly
	| Tilde
	;

encodedText: encodedChar+;

encodedChar
	: alpha
	| Exclamation
	| DQuote
	| Hash
	| Dollar
	| Percent
	| Ampersand
	| SQuote
	| LParens
	| RParens
	| Asterisk
	| Plus
	| Comma
	| Minus
	| Period
	| Slash
	| Digit
	| Colon
	| Semicolon
	| Less
	| Equal
	| Greater
	| At
	| LBracket
	| Backslash
	| RBracket
	| Caret
	| Underscore
	| Backtick
	| LCurly
	| Pipe
	| RCurly
	| Tilde
	;


// -------------------------
// B.1. Core Rules (RFC5234)
// -------------------------

crlf: CR LF;

wsp: SP | TAB;

vchar
	: alpha
	| Exclamation
	| DQuote
	| Hash
	| Dollar
	| Percent
	| Ampersand
	| SQuote
	| LParens
	| RParens
	| Asterisk
	| Plus
	| Comma
	| Minus
	| Period
	| Slash
	| Digit
	| Colon
	| Semicolon
	| Less
	| Equal
	| Greater
	| Question
	| At
	| LBracket
	| Backslash
	| RBracket
	| Caret
	| Underscore
	| Backtick
	| LCurly
	| Pipe
	| RCurly
	| Tilde
	| UTF8NonAscii
	;

alpha: A | B | C | D | E | F | G | H | I | J | K | L | M | N | O | P | Q | R | S | T | U | V | W | X | Y | Z ;
