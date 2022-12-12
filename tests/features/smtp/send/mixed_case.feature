Feature: SMTP sending with mixed case address
  Background:
    Given there exists an account with username "user" and password "password"
    And there exists an account with username "bridgetest" and password "password"
    And bridge starts
    And the user logs in with username "user" and password "password"
    And user "user" connects and authenticates SMTP client "1"

  Scenario: Mixed sender case in sender address
    When SMTP client "1" sends the following message from "user@[domain]" to "bridgetest@[domain]":
      """
      From: Bridge Test <USER@[domain]>
      To: Internal Bridge <bridgetest@[domain]>

      hello

      """
    Then it succeeds
    When user "user" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from          | to                  | subject |
      | user@[domain] | bridgetest@[domain] |         |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "",
          "Sender": {
            "Name": "Bridge Test",
            "Address": "user@[domain]"
          },
          "ToList": [
            {
              "Address": "bridgetest@[domain]",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [],
          "MIMEType": "text/plain"
        }
      }
      """
