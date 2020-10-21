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

// Code generated from RFC5322Parser.g4 by ANTLR 4.8. DO NOT EDIT.

package parser // RFC5322Parser

import "github.com/antlr/antlr4/runtime/Go/antlr"

// BaseRFC5322ParserListener is a complete listener for a parse tree produced by RFC5322Parser.
type BaseRFC5322ParserListener struct{}

var _ RFC5322ParserListener = &BaseRFC5322ParserListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseRFC5322ParserListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseRFC5322ParserListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseRFC5322ParserListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseRFC5322ParserListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterQuotedChar is called when production quotedChar is entered.
func (s *BaseRFC5322ParserListener) EnterQuotedChar(ctx *QuotedCharContext) {}

// ExitQuotedChar is called when production quotedChar is exited.
func (s *BaseRFC5322ParserListener) ExitQuotedChar(ctx *QuotedCharContext) {}

// EnterQuotedPair is called when production quotedPair is entered.
func (s *BaseRFC5322ParserListener) EnterQuotedPair(ctx *QuotedPairContext) {}

// ExitQuotedPair is called when production quotedPair is exited.
func (s *BaseRFC5322ParserListener) ExitQuotedPair(ctx *QuotedPairContext) {}

// EnterFws is called when production fws is entered.
func (s *BaseRFC5322ParserListener) EnterFws(ctx *FwsContext) {}

// ExitFws is called when production fws is exited.
func (s *BaseRFC5322ParserListener) ExitFws(ctx *FwsContext) {}

// EnterCtext is called when production ctext is entered.
func (s *BaseRFC5322ParserListener) EnterCtext(ctx *CtextContext) {}

// ExitCtext is called when production ctext is exited.
func (s *BaseRFC5322ParserListener) ExitCtext(ctx *CtextContext) {}

// EnterCcontent is called when production ccontent is entered.
func (s *BaseRFC5322ParserListener) EnterCcontent(ctx *CcontentContext) {}

// ExitCcontent is called when production ccontent is exited.
func (s *BaseRFC5322ParserListener) ExitCcontent(ctx *CcontentContext) {}

// EnterComment is called when production comment is entered.
func (s *BaseRFC5322ParserListener) EnterComment(ctx *CommentContext) {}

// ExitComment is called when production comment is exited.
func (s *BaseRFC5322ParserListener) ExitComment(ctx *CommentContext) {}

// EnterCfws is called when production cfws is entered.
func (s *BaseRFC5322ParserListener) EnterCfws(ctx *CfwsContext) {}

// ExitCfws is called when production cfws is exited.
func (s *BaseRFC5322ParserListener) ExitCfws(ctx *CfwsContext) {}

// EnterAtext is called when production atext is entered.
func (s *BaseRFC5322ParserListener) EnterAtext(ctx *AtextContext) {}

// ExitAtext is called when production atext is exited.
func (s *BaseRFC5322ParserListener) ExitAtext(ctx *AtextContext) {}

// EnterAtom is called when production atom is entered.
func (s *BaseRFC5322ParserListener) EnterAtom(ctx *AtomContext) {}

// ExitAtom is called when production atom is exited.
func (s *BaseRFC5322ParserListener) ExitAtom(ctx *AtomContext) {}

// EnterDotAtom is called when production dotAtom is entered.
func (s *BaseRFC5322ParserListener) EnterDotAtom(ctx *DotAtomContext) {}

// ExitDotAtom is called when production dotAtom is exited.
func (s *BaseRFC5322ParserListener) ExitDotAtom(ctx *DotAtomContext) {}

// EnterQtext is called when production qtext is entered.
func (s *BaseRFC5322ParserListener) EnterQtext(ctx *QtextContext) {}

// ExitQtext is called when production qtext is exited.
func (s *BaseRFC5322ParserListener) ExitQtext(ctx *QtextContext) {}

// EnterQuotedContent is called when production quotedContent is entered.
func (s *BaseRFC5322ParserListener) EnterQuotedContent(ctx *QuotedContentContext) {}

// ExitQuotedContent is called when production quotedContent is exited.
func (s *BaseRFC5322ParserListener) ExitQuotedContent(ctx *QuotedContentContext) {}

// EnterQuotedValue is called when production quotedValue is entered.
func (s *BaseRFC5322ParserListener) EnterQuotedValue(ctx *QuotedValueContext) {}

// ExitQuotedValue is called when production quotedValue is exited.
func (s *BaseRFC5322ParserListener) ExitQuotedValue(ctx *QuotedValueContext) {}

// EnterQuotedString is called when production quotedString is entered.
func (s *BaseRFC5322ParserListener) EnterQuotedString(ctx *QuotedStringContext) {}

// ExitQuotedString is called when production quotedString is exited.
func (s *BaseRFC5322ParserListener) ExitQuotedString(ctx *QuotedStringContext) {}

// EnterWord is called when production word is entered.
func (s *BaseRFC5322ParserListener) EnterWord(ctx *WordContext) {}

// ExitWord is called when production word is exited.
func (s *BaseRFC5322ParserListener) ExitWord(ctx *WordContext) {}

// EnterDateTime is called when production dateTime is entered.
func (s *BaseRFC5322ParserListener) EnterDateTime(ctx *DateTimeContext) {}

// ExitDateTime is called when production dateTime is exited.
func (s *BaseRFC5322ParserListener) ExitDateTime(ctx *DateTimeContext) {}

// EnterDayOfweek is called when production dayOfweek is entered.
func (s *BaseRFC5322ParserListener) EnterDayOfweek(ctx *DayOfweekContext) {}

// ExitDayOfweek is called when production dayOfweek is exited.
func (s *BaseRFC5322ParserListener) ExitDayOfweek(ctx *DayOfweekContext) {}

// EnterDayName is called when production dayName is entered.
func (s *BaseRFC5322ParserListener) EnterDayName(ctx *DayNameContext) {}

// ExitDayName is called when production dayName is exited.
func (s *BaseRFC5322ParserListener) ExitDayName(ctx *DayNameContext) {}

// EnterDay is called when production day is entered.
func (s *BaseRFC5322ParserListener) EnterDay(ctx *DayContext) {}

// ExitDay is called when production day is exited.
func (s *BaseRFC5322ParserListener) ExitDay(ctx *DayContext) {}

// EnterMonth is called when production month is entered.
func (s *BaseRFC5322ParserListener) EnterMonth(ctx *MonthContext) {}

// ExitMonth is called when production month is exited.
func (s *BaseRFC5322ParserListener) ExitMonth(ctx *MonthContext) {}

// EnterYear is called when production year is entered.
func (s *BaseRFC5322ParserListener) EnterYear(ctx *YearContext) {}

// ExitYear is called when production year is exited.
func (s *BaseRFC5322ParserListener) ExitYear(ctx *YearContext) {}

// EnterHour is called when production hour is entered.
func (s *BaseRFC5322ParserListener) EnterHour(ctx *HourContext) {}

// ExitHour is called when production hour is exited.
func (s *BaseRFC5322ParserListener) ExitHour(ctx *HourContext) {}

// EnterMinute is called when production minute is entered.
func (s *BaseRFC5322ParserListener) EnterMinute(ctx *MinuteContext) {}

// ExitMinute is called when production minute is exited.
func (s *BaseRFC5322ParserListener) ExitMinute(ctx *MinuteContext) {}

// EnterSecond is called when production second is entered.
func (s *BaseRFC5322ParserListener) EnterSecond(ctx *SecondContext) {}

// ExitSecond is called when production second is exited.
func (s *BaseRFC5322ParserListener) ExitSecond(ctx *SecondContext) {}

// EnterOffset is called when production offset is entered.
func (s *BaseRFC5322ParserListener) EnterOffset(ctx *OffsetContext) {}

// ExitOffset is called when production offset is exited.
func (s *BaseRFC5322ParserListener) ExitOffset(ctx *OffsetContext) {}

// EnterZone is called when production zone is entered.
func (s *BaseRFC5322ParserListener) EnterZone(ctx *ZoneContext) {}

// ExitZone is called when production zone is exited.
func (s *BaseRFC5322ParserListener) ExitZone(ctx *ZoneContext) {}

// EnterAddress is called when production address is entered.
func (s *BaseRFC5322ParserListener) EnterAddress(ctx *AddressContext) {}

// ExitAddress is called when production address is exited.
func (s *BaseRFC5322ParserListener) ExitAddress(ctx *AddressContext) {}

// EnterMailbox is called when production mailbox is entered.
func (s *BaseRFC5322ParserListener) EnterMailbox(ctx *MailboxContext) {}

// ExitMailbox is called when production mailbox is exited.
func (s *BaseRFC5322ParserListener) ExitMailbox(ctx *MailboxContext) {}

// EnterNameAddr is called when production nameAddr is entered.
func (s *BaseRFC5322ParserListener) EnterNameAddr(ctx *NameAddrContext) {}

// ExitNameAddr is called when production nameAddr is exited.
func (s *BaseRFC5322ParserListener) ExitNameAddr(ctx *NameAddrContext) {}

// EnterAngleAddr is called when production angleAddr is entered.
func (s *BaseRFC5322ParserListener) EnterAngleAddr(ctx *AngleAddrContext) {}

// ExitAngleAddr is called when production angleAddr is exited.
func (s *BaseRFC5322ParserListener) ExitAngleAddr(ctx *AngleAddrContext) {}

// EnterGroup is called when production group is entered.
func (s *BaseRFC5322ParserListener) EnterGroup(ctx *GroupContext) {}

// ExitGroup is called when production group is exited.
func (s *BaseRFC5322ParserListener) ExitGroup(ctx *GroupContext) {}

// EnterDisplayName is called when production displayName is entered.
func (s *BaseRFC5322ParserListener) EnterDisplayName(ctx *DisplayNameContext) {}

// ExitDisplayName is called when production displayName is exited.
func (s *BaseRFC5322ParserListener) ExitDisplayName(ctx *DisplayNameContext) {}

// EnterMailboxList is called when production mailboxList is entered.
func (s *BaseRFC5322ParserListener) EnterMailboxList(ctx *MailboxListContext) {}

// ExitMailboxList is called when production mailboxList is exited.
func (s *BaseRFC5322ParserListener) ExitMailboxList(ctx *MailboxListContext) {}

// EnterAddressList is called when production addressList is entered.
func (s *BaseRFC5322ParserListener) EnterAddressList(ctx *AddressListContext) {}

// ExitAddressList is called when production addressList is exited.
func (s *BaseRFC5322ParserListener) ExitAddressList(ctx *AddressListContext) {}

// EnterGroupList is called when production groupList is entered.
func (s *BaseRFC5322ParserListener) EnterGroupList(ctx *GroupListContext) {}

// ExitGroupList is called when production groupList is exited.
func (s *BaseRFC5322ParserListener) ExitGroupList(ctx *GroupListContext) {}

// EnterAddrSpec is called when production addrSpec is entered.
func (s *BaseRFC5322ParserListener) EnterAddrSpec(ctx *AddrSpecContext) {}

// ExitAddrSpec is called when production addrSpec is exited.
func (s *BaseRFC5322ParserListener) ExitAddrSpec(ctx *AddrSpecContext) {}

// EnterLocalPart is called when production localPart is entered.
func (s *BaseRFC5322ParserListener) EnterLocalPart(ctx *LocalPartContext) {}

// ExitLocalPart is called when production localPart is exited.
func (s *BaseRFC5322ParserListener) ExitLocalPart(ctx *LocalPartContext) {}

// EnterPort is called when production port is entered.
func (s *BaseRFC5322ParserListener) EnterPort(ctx *PortContext) {}

// ExitPort is called when production port is exited.
func (s *BaseRFC5322ParserListener) ExitPort(ctx *PortContext) {}

// EnterDomain is called when production domain is entered.
func (s *BaseRFC5322ParserListener) EnterDomain(ctx *DomainContext) {}

// ExitDomain is called when production domain is exited.
func (s *BaseRFC5322ParserListener) ExitDomain(ctx *DomainContext) {}

// EnterDomainLiteral is called when production domainLiteral is entered.
func (s *BaseRFC5322ParserListener) EnterDomainLiteral(ctx *DomainLiteralContext) {}

// ExitDomainLiteral is called when production domainLiteral is exited.
func (s *BaseRFC5322ParserListener) ExitDomainLiteral(ctx *DomainLiteralContext) {}

// EnterDtext is called when production dtext is entered.
func (s *BaseRFC5322ParserListener) EnterDtext(ctx *DtextContext) {}

// ExitDtext is called when production dtext is exited.
func (s *BaseRFC5322ParserListener) ExitDtext(ctx *DtextContext) {}

// EnterObsNoWSCTL is called when production obsNoWSCTL is entered.
func (s *BaseRFC5322ParserListener) EnterObsNoWSCTL(ctx *ObsNoWSCTLContext) {}

// ExitObsNoWSCTL is called when production obsNoWSCTL is exited.
func (s *BaseRFC5322ParserListener) ExitObsNoWSCTL(ctx *ObsNoWSCTLContext) {}

// EnterObsCtext is called when production obsCtext is entered.
func (s *BaseRFC5322ParserListener) EnterObsCtext(ctx *ObsCtextContext) {}

// ExitObsCtext is called when production obsCtext is exited.
func (s *BaseRFC5322ParserListener) ExitObsCtext(ctx *ObsCtextContext) {}

// EnterObsQtext is called when production obsQtext is entered.
func (s *BaseRFC5322ParserListener) EnterObsQtext(ctx *ObsQtextContext) {}

// ExitObsQtext is called when production obsQtext is exited.
func (s *BaseRFC5322ParserListener) ExitObsQtext(ctx *ObsQtextContext) {}

// EnterObsQP is called when production obsQP is entered.
func (s *BaseRFC5322ParserListener) EnterObsQP(ctx *ObsQPContext) {}

// ExitObsQP is called when production obsQP is exited.
func (s *BaseRFC5322ParserListener) ExitObsQP(ctx *ObsQPContext) {}

// EnterObsFWS is called when production obsFWS is entered.
func (s *BaseRFC5322ParserListener) EnterObsFWS(ctx *ObsFWSContext) {}

// ExitObsFWS is called when production obsFWS is exited.
func (s *BaseRFC5322ParserListener) ExitObsFWS(ctx *ObsFWSContext) {}

// EnterObsZone is called when production obsZone is entered.
func (s *BaseRFC5322ParserListener) EnterObsZone(ctx *ObsZoneContext) {}

// ExitObsZone is called when production obsZone is exited.
func (s *BaseRFC5322ParserListener) ExitObsZone(ctx *ObsZoneContext) {}

// EnterObsAngleAddr is called when production obsAngleAddr is entered.
func (s *BaseRFC5322ParserListener) EnterObsAngleAddr(ctx *ObsAngleAddrContext) {}

// ExitObsAngleAddr is called when production obsAngleAddr is exited.
func (s *BaseRFC5322ParserListener) ExitObsAngleAddr(ctx *ObsAngleAddrContext) {}

// EnterObsRoute is called when production obsRoute is entered.
func (s *BaseRFC5322ParserListener) EnterObsRoute(ctx *ObsRouteContext) {}

// ExitObsRoute is called when production obsRoute is exited.
func (s *BaseRFC5322ParserListener) ExitObsRoute(ctx *ObsRouteContext) {}

// EnterObsDomainList is called when production obsDomainList is entered.
func (s *BaseRFC5322ParserListener) EnterObsDomainList(ctx *ObsDomainListContext) {}

// ExitObsDomainList is called when production obsDomainList is exited.
func (s *BaseRFC5322ParserListener) ExitObsDomainList(ctx *ObsDomainListContext) {}

// EnterObsMboxList is called when production obsMboxList is entered.
func (s *BaseRFC5322ParserListener) EnterObsMboxList(ctx *ObsMboxListContext) {}

// ExitObsMboxList is called when production obsMboxList is exited.
func (s *BaseRFC5322ParserListener) ExitObsMboxList(ctx *ObsMboxListContext) {}

// EnterObsAddrList is called when production obsAddrList is entered.
func (s *BaseRFC5322ParserListener) EnterObsAddrList(ctx *ObsAddrListContext) {}

// ExitObsAddrList is called when production obsAddrList is exited.
func (s *BaseRFC5322ParserListener) ExitObsAddrList(ctx *ObsAddrListContext) {}

// EnterObsGroupList is called when production obsGroupList is entered.
func (s *BaseRFC5322ParserListener) EnterObsGroupList(ctx *ObsGroupListContext) {}

// ExitObsGroupList is called when production obsGroupList is exited.
func (s *BaseRFC5322ParserListener) ExitObsGroupList(ctx *ObsGroupListContext) {}

// EnterObsLocalPart is called when production obsLocalPart is entered.
func (s *BaseRFC5322ParserListener) EnterObsLocalPart(ctx *ObsLocalPartContext) {}

// ExitObsLocalPart is called when production obsLocalPart is exited.
func (s *BaseRFC5322ParserListener) ExitObsLocalPart(ctx *ObsLocalPartContext) {}

// EnterObsDomain is called when production obsDomain is entered.
func (s *BaseRFC5322ParserListener) EnterObsDomain(ctx *ObsDomainContext) {}

// ExitObsDomain is called when production obsDomain is exited.
func (s *BaseRFC5322ParserListener) ExitObsDomain(ctx *ObsDomainContext) {}

// EnterEncodedWord is called when production encodedWord is entered.
func (s *BaseRFC5322ParserListener) EnterEncodedWord(ctx *EncodedWordContext) {}

// ExitEncodedWord is called when production encodedWord is exited.
func (s *BaseRFC5322ParserListener) ExitEncodedWord(ctx *EncodedWordContext) {}

// EnterCharset is called when production charset is entered.
func (s *BaseRFC5322ParserListener) EnterCharset(ctx *CharsetContext) {}

// ExitCharset is called when production charset is exited.
func (s *BaseRFC5322ParserListener) ExitCharset(ctx *CharsetContext) {}

// EnterEncoding is called when production encoding is entered.
func (s *BaseRFC5322ParserListener) EnterEncoding(ctx *EncodingContext) {}

// ExitEncoding is called when production encoding is exited.
func (s *BaseRFC5322ParserListener) ExitEncoding(ctx *EncodingContext) {}

// EnterToken is called when production token is entered.
func (s *BaseRFC5322ParserListener) EnterToken(ctx *TokenContext) {}

// ExitToken is called when production token is exited.
func (s *BaseRFC5322ParserListener) ExitToken(ctx *TokenContext) {}

// EnterTokenChar is called when production tokenChar is entered.
func (s *BaseRFC5322ParserListener) EnterTokenChar(ctx *TokenCharContext) {}

// ExitTokenChar is called when production tokenChar is exited.
func (s *BaseRFC5322ParserListener) ExitTokenChar(ctx *TokenCharContext) {}

// EnterEncodedText is called when production encodedText is entered.
func (s *BaseRFC5322ParserListener) EnterEncodedText(ctx *EncodedTextContext) {}

// ExitEncodedText is called when production encodedText is exited.
func (s *BaseRFC5322ParserListener) ExitEncodedText(ctx *EncodedTextContext) {}

// EnterEncodedChar is called when production encodedChar is entered.
func (s *BaseRFC5322ParserListener) EnterEncodedChar(ctx *EncodedCharContext) {}

// ExitEncodedChar is called when production encodedChar is exited.
func (s *BaseRFC5322ParserListener) ExitEncodedChar(ctx *EncodedCharContext) {}

// EnterCrlf is called when production crlf is entered.
func (s *BaseRFC5322ParserListener) EnterCrlf(ctx *CrlfContext) {}

// ExitCrlf is called when production crlf is exited.
func (s *BaseRFC5322ParserListener) ExitCrlf(ctx *CrlfContext) {}

// EnterWsp is called when production wsp is entered.
func (s *BaseRFC5322ParserListener) EnterWsp(ctx *WspContext) {}

// ExitWsp is called when production wsp is exited.
func (s *BaseRFC5322ParserListener) ExitWsp(ctx *WspContext) {}

// EnterVchar is called when production vchar is entered.
func (s *BaseRFC5322ParserListener) EnterVchar(ctx *VcharContext) {}

// ExitVchar is called when production vchar is exited.
func (s *BaseRFC5322ParserListener) ExitVchar(ctx *VcharContext) {}

// EnterAlpha is called when production alpha is entered.
func (s *BaseRFC5322ParserListener) EnterAlpha(ctx *AlphaContext) {}

// ExitAlpha is called when production alpha is exited.
func (s *BaseRFC5322ParserListener) ExitAlpha(ctx *AlphaContext) {}
