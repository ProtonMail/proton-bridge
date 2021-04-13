Feature: IMAP fetch header of message
  Background: Fetch header deterministic content type and boundary
    Given there is connected user "user"
    And there are messages in mailbox "INBOX" for "user"
      | id | from   | to      | subject                          | n attachments | content type | body |
      | 1  | f@m.co | t@pm.me | A message with attachment        | 2             | html         | body |
      | 2  | f@m.co | t@pm.me | A simple html message            | 0             | html         | body |
      | 3  | f@m.co | t@pm.me | A simple plain message           | 0             | plain        | body |
    # | 4  | f@m.co | t@pm.me | An externally encrypted message  | 0             | mixed        | body |
    # | 5  | f@m.co | t@pm.me | A simple plain message in latin1 | 0             | plain-latin1 | body |

    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"

  @ignore-live
  Scenario Outline: Fetch header deterministic content type and boundary
    Given header is not cached for message "<id>" in "INBOX" for "user"
    # First time need to download and cache
    When IMAP client fetches header of "<id>"
    Then IMAP response is "OK"
    And IMAP response contains "Content-Type: <contentType>"
    And IMAP response contains "<parameter>"
    And header is cached for message "<id>" in "INBOX" for "user"
    # Second time it's taken from imap cache
    When IMAP client fetches body "<id>"
    Then IMAP response is "OK"
    And IMAP response contains "Content-Type: <contentType>"
    And IMAP response contains "<parameter>"
    # Third time header taken from DB
    When IMAP client fetches header of "<id>"
    Then IMAP response is "OK"
    And IMAP response contains "Content-Type: <contentType>"
    And IMAP response contains "<parameter>"

    Examples:
        | id | contentType     | parameter                                                                 |
        | 1  | multipart/mixed | boundary=4e07408562bedb8b60ce05c1decfe3ad16b72230967de01f640b7e4729b49fce |
        | 2  | text/html       | charset=utf-8                                                             |
        | 3  | text/plain      | charset=utf-8                                                             |


