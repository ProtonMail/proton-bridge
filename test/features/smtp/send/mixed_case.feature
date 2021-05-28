Feature: SMTP sending with mixed case address
  Background:
    Given there is connected user "user"
    And there is SMTP client logged in as "user"

  Scenario: Mixed sender case in sender address
    When SMTP client sends message
      """
      From: Bridge Test <[userAddress|capitalize]>
      To: Internal Bridge <bridgetest@protonmail.com>

      hello

      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                        | subject |
      | [userAddress] | bridgetest@protonmail.com |         |
    And message is sent with API call
      """
      {
        "Message": {
          "Subject": "",
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
