Feature: IMAP list mailboxes
  Background:
    Given there is connected user "user"

  Scenario: List mailboxes
    Given there is "user" with mailbox "Folders/mbox1"
    And there is "user" with mailbox "Labels/mbox2"
    And there is IMAP client logged in as "user"
    When IMAP client lists mailboxes
    Then IMAP response contains "INBOX"
    Then IMAP response contains "Sent"
    Then IMAP response contains "Archive"
    Then IMAP response contains "Trash"
    Then IMAP response contains "All Mail"
    Then IMAP response contains "Folders/mbox1"
    Then IMAP response contains "Labels/mbox2"

  Scenario: List mailboxes without All Mail
    Given there is IMAP client logged in as "user"
    When IMAP client lists mailboxes
    Then IMAP response contains "INBOX"
    Then IMAP response contains "Sent"
    Then IMAP response contains "Archive"
    Then IMAP response contains "Trash"
    Then IMAP response contains "All Mail"
    When All mail mailbox is hidden
    And IMAP client lists mailboxes
    Then IMAP response contains "INBOX"
    Then IMAP response contains "Sent"
    Then IMAP response contains "Archive"
    Then IMAP response contains "Trash"
    Then IMAP response doesn't contain "All Mail"
    When All mail mailbox is visible
    And IMAP client lists mailboxes
    Then IMAP response contains "INBOX"
    Then IMAP response contains "Sent"
    Then IMAP response contains "Archive"
    Then IMAP response contains "Trash"
    Then IMAP response contains "All Mail"

  Scenario: List multiple times in parallel without crash
    Given there is "user" with mailboxes
      | Folders/mbox1  |
      | Folders/mbox2  |
      | Folders/mbox3  |
      | Folders/mbox4  |
      | Folders/mbox5  |
      | Folders/mbox6  |
      | Folders/mbox7  |
      | Folders/mbox8  |
      | Folders/mbox9  |
      | Folders/mbox10 |
      | Folders/mbox11 |
      | Folders/mbox12 |
      | Folders/mbox13 |
      | Folders/mbox14 |
      | Folders/mbox15 |
      | Folders/mbox16 |
      | Folders/mbox17 |
      | Folders/mbox18 |
      | Folders/mbox19 |
      | Folders/mbox20 |
      | Labels/lab1  |
      | Labels/lab2  |
      | Labels/lab3  |
      | Labels/lab4  |
      | Labels/lab5  |
      | Labels/lab6  |
      | Labels/lab7  |
      | Labels/lab8  |
      | Labels/lab9  |
      | Labels/lab10 |
      | Labels/lab11 |
      | Labels/lab12 |
      | Labels/lab13 |
      | Labels/lab14 |
      | Labels/lab15 |
      | Labels/lab16 |
      | Labels/lab17 |
      | Labels/lab18 |
      | Labels/lab19 |
      | Labels/lab20 |
      | Labels/lab1.1  |
      | Labels/lab1.2  |
      | Labels/lab1.3  |
      | Labels/lab1.4  |
      | Labels/lab1.5  |
      | Labels/lab1.6  |
      | Labels/lab1.7  |
      | Labels/lab1.8  |
      | Labels/lab1.9  |
      | Labels/lab1.10 |
      | Labels/lab1.11 |
      | Labels/lab1.12 |
      | Labels/lab1.13 |
      | Labels/lab1.14 |
      | Labels/lab1.15 |
      | Labels/lab1.16 |
      | Labels/lab1.17 |
      | Labels/lab1.18 |
      | Labels/lab1.19 |
      | Labels/lab1.20 |
      | Labels/lab2.1  |
      | Labels/lab2.2  |
      | Labels/lab2.3  |
      | Labels/lab2.4  |
      | Labels/lab2.5  |
      | Labels/lab2.6  |
      | Labels/lab2.7  |
      | Labels/lab2.8  |
      | Labels/lab2.9  |
      | Labels/lab2.10 |
      | Labels/lab2.11 |
      | Labels/lab2.12 |
      | Labels/lab2.13 |
      | Labels/lab2.14 |
      | Labels/lab2.15 |
      | Labels/lab2.16 |
      | Labels/lab2.17 |
      | Labels/lab2.18 |
      | Labels/lab2.19 |
      | Labels/lab2.20 |
    And there is IMAP client "A" logged in as "user"
    And there is IMAP client "B" logged in as "user"
    When  IMAP client "A" lists mailboxes
    And  IMAP client "B" lists mailboxes
    Then IMAP response to "A" is "OK"
    And IMAP response to "A" contains "mbox1"
    And IMAP response to "A" contains "mbox10"
    And IMAP response to "A" contains "mbox20"

  @ignore-live
  Scenario: List mailboxes with subfolders
    # Escaped slash in the name contains slash in the name.
    # Not-escaped slash in the name means tree structure.
    # We keep escaping in an IMAP communication so each mailbox is unique and
    # both mailboxes are accessible. The slash is visible in the IMAP client.
    Given there is "user" with mailbox "Folders/a\/b"
    And there is "user" with mailbox "Folders/a/b"
    And there is IMAP client logged in as "user"
    When IMAP client lists mailboxes
    Then IMAP response contains "Folders/a\/b"
    Then IMAP response contains "Folders/a/b"
