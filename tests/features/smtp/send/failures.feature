Feature: SMTP wrong messages
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And there exists an account with username "[user:to]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" connects and authenticates SMTP client "1"
    Then it succeeds

  Scenario: Message with attachment and wrong boundaries
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: Internal Bridge <[user:to]@[domain]>
      Subject: With attachment (wrong boundaries)
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
      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606


      """
    Then it fails

  Scenario: Invalid from
    When SMTP client "1" sends the following message from "unowned@[domain]" to "[user:to]@[domain]":
      """
      From: Bridge Test <unowned@[domain]>
      To: Internal Bridge <[user:to]@[domain]>

      hello

      """
    Then it fails
