Feature: SMTP sending of plain messages with attachments
  Background:
    Given there is connected user "user"
    And there is SMTP client logged in as "user"

  Scenario: Basic message with attachment to internal account
    When SMTP client sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Plain with attachment
      Content-Type: multipart/related; boundary=bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: inline
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/plain; charset=utf-8

      This is body of mail with attachment

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
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                        | subject               |
      | [userAddress] | bridgetest@protonmail.com | Plain with attachment |
    And message is sent with API call
      """
      {
        "Message": {
          "Subject": "Plain with attachment",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "bridgetest@protonmail.com",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [],
          "MIMEType": "text/plain"
        }
      }
      """

  Scenario: Plain message with attachment to external account
    When SMTP client sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: External Bridge <pm.bridge.qa@gmail.com>
      Subject: Plain with attachment external
      Content-Type: multipart/related; boundary=bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: inline
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/plain; charset=utf-8

      This is body of mail with attachment

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
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                     | subject                        |
      | [userAddress] | pm.bridge.qa@gmail.com | Plain with attachment external |
    And message is sent with API call
      """
      {
        "Message": {
          "Subject": "Plain with attachment external",
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
          "MIMEType": "text/plain"
        }
      }
      """

  Scenario: Plain message with attachment to two external accounts
    When SMTP client sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: External Bridge 1 <pm.bridge.qa@gmail.com>
      CC: External Bridge 2 <bridgeqa@seznam.cz>
      Subject: Plain with attachment external PGP and external CC
      Content-Type: multipart/mixed; boundary=bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: inline
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/plain; charset=utf-8

      This is body of mail with attachment

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
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                     | cc                 | subject                                            |
      | [userAddress] | pm.bridge.qa@gmail.com | bridgeqa@seznam.cz | Plain with attachment external PGP and external CC |
    And message is sent with API call
      """
      {
        "Message": {
          "Subject": "Plain with attachment external PGP and external CC",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "pm.bridge.qa@gmail.com",
              "Name": "External Bridge 1"
            }
          ],
          "CCList": [
            {
              "Address": "bridgeqa@seznam.cz",
              "Name": "External Bridge 2"
            }
          ],
          "BCCList": [],
          "MIMEType": "text/plain"
        }
      }
      """
