# IMAP clients can move message to local folder (setting \Deleted flag)
# and then move it back (IMAP client does not remember the message,
# so instead removing the flag it imports duplicate message).
# Regular IMAP server would keep the message twice and later EXPUNGE would
# not delete the message (EXPUNGE would delete the original message and
# the new duplicate one would stay). Both Bridge and API detects duplicates;
# therefore we need to remove \Deleted flag if IMAP client re-imports.
Feature: IMAP move message out to and back from local folder
  Background:
    Given there is connected user "user"
    Given there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"

  Scenario: Mark message as deleted and re-append again
    When IMAP client imports message to "INBOX"
      """
      From: <john.doe@mail.com>
      To: <user@pm.me>
      Subject: foo
      Date: Mon, 02 Jan 2006 15:04:05 +0000
      Message-Id: <msgID>

      hello
      """
    Then IMAP response is "OK"
    When IMAP client marks message seq "1" as deleted
    Then IMAP response is "OK"
    When IMAP client imports message to "INBOX"
      """
      From: <john.doe@mail.com>
      To: <user@pm.me>
      Subject: foo
      Date: Mon, 02 Jan 2006 15:04:05 +0000
      Message-Id: <msgID>

      hello
      """
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has 1 message
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject | deleted |
      | john.doe@mail.com | user@pm.me | foo     | false   |

  # We cannot control ID generation on API.
  @ignore-live
  Scenario: Mark internal message as deleted and re-append again
    # Each message has different subject so if the ID generations on fake API
    # changes, test will fail because not even external ID mechanism will work.
    When IMAP client imports message to "INBOX"
      """
      From: <john.doe@mail.com>
      To: <user@pm.me>
      Subject: foo
      Date: Mon, 02 Jan 2006 15:04:05 +0000

      hello
      """
    Then IMAP response is "OK"
    When IMAP client marks message seq "1" as deleted
    Then IMAP response is "OK"
    # Fake API generates for the first message simple ID 1.
    When IMAP client imports message to "INBOX"
      """
      From: <john.doe@mail.com>
      To: <user@pm.me>
      Subject: bar
      Date: Mon, 02 Jan 2006 15:04:05 +0000
      X-Pm-Internal-Id: 1

      hello
      """
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has 1 message
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject | deleted |
      | john.doe@mail.com | user@pm.me | foo     | false   |
