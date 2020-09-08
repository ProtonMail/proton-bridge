Feature: SMTP with bcc
  Background:
    Given there is connected user "user"
    And there is SMTP client logged in as "user"

  Scenario: Send message to address in to and bcc
    When SMTP client sends message with bcc "bridgetest2@protonmail.com"
      """
      Subject: hello
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>

      hello

      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                        | subject |
      | [userAddress] | bridgetest@protonmail.com | hello   |
    And message is sent with API call:
      """
      {
        "Message": {
          "Subject": "hello",
          "ToList": [
            {
              "Address": "bridgetest@protonmail.com",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [
            {
              "Address": "bridgetest2@protonmail.com"
            }
          ]
        }
      }
      """


  Scenario: Send message only to bcc
    When SMTP client sends message with bcc "bridgetest@protonmail.com"
      """
      Subject: hello
      From: Bridge Test <bridgetest@pm.test>

      hello

      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to | subject |
      | [userAddress] |    | hello   |
    And message is sent with API call:
      """
      {
        "Message": {
          "Subject": "hello",
          "ToList": [],
          "CCList": [],
          "BCCList": [
            {
              "Address": "bridgetest@protonmail.com"
            }
          ]
        }
      }
      """
