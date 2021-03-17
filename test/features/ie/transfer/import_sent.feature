Feature: Import to sent
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Labels/label"
    And there is EML file "Sent/one.eml"
      """
      Subject: one
      From: Foo <foo@example.com>
      To: Bridge Test <bridgetest@pm.test>
      Message-ID: one.integrationtest

      one

      """
    And there is EML file "Sent/two.eml"
      """
      Subject: two
      From: Bar <bar@example.com>
      To: Bridge Test <bridgetest@pm.test>
      Message-ID: two.integrationtest

      two

      """

  Scenario: Import sent only
    When user "user" imports local files
    Then progress result is "OK"
    And transfer exported 2 messages
    And transfer imported 2 messages
    And transfer failed for 0 messages
    And API mailbox "INBOX" for "user" has 0 message
    And API mailbox "Sent" for "user" has messages
      | from            | to                 | subject |
      | foo@example.com | bridgetest@pm.test | one     |
      | bar@example.com | bridgetest@pm.test | two     |

  # Messages imported to label only are added automatically to Archive folder.
  # Then it depends on the order: if the message is first imported to Sent
  # folder and later to that label with importing to Archive, message will not
  # be in Sent but Archive. The order is semi-random for the big messages,
  # e.g., it will do alphabetical order of mailboxes, but for under ten small
  # messages the order is random every time (because we are importing in
  # batches of up to ten messages and iterating through map we use to collect
  # messages is random). So we cannot for this test ensure the same output
  # every time.
  @ignore-live
  Scenario: Import to sent and custom label
    And there is EML file "Label/one.eml"
      """
      Subject: one
      From: Foo <foo@example.com>
      To: Bridge Test <bridgetest@pm.test>
      Message-ID: one.integrationtest

      one

      """
    When user "user" imports local files
    Then progress result is "OK"
    And transfer exported 3 messages
    And transfer imported 3 messages
    And transfer failed for 0 messages
    # We had an issue that moving message to Sent automatically added
    # the message also into Inbox if the message was in some custom label.
    And API mailbox "INBOX" for "user" has 0 message
    And API mailbox "Labels/label" for "user" has messages
      | from            | to                 | subject |
      | foo@example.com | bridgetest@pm.test | one     |
    And API mailbox "Sent" for "user" has messages
      | from            | to                 | subject |
      | foo@example.com | bridgetest@pm.test | one     |
      | bar@example.com | bridgetest@pm.test | two     |

  Scenario: Import to sent and inbox is in both mailboxes
    And there is EML file "Inbox/one.eml"
      """
      Subject: one
      From: Foo <foo@example.com>
      To: Bridge Test <bridgetest@pm.test>
      Message-ID: one.integrationtest
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      one

      """
    When user "user" imports local files
    Then progress result is "OK"
    And transfer exported 3 messages
    And transfer imported 3 messages
    And transfer failed for 0 messages
    And API mailbox "INBOX" for "user" has messages
      | from            | to                 | subject |
      | foo@example.com | bridgetest@pm.test | one     |
    And API mailbox "Sent" for "user" has messages
      | from            | to                 | subject |
      | foo@example.com | bridgetest@pm.test | one     |
      | bar@example.com | bridgetest@pm.test | two     |
