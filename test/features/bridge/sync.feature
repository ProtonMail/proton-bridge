Feature: Sync bridge
  Background:
    Given there is connected user "userMoreAddresses"
    And there is "userMoreAddresses" with mailboxes
      | Folders/one  |
      | Folders/two  |
      | Labels/three |
      | Labels/four  |
    And there are messages for "userMoreAddresses" as follows
      | address   | mailboxes            | messages |
      | primary   | INBOX,Folders/one    | 1        |
      | primary   | Archive,Labels/three | 2        |
      | primary   | INBOX                | 1000     |
      | primary   | Archive              | 1200     |
      | primary   | Folders/one          | 1400     |
      | primary   | Folders/two          | 1600     |
      | primary   | Labels/three         | 1800     |
      | primary   | Labels/four          | 2000     |
      | secondary | INBOX                | 100      |
      | secondary | Archive              | 120      |
      | secondary | Folders/one          | 140      |
      | secondary | Folders/two          | 160      |
      | secondary | Labels/three         | 180      |
      | secondary | Labels/four          | 200      |

  # Too heavy for live.
  @ignore-live
  Scenario: Sync in combined mode
    And there is "userMoreAddresses" in "combined" address mode
    When bridge syncs "userMoreAddresses"
    Then bridge response is "OK"
    And "userMoreAddresses" has the following messages
      | mailboxes    | messages |
      | INBOX        | 1101     |
      | Archive      | 1322     |
      | Folders/one  | 1541     |
      | Folders/two  | 1760     |
      | Labels/three | 1982     |
      | Labels/four  | 2200     |

  # Too heavy for live.
  @ignore-live
  Scenario: Sync in split mode
    And there is "userMoreAddresses" in "split" address mode
    When bridge syncs "userMoreAddresses"
    Then bridge response is "OK"
    And "userMoreAddresses" has the following messages
      | address   | mailboxes    | messages |
      | primary   | INBOX        | 1001     |
      | primary   | Archive      | 1202     |
      | primary   | Folders/one  | 1401     |
      | primary   | Folders/two  | 1600     |
      | primary   | Labels/three | 1802     |
      | primary   | Labels/four  | 2000     |
      | secondary | INBOX        | 100      |
      | secondary | Archive      | 120      |
      | secondary | Folders/one  | 140      |
      | secondary | Folders/two  | 160      |
      | secondary | Labels/three | 180      |
      | secondary | Labels/four  | 200      |
