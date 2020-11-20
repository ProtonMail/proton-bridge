Feature: IMAP remove messages from mailbox
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/mbox"
    And there is "user" with mailbox "Labels/label"

  @ignore
  Scenario Outline: Mark message as deleted and EXPUNGE
    Given there are 10 messages in mailbox "<mailbox>" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "<mailbox>"
    When IMAP client marks message seq "2" as deleted
    Then IMAP response is "OK"
    And mailbox "<mailbox>" for "user" has 10 messages
    And message "2" in "<mailbox>" for "user" is marked as deleted
    And IMAP response contains "\* 2 FETCH[ (]*FLAGS \([^)]*\\Deleted"
    When IMAP client sends expunge
    Then IMAP response is "OK"
    And IMAP response contains "\* 2 EXPUNGE"
    And mailbox "<mailbox>" for "user" has 9 messages

    Examples:
      | mailbox      |
      | INBOX        |
      | Folders/mbox |
      | Labels/label |
      | Spam         |
      | Trash        |

  @ignore
  Scenario Outline: Mark all messages as deleted and EXPUNGE
    Given there are 5 messages in mailbox "<mailbox>" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "<mailbox>"
    When IMAP client marks message seq "1:*" as deleted
    Then IMAP response is "OK"
    When IMAP client sends expunge
    Then IMAP response is "OK"
    And IMAP response contains "\* 1 EXPUNGE"
    And IMAP response contains "\* 2 EXPUNGE"
    And IMAP response contains "\* 3 EXPUNGE"
    And IMAP response contains "\* 4 EXPUNGE"
    And IMAP response contains "\* 5 EXPUNGE"
    And mailbox "<mailbox>" for "user" has 0 messages

    Examples:
      | mailbox      |
      | INBOX        |
      | Folders/mbox |
      | Labels/label |
      | Spam         |
      | Trash        |

  @ignore
  Scenario Outline: Mark messages as undeleted and EXPUNGE
    Given there are 5 messages in mailbox "<mailbox>" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "<mailbox>"
    When IMAP client marks message seq "1:*" as deleted
    Then IMAP response is "OK"
    When IMAP client marks message seq "1:3" as undeleted
    Then IMAP response is "OK"
    When IMAP client sends expunge
    Then IMAP response is "OK"
    And IMAP response contains "\* 4 EXPUNGE"
    And IMAP response contains "\* 5 EXPUNGE"
    And mailbox "<mailbox>" for "user" has 3 messages

    Examples:
      | mailbox      |
      | INBOX        |
      | Folders/mbox |
      | Labels/label |
      | Spam         |
      | Trash        |

  @ignore
  Scenario Outline: Mark message as deleted and leave mailbox
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client marks message seq "2" as deleted
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has 10 messages
    And message "2" in "INBOX" for "user" is marked as deleted
    When IMAP client sends command "<leave>"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has <n> messages

    Examples:
      | leave         | n  |
      | CLOSE         | 9  |
      | SELECT INBOX  | 9  |
      | SELECT Trash  | 9  |
      | EXAMINE INBOX | 9  |
      | EXAMINE Trash | 9  |
      | LOGOUT        | 9  |
      | UNSELECT      | 10 |

  @ignore
  Scenario: Not possible to delete from All Mail
    Given there are 1 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "All Mail"
    When IMAP client marks message seq "1" as deleted
    Then IMAP response is "IMAP error: NO operation not allowed for 'All Mail' folder"
