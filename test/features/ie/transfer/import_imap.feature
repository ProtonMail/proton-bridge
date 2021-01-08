Feature: Import from IMAP server
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/Foo"
    And there is "user" with mailbox "Folders/Bar"
    And there are IMAP mailboxes
      | name   |
      | Inbox  |
      | Foo    |
      | Broken |
    And there are IMAP messages
      | mailbox | seqnum | uid | from            | to                        | subject | time                |
      | Foo     | 1      | 12  | foo@example.com | bridgetest@protonmail.com | one     | 2020-01-01T12:00:00 |
      | Foo     | 2      | 14  | bar@example.com | bridgetest@protonmail.com | two     | 2020-01-01T13:00:00 |
      | Foo     | 3      | 15  | bar@example.com | bridgetest@protonmail.com | three   | 2020-01-01T12:30:00 |
    And there is IMAP message in mailbox "Inbox" with seq 1, uid 42, time "2020-01-01T12:34:56" and subject "hello"
      """
      Subject: hello
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <test@protonmail.com>
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      hello

      """

  Scenario: Import all
    When user "user" imports remote messages
    Then progress result is "OK"
    And transfer exported 4 messages
    And transfer imported 4 messages
    And transfer failed for 0 messages
    And API mailbox "INBOX" for "user" has messages
      | from               | to                  | subject |
      | bridgetest@pm.test | test@protonmail.com | hello   |
    And API mailbox "Folders/Foo" for "user" has messages
      | from            | to                        | subject |
      | foo@example.com | bridgetest@protonmail.com | one     |
      | bar@example.com | bridgetest@protonmail.com | two     |
      | bar@example.com | bridgetest@protonmail.com | three   |

  Scenario: Import only Foo to Bar with time limit
    When user "user" imports remote messages with rules
      | source | target | from                | to                  |
      | Foo    | Bar    | 2020-01-01T12:10:00 | 2020-01-01T13:00:00 |
    Then progress result is "OK"
    And transfer exported 2 messages
    And transfer imported 2 messages
    And transfer failed for 0 messages
    And API mailbox "Folders/Bar" for "user" has messages
      | from            | to                        | subject |
      | bar@example.com | bridgetest@protonmail.com | two     |
      | bar@example.com | bridgetest@protonmail.com | three   |

  # Note we need to have message which we can parse and use in go-imap
  # but which has problem on our side. Used example with missing boundary
  # is real example which we want to solve one day. Probabl this test
  # can be removed once we import any time of message or switch is to
  # something we will never allow.
  Scenario: Import broken message
    Given there is IMAP message in mailbox "Broken" with seq 1, uid 42, time "2020-01-01T12:34:56" and subject "broken"
      """
      Subject: missing boundary end
      Content-Type: multipart/related; boundary=boundary

      --boundary
      Content-Disposition: inline
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/plain; charset=utf-8

      body

      """
    When user "user" imports remote messages with rules
      | source | target |
      | Broken | Foo    |
    Then progress result is "OK"
    And transfer exported 1 messages
    And transfer imported 0 messages
    And transfer failed for 1 messages
