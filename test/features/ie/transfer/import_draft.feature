Feature: Import from EML files
  Background:
    Given there is connected user "user"

  Scenario: Import draft without from fallbacks to primary address
    Given there is EML file "Drafts/one.eml"
      """
      Subject: no from yet
      To: Internal Bridge <test@protonmail.com>
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      hello

      """
    When user "user" imports local files with rules
      | source | target |
      | Drafts | Drafts |
    Then progress result is "OK"
    And transfer exported 1 messages
    And transfer imported 1 messages
    And transfer failed for 0 messages
    And API mailbox "Drafts" for "user" has messages
      | from          | to                  | subject     |
      | [userAddress] | test@protonmail.com | no from yet |
