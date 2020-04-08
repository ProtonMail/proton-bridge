Feature: IMAP create messages
  Background:
    Given there is connected user "userMoreAddresses"
    And there is IMAP client logged in as "userMoreAddresses"

  Scenario: Creates message to user's primary address
    Given there is IMAP client selected in "INBOX"
    When IMAP client creates message "foo" from "john.doe@email.com" to address "primary" of "userMoreAddresses" with body "hello world" in "INBOX"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "userMoreAddresses" has messages
      | from               | to        | subject | read |
      | john.doe@email.com | [primary] | foo     | true |

  Scenario: Creates draft
    When IMAP client creates message "foo" from address "primary" of "userMoreAddresses" to "john.doe@email.com" with body "hello world" in "Drafts"
    Then IMAP response is "OK"
    And mailbox "Drafts" for "userMoreAddresses" has messages
      | from      | to                 | subject | read |
      | [primary] | john.doe@email.com | foo     | true |

  @ignore
  Scenario: Creates message sent from user's primary address
    Given there is IMAP client selected in "Sent"
    When IMAP client creates message "foo" from address "primary" of "userMoreAddresses" to "john.doe@email.com" with body "hello world" in "Sent"
    Then IMAP response is "OK"
    When the event loop of "userMoreAddresses" loops once
    Then mailbox "Sent" for "userMoreAddresses" has messages
      | from      | to                 | subject | read |
      | [primary] | john.doe@email.com | foo     | true |
    And mailbox "INBOX" for "userMoreAddresses" has no messages

  @ignore
  Scenario: Creates message sent from user's secondary address
    Given there is IMAP client selected in "Sent"
    When IMAP client creates message "foo" from address "secondary" of "userMoreAddresses" to "john.doe@email.com" with body "hello world" in "Sent"
    Then IMAP response is "OK"
    When the event loop of "userMoreAddresses" loops once
    Then mailbox "Sent" for "userMoreAddresses" has messages
      | from      | to                 | subject | read |
      | [secondary] | john.doe@email.com | foo     | true |
    And mailbox "INBOX" for "userMoreAddresses" has no messages

  Scenario: Imports an unrelated message to inbox
    Given there is IMAP client selected in "INBOX"
    When IMAP client creates message "foo" from "john.doe@email.com" to "john.doe2@email.com" with body "hello world" in "INBOX"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "userMoreAddresses" has messages
      | from               | to                  | subject | read |
      | john.doe@email.com | john.doe2@email.com | foo     | true |

  Scenario: Imports an unrelated message to sent
    Given there is IMAP client selected in "Sent"
    When IMAP client creates message "foo" from "notuser@gmail.com" to "alsonotuser@gmail.com" with body "hello world" in "Sent"
    Then IMAP response is "OK"
    When the event loop of "userMoreAddresses" loops once
    Then mailbox "Sent" for "userMoreAddresses" has messages
      | from              | to                    | subject | read |
      | notuser@gmail.com | alsonotuser@gmail.com | foo     | true |
    And mailbox "INBOX" for "userMoreAddresses" has no messages
