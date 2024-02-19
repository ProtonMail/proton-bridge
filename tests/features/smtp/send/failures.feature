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

      iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAACklEQVR4nGMAAQAABQABDQot
      tAAAAABJRU5ErkJggg==
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
    Then it fails with error "invalid return path"
