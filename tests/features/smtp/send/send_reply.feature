Feature: SMTP send reply

  Background:
    Given there exists an account with username "[user:user1]" and password "password"
    And there exists an account with username "[user:user2]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user1]" and password "password"
    And user "[user:user1]" finishes syncing
    And user "[user:user1]" connects and authenticates SMTP client "1"
    And user "[user:user1]" connects and authenticates IMAP client "1"
    Then it succeeds

  @long-black
  Scenario: Reply with In-Reply-To but no References
    # User1 send the initial message.
    When SMTP client "1" sends the following message from "[user:user1]@[domain]" to "[user:user2]@[domain]":
      """
      From: Bridge Test <[user:user1]@[domain]>
      To: Internal Bridge <[user:user2]@[domain]>
      Subject: Please Reply
      Message-ID: <something@protonmail.ch>

      hello

      """
    Then it succeeds
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                  | to                    | subject      | message-id                |
      | [user:user1]@[domain] | [user:user2]@[domain] | Please Reply | <something@protonmail.ch> |
    # login user2.
    And the user logs in with username "[user:user2]" and password "password"
    And user "[user:user2]" connects and authenticates IMAP client "2"
    And user "[user:user2]" connects and authenticates SMTP client "2"
    And user "[user:user2]" finishes syncing
    # User2 receive the message.
    Then IMAP client "2" eventually sees the following messages in "INBOX":
      | from                  |  subject     | message-id                | reply-to              |
      | [user:user1]@[domain] | Please Reply | <something@protonmail.ch> | [user:user1]@[domain] |
    # User2 reply to it.
    When SMTP client "2" sends the following message from "[user:user2]@[domain]" to "[user:user1]@[domain]":
      """
      From: Internal Bridge <[user:user2]@[domain]>
      To: Bridge Test <[user:user1]@[domain]>
      Content-Type: text/plain
      Subject: FW - Please Reply
      In-Reply-To: <something@protonmail.ch>

      Heya

      """
    Then it succeeds
    Then IMAP client "2" eventually sees the following messages in "Sent":
      | from                  | to                    | subject           | in-reply-to               | references                |
      | [user:user2]@[domain] | [user:user1]@[domain] | FW - Please Reply | <something@protonmail.ch> | <something@protonmail.ch> |
    # User1 receive the reply.|
    And IMAP client "1" eventually sees the following messages in "INBOX":
      | from                  | subject           | body  | in-reply-to               | references                |  reply-to             |
      | [user:user2]@[domain] | FW - Please Reply | Heya  | <something@protonmail.ch> | <something@protonmail.ch> | [user:user2]@[domain] |

  @long-black
  Scenario: Reply with References but no In-Reply-To
    # User1 send the initial message.
    When SMTP client "1" sends the following message from "[user:user1]@[domain]" to "[user:user2]@[domain]":
      """
      From: Bridge Test <[user:user1]@[domain]>
      To: Internal Bridge <[user:user2]@[domain]>
      Subject: Please Reply
      Message-ID: <something@protonmail.ch>

      hello

      """
    Then it succeeds
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                  | to                    | subject      | message-id                |
      | [user:user1]@[domain] | [user:user2]@[domain] | Please Reply | <something@protonmail.ch> |
    # login user2.
    And the user logs in with username "[user:user2]" and password "password"
    And user "[user:user2]" connects and authenticates IMAP client "2"
    And user "[user:user2]" connects and authenticates SMTP client "2"
    And user "[user:user2]" finishes syncing
    # User2 receive the message.
    Then IMAP client "2" eventually sees the following messages in "INBOX":
      | from                  |  subject     | message-id                |  reply-to             |
      | [user:user1]@[domain] | Please Reply | <something@protonmail.ch> | [user:user1]@[domain] |
    # User2 reply to it.
    When SMTP client "2" sends the following message from "[user:user2]@[domain]" to "[user:user1]@[domain]":
      """
      From: Internal Bridge <[user:user2]@[domain]>
      To: Bridge Test <[user:user1]@[domain]>
      Content-Type: text/plain
      Subject: FW - Please Reply
      References: <something@protonmail.ch>

      Heya

      """
    Then it succeeds
    Then IMAP client "2" eventually sees the following messages in "Sent":
      | from                  | to                    | subject           | in-reply-to               | references                |
      | [user:user2]@[domain] | [user:user1]@[domain] | FW - Please Reply | <something@protonmail.ch> | <something@protonmail.ch> |
    # User1 receive the reply.|
    And IMAP client "1" eventually sees the following messages in "INBOX":
      | from                  | subject           | body  | in-reply-to               | references                | reply-to              |
      | [user:user2]@[domain] | FW - Please Reply | Heya  | <something@protonmail.ch> | <something@protonmail.ch> | [user:user2]@[domain] |


  @long-black
  Scenario: Reply with both  References and In-Reply-To
    # User1 send the initial message.
    When SMTP client "1" sends the following message from "[user:user1]@[domain]" to "[user:user2]@[domain]":
      """
      From: Bridge Test <[user:user1]@[domain]>
      To: Internal Bridge <[user:user2]@[domain]>
      Subject: Please Reply
      Message-ID: <something@protonmail.ch>

      hello

      """
    Then it succeeds
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                  | to                    | subject      | message-id                |
      | [user:user1]@[domain] | [user:user2]@[domain] | Please Reply | <something@protonmail.ch> |
    # login user2.
    And the user logs in with username "[user:user2]" and password "password"
    And user "[user:user2]" connects and authenticates IMAP client "2"
    And user "[user:user2]" connects and authenticates SMTP client "2"
    And user "[user:user2]" finishes syncing
    # User2 receive the message.
    Then IMAP client "2" eventually sees the following messages in "INBOX":
      | from                  |  subject     | message-id                | reply-to              |
      | [user:user1]@[domain] | Please Reply | <something@protonmail.ch> | [user:user1]@[domain] |
    # User2 reply to it.
    When SMTP client "2" sends the following message from "[user:user2]@[domain]" to "[user:user1]@[domain]":
      """
      From: Internal Bridge <[user:user2]@[domain]>
      To: Bridge Test <[user:user1]@[domain]>
      Content-Type: text/plain
      Subject: FW - Please Reply
      In-Reply-To: <something@protonmail.ch>
      References: <something@protonmail.ch>

      Heya

      """
    Then it succeeds
    Then IMAP client "2" eventually sees the following messages in "Sent":
      | from                  | to                    | subject           | in-reply-to               | references                |
      | [user:user2]@[domain] | [user:user1]@[domain] | FW - Please Reply | <something@protonmail.ch> | <something@protonmail.ch> |
    # User1 receive the reply.|
    And IMAP client "1" eventually sees the following messages in "INBOX":
      | from                  | subject           | body  | in-reply-to               | references                | reply-to              |
      | [user:user2]@[domain] | FW - Please Reply | Heya  | <something@protonmail.ch> | <something@protonmail.ch> | [user:user2]@[domain] |


  @long-black
  Scenario: Reply with In-Reply-To matching several received ExternalID
    # User1 send the initial message.
    When SMTP client "1" sends the following message from "[user:user1]@[domain]" to "[user:user2]@[domain]":
      """
      From: Bridge Test <[user:user1]@[domain]>
      To: Internal Bridge <[user:user2]@[domain]>
      Subject: Please Reply
      Message-ID: <something@external.com>

      hello

      """
    Then it succeeds
    When SMTP client "1" sends the following message from "[user:user1]@[domain]" to "[user:user2]@[domain]":
      """
      From: Bridge Test <[user:user1]@[domain]>
      To: Internal Bridge <[user:user2]@[domain]>
      Subject: Please Reply Again
      Message-ID: <something@external.com>

      hello

      """
    Then it succeeds
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                  | to                    | subject            | message-id               |
      | [user:user1]@[domain] | [user:user2]@[domain] | Please Reply       | <something@external.com> |
      | [user:user1]@[domain] | [user:user2]@[domain] | Please Reply Again | <something@external.com> |
    # login user2.
    And the user logs in with username "[user:user2]" and password "password"
    And user "[user:user2]" connects and authenticates IMAP client "2"
    And user "[user:user2]" connects and authenticates SMTP client "2"
    And user "[user:user2]" finishes syncing
    # User2 receive the message.
    Then IMAP client "2" eventually sees the following messages in "INBOX":
      | from                  |  subject           | message-id              | reply-to               |
      | [user:user1]@[domain] | Please Reply       | <something@external.com> | [user:user1]@[domain] |
      | [user:user1]@[domain] | Please Reply Again | <something@external.com> | [user:user1]@[domain] |
    # User2 reply to it.
    When SMTP client "2" sends the following message from "[user:user2]@[domain]" to "[user:user1]@[domain]":
      """
      From: Internal Bridge <[user:user2]@[domain]>
      To: Bridge Test <[user:user1]@[domain]>
      Content-Type: text/plain
      Subject: FW - Please Reply
      In-Reply-To: <something@external.com>

      Heya

      """
    Then it succeeds
    Then IMAP client "2" eventually sees the following messages in "Sent":
      | from                  | to                    | subject           | in-reply-to | references |
      | [user:user2]@[domain] | [user:user1]@[domain] | FW - Please Reply |             |            |
    # User1 receive the reply.|
    And IMAP client "1" eventually sees the following messages in "INBOX":
      | from                  | subject           | body  | in-reply-to  | references |
      | [user:user2]@[domain] | FW - Please Reply | Heya  |              |            |


  @long-black
  Scenario: Reply with In-Reply-To matching several ExternalID but one sent by us
    # User1 send the initial message.
    When SMTP client "1" sends the following message from "[user:user1]@[domain]" to "[user:user2]@[domain]":
      """
      From: Bridge Test <[user:user1]@[domain]>
      To: Internal Bridge <[user:user2]@[domain]>
      Subject: Please Reply
      Message-ID: <something@external.com>

      hello

      """
    Then it succeeds
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                  | to                    | subject            | message-id               |
      | [user:user1]@[domain] | [user:user2]@[domain] | Please Reply       | <something@external.com> |
    # login user2.
    And the user logs in with username "[user:user2]" and password "password"
    And user "[user:user2]" connects and authenticates IMAP client "2"
    And user "[user:user2]" connects and authenticates SMTP client "2"
    And user "[user:user2]" finishes syncing
    # User2 receive the message.
    Then IMAP client "2" eventually sees the following messages in "INBOX":
      | from                  |  subject           | message-id              | reply-to               |
      | [user:user1]@[domain] | Please Reply       | <something@external.com> | [user:user1]@[domain] |
    # User2 reply to it.
    When SMTP client "2" sends the following message from "[user:user2]@[domain]" to "[user:user1]@[domain]":
      """
      From: Internal Bridge <[user:user2]@[domain]>
      To: Bridge Test <[user:user1]@[domain]>
      Content-Type: text/plain
      Subject: FW - Please Reply
      In-Reply-To: <something@external.com>
      Message-ID: <something@external.com>

      Heya

      """
    Then it succeeds
    When SMTP client "2" sends the following message from "[user:user2]@[domain]" to "[user:user1]@[domain]":
      """
      From: Internal Bridge <[user:user2]@[domain]>
      To: Bridge Test <[user:user1]@[domain]>
      Content-Type: text/plain
      Subject: FW - Please Reply Again
      In-Reply-To: <something@external.com>
      Message-ID: <something@external.com>

      Heya

      """
    Then it succeeds
    Then IMAP client "2" eventually sees the following messages in "Sent":
      | from                  | to                    | subject                 | in-reply-to              | references               |
      | [user:user2]@[domain] | [user:user1]@[domain] | FW - Please Reply       | <something@external.com> | <something@external.com> |
      | [user:user2]@[domain] | [user:user1]@[domain] | FW - Please Reply Again | <something@external.com> | <something@external.com> |
    # User1 receive the reply.|
    And IMAP client "1" eventually sees the following messages in "INBOX":
      | from                  | subject                 | in-reply-to              | references                |
      | [user:user2]@[domain] | FW - Please Reply       | <something@external.com> | <something@external.com>  |
      | [user:user2]@[domain] | FW - Please Reply Again | <something@external.com> | <something@external.com>  |