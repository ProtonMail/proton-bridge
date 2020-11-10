Feature: SMTP sending with APPENDing to Sent
  Background:
    Given there is connected user "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Sent"
    And there is SMTP client logged in as "user"

  Scenario: Send message and append to Sent
    # First do sending.
    When SMTP client sends message
      """
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Manual send and append
      Message-ID: bridgemessage42

      hello

      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has 1 messages
    And mailbox "Sent" for "user" has messages
      | externalid      | from          | to                        | subject                |
      | bridgemessage42 | [userAddress] | bridgetest@protonmail.com | Manual send and append |
    And message is sent with API call:
      """
      {
        "Message": {
          "Subject": "Manual send and append",
          "ExternalID": "bridgemessage42"
        }
      }
      """

    # Then simulate manual append to Sent mailbox - message should be detected as a duplicate.
    When IMAP client imports message to "Sent"
      """
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Manual send and append
      Message-ID: bridgemessage42

      hello

      """
    Then IMAP response is "OK"
    And mailbox "Sent" for "user" has 1 messages

    # Check that the external ID was not lost in the process.
    When IMAP client sends command "FETCH 1 body.peek[header]"
    Then IMAP response is "OK"
    And IMAP response contains "bridgemessage42"
