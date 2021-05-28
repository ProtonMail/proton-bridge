Feature: Servers are closed when no internet

  Scenario: All connection are closed and then restored multiple times
    Given there is connected user "user"
    And there is IMAP client "i1" logged in as "user"
    And there is SMTP client "s1" logged in as "user"
    When there is no internet connection
    And 1 second pass
    Then IMAP client "i1" is logged out
    And SMTP client "s1" is logged out
    Given the internet connection is restored
    And 1 second pass
    And there is IMAP client "i2" logged in as "user"
    And there is SMTP client "s2" logged in as "user"
    When IMAP client "i2" gets info of "INBOX"
    When SMTP client "s2" sends "HELO example.com"
    Then IMAP response to "i2" is "OK"
    Then SMTP response to "s2" is "OK"
    When there is no internet connection
    And 1 second pass
    Then IMAP client "i2" is logged out
    And SMTP client "s2" is logged out
    Given the internet connection is restored
    And 1 second pass
    And there is IMAP client "i3" logged in as "user"
    And there is SMTP client "s3" logged in as "user"
    When IMAP client "i3" gets info of "INBOX"
    When SMTP client "s3" sends "HELO example.com"
    Then IMAP response to "i3" is "OK"
    Then SMTP response to "s3" is "OK"
