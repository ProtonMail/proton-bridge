Feature: IMAP move messages by append and delete (without MOVE support, e.g., Outlook)
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name | type   |
      | mbox | folder |
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "source"
    And user "[user:user]" connects and authenticates IMAP client "target"
    Then it succeeds

  Scenario Outline: Move message from <srcMailbox> to <dstMailbox> by <order>
    When IMAP client "source" appends the following message to "<srcMailbox>":
      """
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000
      From: sndr1@[domain]
      Date: 01 Jan 1980 00:00:00 +0000
      To: rcvr1@[domain]
      Subject: subj1

      body1
      """
    Then it succeeds
    When IMAP client "source" appends the following message to "<srcMailbox>":
      """
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000
      From: sndr2@[domain]
      Date: 01 Jan 1980 00:00:00 +0000
      To: rcvr2@[domain]
      Subject: subj2

      body2
      """
    Then it succeeds
    And IMAP client "source" selects "<srcMailbox>"
    And IMAP client "target" selects "<dstMailbox>"
    When IMAP clients "source" and "target" move message with subject "subj2" of "[user:user]" to "<dstMailbox>" by <order>
    And IMAP client "source" eventually sees 1 messages in "<srcMailbox>"
    And IMAP client "source" eventually sees the following messages in "<srcMailbox>":
      | from           | to             | subject |
      | sndr1@[domain] | rcvr1@[domain] | subj1   |
    And IMAP client "target" eventually sees 1 messages in "<dstMailbox>"
    And IMAP client "target" eventually sees the following messages in "<dstMailbox>":
      | from           | to             | subject |
      | sndr2@[domain] | rcvr2@[domain] | subj2   |
    Examples:
      | srcMailbox | dstMailbox   | order                 |
      | Trash      | INBOX        | APPEND DELETE EXPUNGE |
      | Spam       | INBOX        | APPEND DELETE EXPUNGE |
      | INBOX      | Archive      | APPEND DELETE EXPUNGE |
      | INBOX      | Folders/mbox | APPEND DELETE EXPUNGE |
      | INBOX      | Spam         | APPEND DELETE EXPUNGE |
      | INBOX      | Trash        | APPEND DELETE EXPUNGE |
      | Trash      | INBOX        | DELETE APPEND EXPUNGE |
      | Spam       | INBOX        | DELETE APPEND EXPUNGE |
      | INBOX      | Archive      | DELETE APPEND EXPUNGE |
      | INBOX      | Folders/mbox | DELETE APPEND EXPUNGE |
      | INBOX      | Spam         | DELETE APPEND EXPUNGE |
      | INBOX      | Trash        | DELETE APPEND EXPUNGE |
      | Trash      | INBOX        | DELETE EXPUNGE APPEND |
      | Spam       | INBOX        | DELETE EXPUNGE APPEND |
      | INBOX      | Archive      | DELETE EXPUNGE APPEND |
      | INBOX      | Folders/mbox | DELETE EXPUNGE APPEND |
      | INBOX      | Spam         | DELETE EXPUNGE APPEND |
      | INBOX      | Trash        | DELETE EXPUNGE APPEND |
