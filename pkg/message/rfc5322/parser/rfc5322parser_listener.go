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

// RFC5322ParserListener is a complete listener for a parse tree produced by RFC5322Parser.
type RFC5322ParserListener interface {
	antlr.ParseTreeListener

	// EnterQuotedChar is called when entering the quotedChar production.
	EnterQuotedChar(c *QuotedCharContext)

	// EnterQuotedPair is called when entering the quotedPair production.
	EnterQuotedPair(c *QuotedPairContext)

	// EnterFws is called when entering the fws production.
	EnterFws(c *FwsContext)

	// EnterCtext is called when entering the ctext production.
	EnterCtext(c *CtextContext)

	// EnterCcontent is called when entering the ccontent production.
	EnterCcontent(c *CcontentContext)

	// EnterComment is called when entering the comment production.
	EnterComment(c *CommentContext)

	// EnterCfws is called when entering the cfws production.
	EnterCfws(c *CfwsContext)

	// EnterAtext is called when entering the atext production.
	EnterAtext(c *AtextContext)

	// EnterAtom is called when entering the atom production.
	EnterAtom(c *AtomContext)

	// EnterDotAtom is called when entering the dotAtom production.
	EnterDotAtom(c *DotAtomContext)

	// EnterQtext is called when entering the qtext production.
	EnterQtext(c *QtextContext)

	// EnterQuotedContent is called when entering the quotedContent production.
	EnterQuotedContent(c *QuotedContentContext)

	// EnterQuotedValue is called when entering the quotedValue production.
	EnterQuotedValue(c *QuotedValueContext)

	// EnterQuotedString is called when entering the quotedString production.
	EnterQuotedString(c *QuotedStringContext)

	// EnterWord is called when entering the word production.
	EnterWord(c *WordContext)

	// EnterDateTime is called when entering the dateTime production.
	EnterDateTime(c *DateTimeContext)

	// EnterDayOfweek is called when entering the dayOfweek production.
	EnterDayOfweek(c *DayOfweekContext)

	// EnterDayName is called when entering the dayName production.
	EnterDayName(c *DayNameContext)

	// EnterDay is called when entering the day production.
	EnterDay(c *DayContext)

	// EnterMonth is called when entering the month production.
	EnterMonth(c *MonthContext)

	// EnterYear is called when entering the year production.
	EnterYear(c *YearContext)

	// EnterHour is called when entering the hour production.
	EnterHour(c *HourContext)

	// EnterMinute is called when entering the minute production.
	EnterMinute(c *MinuteContext)

	// EnterSecond is called when entering the second production.
	EnterSecond(c *SecondContext)

	// EnterOffset is called when entering the offset production.
	EnterOffset(c *OffsetContext)

	// EnterZone is called when entering the zone production.
	EnterZone(c *ZoneContext)

	// EnterAddress is called when entering the address production.
	EnterAddress(c *AddressContext)

	// EnterMailbox is called when entering the mailbox production.
	EnterMailbox(c *MailboxContext)

	// EnterNameAddr is called when entering the nameAddr production.
	EnterNameAddr(c *NameAddrContext)

	// EnterAngleAddr is called when entering the angleAddr production.
	EnterAngleAddr(c *AngleAddrContext)

	// EnterGroup is called when entering the group production.
	EnterGroup(c *GroupContext)

	// EnterDisplayName is called when entering the displayName production.
	EnterDisplayName(c *DisplayNameContext)

	// EnterMailboxList is called when entering the mailboxList production.
	EnterMailboxList(c *MailboxListContext)

	// EnterAddressList is called when entering the addressList production.
	EnterAddressList(c *AddressListContext)

	// EnterGroupList is called when entering the groupList production.
	EnterGroupList(c *GroupListContext)

	// EnterAddrSpec is called when entering the addrSpec production.
	EnterAddrSpec(c *AddrSpecContext)

	// EnterLocalPart is called when entering the localPart production.
	EnterLocalPart(c *LocalPartContext)

	// EnterPort is called when entering the port production.
	EnterPort(c *PortContext)

	// EnterDomain is called when entering the domain production.
	EnterDomain(c *DomainContext)

	// EnterDomainLiteral is called when entering the domainLiteral production.
	EnterDomainLiteral(c *DomainLiteralContext)

	// EnterDtext is called when entering the dtext production.
	EnterDtext(c *DtextContext)

	// EnterObsNoWSCTL is called when entering the obsNoWSCTL production.
	EnterObsNoWSCTL(c *ObsNoWSCTLContext)

	// EnterObsCtext is called when entering the obsCtext production.
	EnterObsCtext(c *ObsCtextContext)

	// EnterObsQtext is called when entering the obsQtext production.
	EnterObsQtext(c *ObsQtextContext)

	// EnterObsQP is called when entering the obsQP production.
	EnterObsQP(c *ObsQPContext)

	// EnterObsFWS is called when entering the obsFWS production.
	EnterObsFWS(c *ObsFWSContext)

	// EnterObsZone is called when entering the obsZone production.
	EnterObsZone(c *ObsZoneContext)

	// EnterObsAngleAddr is called when entering the obsAngleAddr production.
	EnterObsAngleAddr(c *ObsAngleAddrContext)

	// EnterObsRoute is called when entering the obsRoute production.
	EnterObsRoute(c *ObsRouteContext)

	// EnterObsDomainList is called when entering the obsDomainList production.
	EnterObsDomainList(c *ObsDomainListContext)

	// EnterObsMboxList is called when entering the obsMboxList production.
	EnterObsMboxList(c *ObsMboxListContext)

	// EnterObsAddrList is called when entering the obsAddrList production.
	EnterObsAddrList(c *ObsAddrListContext)

	// EnterObsGroupList is called when entering the obsGroupList production.
	EnterObsGroupList(c *ObsGroupListContext)

	// EnterObsLocalPart is called when entering the obsLocalPart production.
	EnterObsLocalPart(c *ObsLocalPartContext)

	// EnterObsDomain is called when entering the obsDomain production.
	EnterObsDomain(c *ObsDomainContext)

	// EnterEncodedWord is called when entering the encodedWord production.
	EnterEncodedWord(c *EncodedWordContext)

	// EnterCharset is called when entering the charset production.
	EnterCharset(c *CharsetContext)

	// EnterEncoding is called when entering the encoding production.
	EnterEncoding(c *EncodingContext)

	// EnterToken is called when entering the token production.
	EnterToken(c *TokenContext)

	// EnterTokenChar is called when entering the tokenChar production.
	EnterTokenChar(c *TokenCharContext)

	// EnterEncodedText is called when entering the encodedText production.
	EnterEncodedText(c *EncodedTextContext)

	// EnterEncodedChar is called when entering the encodedChar production.
	EnterEncodedChar(c *EncodedCharContext)

	// EnterCrlf is called when entering the crlf production.
	EnterCrlf(c *CrlfContext)

	// EnterWsp is called when entering the wsp production.
	EnterWsp(c *WspContext)

	// EnterVchar is called when entering the vchar production.
	EnterVchar(c *VcharContext)

	// EnterAlpha is called when entering the alpha production.
	EnterAlpha(c *AlphaContext)

	// ExitQuotedChar is called when exiting the quotedChar production.
	ExitQuotedChar(c *QuotedCharContext)

	// ExitQuotedPair is called when exiting the quotedPair production.
	ExitQuotedPair(c *QuotedPairContext)

	// ExitFws is called when exiting the fws production.
	ExitFws(c *FwsContext)

	// ExitCtext is called when exiting the ctext production.
	ExitCtext(c *CtextContext)

	// ExitCcontent is called when exiting the ccontent production.
	ExitCcontent(c *CcontentContext)

	// ExitComment is called when exiting the comment production.
	ExitComment(c *CommentContext)

	// ExitCfws is called when exiting the cfws production.
	ExitCfws(c *CfwsContext)

	// ExitAtext is called when exiting the atext production.
	ExitAtext(c *AtextContext)

	// ExitAtom is called when exiting the atom production.
	ExitAtom(c *AtomContext)

	// ExitDotAtom is called when exiting the dotAtom production.
	ExitDotAtom(c *DotAtomContext)

	// ExitQtext is called when exiting the qtext production.
	ExitQtext(c *QtextContext)

	// ExitQuotedContent is called when exiting the quotedContent production.
	ExitQuotedContent(c *QuotedContentContext)

	// ExitQuotedValue is called when exiting the quotedValue production.
	ExitQuotedValue(c *QuotedValueContext)

	// ExitQuotedString is called when exiting the quotedString production.
	ExitQuotedString(c *QuotedStringContext)

	// ExitWord is called when exiting the word production.
	ExitWord(c *WordContext)

	// ExitDateTime is called when exiting the dateTime production.
	ExitDateTime(c *DateTimeContext)

	// ExitDayOfweek is called when exiting the dayOfweek production.
	ExitDayOfweek(c *DayOfweekContext)

	// ExitDayName is called when exiting the dayName production.
	ExitDayName(c *DayNameContext)

	// ExitDay is called when exiting the day production.
	ExitDay(c *DayContext)

	// ExitMonth is called when exiting the month production.
	ExitMonth(c *MonthContext)

	// ExitYear is called when exiting the year production.
	ExitYear(c *YearContext)

	// ExitHour is called when exiting the hour production.
	ExitHour(c *HourContext)

	// ExitMinute is called when exiting the minute production.
	ExitMinute(c *MinuteContext)

	// ExitSecond is called when exiting the second production.
	ExitSecond(c *SecondContext)

	// ExitOffset is called when exiting the offset production.
	ExitOffset(c *OffsetContext)

	// ExitZone is called when exiting the zone production.
	ExitZone(c *ZoneContext)

	// ExitAddress is called when exiting the address production.
	ExitAddress(c *AddressContext)

	// ExitMailbox is called when exiting the mailbox production.
	ExitMailbox(c *MailboxContext)

	// ExitNameAddr is called when exiting the nameAddr production.
	ExitNameAddr(c *NameAddrContext)

	// ExitAngleAddr is called when exiting the angleAddr production.
	ExitAngleAddr(c *AngleAddrContext)

	// ExitGroup is called when exiting the group production.
	ExitGroup(c *GroupContext)

	// ExitDisplayName is called when exiting the displayName production.
	ExitDisplayName(c *DisplayNameContext)

	// ExitMailboxList is called when exiting the mailboxList production.
	ExitMailboxList(c *MailboxListContext)

	// ExitAddressList is called when exiting the addressList production.
	ExitAddressList(c *AddressListContext)

	// ExitGroupList is called when exiting the groupList production.
	ExitGroupList(c *GroupListContext)

	// ExitAddrSpec is called when exiting the addrSpec production.
	ExitAddrSpec(c *AddrSpecContext)

	// ExitLocalPart is called when exiting the localPart production.
	ExitLocalPart(c *LocalPartContext)

	// ExitPort is called when exiting the port production.
	ExitPort(c *PortContext)

	// ExitDomain is called when exiting the domain production.
	ExitDomain(c *DomainContext)

	// ExitDomainLiteral is called when exiting the domainLiteral production.
	ExitDomainLiteral(c *DomainLiteralContext)

	// ExitDtext is called when exiting the dtext production.
	ExitDtext(c *DtextContext)

	// ExitObsNoWSCTL is called when exiting the obsNoWSCTL production.
	ExitObsNoWSCTL(c *ObsNoWSCTLContext)

	// ExitObsCtext is called when exiting the obsCtext production.
	ExitObsCtext(c *ObsCtextContext)

	// ExitObsQtext is called when exiting the obsQtext production.
	ExitObsQtext(c *ObsQtextContext)

	// ExitObsQP is called when exiting the obsQP production.
	ExitObsQP(c *ObsQPContext)

	// ExitObsFWS is called when exiting the obsFWS production.
	ExitObsFWS(c *ObsFWSContext)

	// ExitObsZone is called when exiting the obsZone production.
	ExitObsZone(c *ObsZoneContext)

	// ExitObsAngleAddr is called when exiting the obsAngleAddr production.
	ExitObsAngleAddr(c *ObsAngleAddrContext)

	// ExitObsRoute is called when exiting the obsRoute production.
	ExitObsRoute(c *ObsRouteContext)

	// ExitObsDomainList is called when exiting the obsDomainList production.
	ExitObsDomainList(c *ObsDomainListContext)

	// ExitObsMboxList is called when exiting the obsMboxList production.
	ExitObsMboxList(c *ObsMboxListContext)

	// ExitObsAddrList is called when exiting the obsAddrList production.
	ExitObsAddrList(c *ObsAddrListContext)

	// ExitObsGroupList is called when exiting the obsGroupList production.
	ExitObsGroupList(c *ObsGroupListContext)

	// ExitObsLocalPart is called when exiting the obsLocalPart production.
	ExitObsLocalPart(c *ObsLocalPartContext)

	// ExitObsDomain is called when exiting the obsDomain production.
	ExitObsDomain(c *ObsDomainContext)

	// ExitEncodedWord is called when exiting the encodedWord production.
	ExitEncodedWord(c *EncodedWordContext)

	// ExitCharset is called when exiting the charset production.
	ExitCharset(c *CharsetContext)

	// ExitEncoding is called when exiting the encoding production.
	ExitEncoding(c *EncodingContext)

	// ExitToken is called when exiting the token production.
	ExitToken(c *TokenContext)

	// ExitTokenChar is called when exiting the tokenChar production.
	ExitTokenChar(c *TokenCharContext)

	// ExitEncodedText is called when exiting the encodedText production.
	ExitEncodedText(c *EncodedTextContext)

	// ExitEncodedChar is called when exiting the encodedChar production.
	ExitEncodedChar(c *EncodedCharContext)

	// ExitCrlf is called when exiting the crlf production.
	ExitCrlf(c *CrlfContext)

	// ExitWsp is called when exiting the wsp production.
	ExitWsp(c *WspContext)

	// ExitVchar is called when exiting the vchar production.
	ExitVchar(c *VcharContext)

	// ExitAlpha is called when exiting the alpha production.
	ExitAlpha(c *AlphaContext)
}
