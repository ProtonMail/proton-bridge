Feature: SMTP sending of plain messages
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And there exists an account with username "[user:to]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" connects and authenticates SMTP client "1"
    Then it succeeds

  Scenario: HTML message with attachment to internal account
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: Internal Bridge <[user:to]@[domain]>
      Subject: HTML with attachment internal
      Content-Type: multipart/related; boundary=bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: inline
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/html; charset=utf-8

      <html><body>This is body of <b>HTML mail</b> with attachment<body></html>

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: attachment; filename=outline-light-instagram-48.png
      Content-Id: <9114fe6f0adfaf7fdf7a@protonmail.com>
      Content-Transfer-Encoding: base64
      Content-Type: image/png

      iVBORw0KGgoAAAANSUhEUgAAADAAAAAwBAMAAAClLOS0AAAALVBMVEUAAAD/////////////////
      //////////////////////////////////////+hSKubAAAADnRSTlMAgO8QQM+/IJ9gj1AwcIQd
      OXUAAAGdSURBVDjLXJC9SgNBFIVPXDURTYhgIQghINgowyLYCAYtRFAIgtYhpAjYhC0srCRW6YIg
      WNpoHVSsg/gEii+Qnfxq4DyDc3cyMfrBwl2+O+fOHTi8p7LS5RUf/9gpMKL7iT9sK47Q95ggpkzv
      1cvRcsGYNMYsmP+zKN27NR2vcDyTNVdfkOuuniNPMWafvIbljt+YoMEvW8y7lt+ARwhvrgPjhA0I
      BTng7S1GLPlypBvtIBPidY4YBDJFdtnkscQ5JGaGqxC9i7jSDwcwnB8qHWBaQjw1ABI8wYgtVoG6
      9pFkH8iZIiJeulFt4JLvJq8I5N2GMWYbHWDWzM3JZTMdeSWla0kW86FcuI0mfStiNKQ/AhEeh8h0
      YUTffFwrMTT5oSwdojIQ0UKcocgAKRH1HiqhFQmmJa5qRaYHNbRiSsOgslY0NdixItUTUWlZkedP
      HXVyAgAIA1F0wP5btQZPIyTwvAqa/Fl4oacuP+e4XHAjSYpkQkxSiMX+T7FPoZJToSStzED70HCy
      KE3NGCg4jJrC6Ti7AFwZLhnW0gMbzFZc0RmmeAAAAABJRU5ErkJggg==
      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject                       |
      | [user:user]@[domain] | [user:to]@[domain] | HTML with attachment internal |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "HTML with attachment internal",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "[user:to]@[domain]",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [],
          "MIMEType": "text/html"
        }
      }
      """

  Scenario: HTML message with attachment to external account
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "pm.bridge.qa@gmail.com":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: External Bridge <pm.bridge.qa@gmail.com>
      Subject: HTML with attachment external PGP
      Content-Type: multipart/mixed; boundary=bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: inline
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/html; charset=utf-8

      <html><body>This is body of <b>HTML mail</b> with attachment<body></html>

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: attachment; filename=outline-light-instagram-48.png
      Content-Id: <9114fe6f0adfaf7fdf7a@protonmail.com>
      Content-Transfer-Encoding: base64
      Content-Type: image/png

      iVBORw0KGgoAAAANSUhEUgAAADAAAAAwBAMAAAClLOS0AAAALVBMVEUAAAD/////////////////
      //////////////////////////////////////+hSKubAAAADnRSTlMAgO8QQM+/IJ9gj1AwcIQd
      OXUAAAGdSURBVDjLXJC9SgNBFIVPXDURTYhgIQghINgowyLYCAYtRFAIgtYhpAjYhC0srCRW6YIg
      WNpoHVSsg/gEii+Qnfxq4DyDc3cyMfrBwl2+O+fOHTi8p7LS5RUf/9gpMKL7iT9sK47Q95ggpkzv
      1cvRcsGYNMYsmP+zKN27NR2vcDyTNVdfkOuuniNPMWafvIbljt+YoMEvW8y7lt+ARwhvrgPjhA0I
      BTng7S1GLPlypBvtIBPidY4YBDJFdtnkscQ5JGaGqxC9i7jSDwcwnB8qHWBaQjw1ABI8wYgtVoG6
      9pFkH8iZIiJeulFt4JLvJq8I5N2GMWYbHWDWzM3JZTMdeSWla0kW86FcuI0mfStiNKQ/AhEeh8h0
      YUTffFwrMTT5oSwdojIQ0UKcocgAKRH1HiqhFQmmJa5qRaYHNbRiSsOgslY0NdixItUTUWlZkedP
      HXVyAgAIA1F0wP5btQZPIyTwvAqa/Fl4oacuP+e4XHAjSYpkQkxSiMX+T7FPoZJToSStzED70HCy
      KE3NGCg4jJrC6Ti7AFwZLhnW0gMbzFZc0RmmeAAAAABJRU5ErkJggg==
      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                     | subject                           |
      | [user:user]@[domain] | pm.bridge.qa@gmail.com | HTML with attachment external PGP |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "HTML with attachment external PGP",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "pm.bridge.qa@gmail.com",
              "Name": "External Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [],
          "MIMEType": "text/html"
        }
      }
      """

  Scenario: Alternative plain and HTML message with rfc822 attachment
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "pm.bridge.qa@gmail.com":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: External Bridge <pm.bridge.qa@gmail.com>
      Subject: Alternative plain and HTML with rfc822 attachment
      Content-Type: multipart/mixed; boundary=main-parts

      This is a multipart message in MIME format

      --main-parts
      Content-Type: multipart/alternative; boundary=alternatives

      --alternatives
      Content-Type: text/plain

      There is an attachment


      --alternatives
      Content-Type: text/html

      <html><body>There <b>is</b> an attachment<body></html>


      --alternatives--

      --main-parts
      Content-Type: message/rfc822
      Content-Transfer-Encoding: 7bit
      Content-Disposition: attachment

      Received: from mx1.opensuse.org (mx1.infra.opensuse.org [192.168.47.95]) by
      mailman3.infra.opensuse.org (Postfix) with ESMTP id 38BE2AC3 for
      <factory@lists.opensuse.org>; Sun, 11 Jul 2021 19:50:34 +0000 (UTC)
      From: "Bob " <Bob@something.net>
      Sender: "Bob" <Bob@gmail.com>
      To: "opensuse-factory" <opensuse-factory@opensuse.org>
      Cc: "Bob" <Bob@something.net>
      References:  <y6ZUV5yEyOVQHETZRmi1GFe-Xumzct7QcLpGoSsi1MefGaoovfrUqdkmQ5gM6uySZ7JPIJhDkPJFDqHS1fb_mQ==@protonmail.internalid>
      Subject: VirtualBox problems with kernel 5.13
      Date: Sun, 11 Jul 2021 21:50:25 +0200
      Message-ID: <71672e5f-24a2-c79f-03cc-4c923eb1790b@lwfinger.net>
      MIME-Version: 1.0
      Content-Type: text/plain; charset="utf-8"
      Content-Transfer-Encoding: quoted-printable
      X-Mailer: Microsoft Outlook 16.0
      List-Unsubscribe: <mailto:factory-leave@lists.opensuse.org>
      Content-Language: en-us
      List-Help: <mailto:factory-request@lists.opensuse.org?subject=help>
      List-Subscribe: <mailto:factory-join@lists.opensuse.org>
      Thread-Index: AQFWvbNSAqFOch49YPlLU4eJWPObaQK2iKDq

      I am writing this message as openSUSE's maintainer of VirtualBox.

      Nearly every update of the Linux kernel to a new 5.X version breaks =
      VirtualBox.

      Bob

      --main-parts--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                     | subject                                           |
      | [user:user]@[domain] | pm.bridge.qa@gmail.com | Alternative plain and HTML with rfc822 attachment |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "Alternative plain and HTML with rfc822 attachment",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "pm.bridge.qa@gmail.com",
              "Name": "External Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [],
          "MIMEType": "text/html"
        }
      }
      """
