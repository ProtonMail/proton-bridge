@regression
Feature: SMTP sending of HTMl messages to Internal recipient
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And there exists an account with username "[user:to]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates SMTP client "1"
    Then it succeeds
  
  Scenario: HTML message with Foreign/Nonascii chars in Subject and Body to Internal
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      From: <[user:user]@[domain]>
      To: <[user:to]@[domain]>
      Subject: =?UTF-8?B?U3Vias61zq3Pgs+EIMK2IMOEIMOI?=
      Content-Type: text/html; charset=UTF-8
      Content-Transfer-Encoding: 8bit

      <html>
        <head>

          <meta http-equiv="content-type" content="text/html; charset=UTF-8">
        </head>
        <body>
          Subjεέςτ ¶ Ä È asd
        </body>
      </html>
      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject        |
      | [user:user]@[domain] | [user:to]@[domain] | Subjεέςτ ¶ Ä È |
    And IMAP client "1" eventually sees 1 messages in "Sent"
    When the user logs in with username "[user:to]" and password "password"
    And user "[user:to]" connects and authenticates IMAP client "2"
    And user "[user:to]" finishes syncing
    And it succeeds
    Then IMAP client "2" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "[user:user]@[domain]",
        "to": "[user:to]@[domain]",
        "subject": "Subjεέςτ ¶ Ä È",
        "content": {
          "content-type": "text/html",
          "content-type-charset": "utf-8",
          "transfer-encoding": "quoted-printable",
          "body-is": "<html>\r\n  <head>\r\n\r\n    <meta http-equiv=\"content-type\" content=\"text/html; charset=UTF-8\">\r\n  </head>\r\n  <body>\r\n    Subjεέςτ ¶ Ä È asd\r\n  </body>\r\n</html>"
        }
      }
      """

  Scenario: HTML message with numbering/ordering in Body to Internal
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      Content-Type: multipart/alternative;
        boundary="------------oYnsP1x8lKf6V060046qa0DG"
      MIME-Version: 1.0
      User-Agent: Mozilla Thunderbird
      Content-Language: en-GB
      To: <[user:to]@[domain]>
      From: <[user:user]@[domain]>
      Subject: Message with Numbering/Ordering in Body

      This is a multi-part message in MIME format.
      --------------oYnsP1x8lKf6V060046qa0DG
      Content-Type: text/plain; charset=UTF-8; format=flowed
      Content-Transfer-Encoding: 7bit

      Unordered list

        * Bullet point 1
        * Bullet point 2
            o Bullet point 2.1
            o Bullet point 2.2
                + Bullet point 2.2.1
            o Bullet point 2.3
        * Bullet point 3
            o Bullet point 3.1


      Ordered list

      1. Number 1
          1. Number 1.1
              1. Number 1.1.1
              2. Number 1.1.2
          2. Number 1.2
      2. Number 2
      3. Number 3
          1. Number 3.1
          2. Number 3.2
              1. Number 3.2.1
          3. Number 3.3
          4. Number 3.4
      4. Number 4

      End

      --------------oYnsP1x8lKf6V060046qa0DG
      Content-Type: text/html; charset=UTF-8
      Content-Transfer-Encoding: 7bit

      <!DOCTYPE html>
      <html>
        <head>

          <meta http-equiv="content-type" content="text/html; charset=UTF-8">
        </head>
        <body>
          <p>Unordered list</p>
          <ul>
            <li>Bullet point 1</li>
            <li>Bullet point 2</li>
            <ul>
              <li>Bullet point 2.1</li>
              <li>Bullet point 2.2</li>
              <ul>
                <li>Bullet point 2.2.1</li>
              </ul>
              <li>Bullet point 2.3</li>
            </ul>
            <li>Bullet point 3</li>
            <ul>
              <li>Bullet point 3.1</li>
            </ul>
          </ul>
          <p><br>
          </p>
          <p>Ordered list</p>
          <ol>
            <li>Number 1</li>
            <ol>
              <li>Number 1.1</li>
              <ol>
                <li>Number 1.1.1</li>
                <li>Number 1.1.2</li>
              </ol>
              <li>Number 1.2<br>
              </li>
            </ol>
            <li>Number 2</li>
            <li>Number 3</li>
            <ol>
              <li>Number 3.1</li>
              <li>Number 3.2</li>
              <ol>
                <li>Number 3.2.1<br>
                </li>
              </ol>
              <li>Number 3.3</li>
              <li>Number 3.4</li>
            </ol>
            <li>Number 4</li>
          </ol>
          <p>End<br>
          </p>
        </body>
      </html>

      --------------oYnsP1x8lKf6V060046qa0DG--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject                                 |
      | [user:user]@[domain] | [user:to]@[domain] | Message with Numbering/Ordering in Body |
    And IMAP client "1" eventually sees 1 messages in "Sent"
    When the user logs in with username "[user:to]" and password "password"
    And user "[user:to]" connects and authenticates IMAP client "2"
    And user "[user:to]" finishes syncing
    And it succeeds
    Then IMAP client "2" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "[user:user]@[domain]",
        "to": "[user:to]@[domain]",
        "subject": "Message with Numbering/Ordering in Body",
        "content": {
          "content-type": "text/html",
          "transfer-encoding": "quoted-printable",
          "body-is": "<!DOCTYPE html>\r\n<html>\r\n  <head>\r\n\r\n    <meta http-equiv=\"content-type\" content=\"text/html; charset=UTF-8\">\r\n  </head>\r\n  <body>\r\n    <p>Unordered list</p>\r\n    <ul>\r\n      <li>Bullet point 1</li>\r\n      <li>Bullet point 2</li>\r\n      <ul>\r\n        <li>Bullet point 2.1</li>\r\n        <li>Bullet point 2.2</li>\r\n        <ul>\r\n          <li>Bullet point 2.2.1</li>\r\n        </ul>\r\n        <li>Bullet point 2.3</li>\r\n      </ul>\r\n      <li>Bullet point 3</li>\r\n      <ul>\r\n        <li>Bullet point 3.1</li>\r\n      </ul>\r\n    </ul>\r\n    <p><br>\r\n    </p>\r\n    <p>Ordered list</p>\r\n    <ol>\r\n      <li>Number 1</li>\r\n      <ol>\r\n        <li>Number 1.1</li>\r\n        <ol>\r\n          <li>Number 1.1.1</li>\r\n          <li>Number 1.1.2</li>\r\n        </ol>\r\n        <li>Number 1.2<br>\r\n        </li>\r\n      </ol>\r\n      <li>Number 2</li>\r\n      <li>Number 3</li>\r\n      <ol>\r\n        <li>Number 3.1</li>\r\n        <li>Number 3.2</li>\r\n        <ol>\r\n          <li>Number 3.2.1<br>\r\n          </li>\r\n        </ol>\r\n        <li>Number 3.3</li>\r\n        <li>Number 3.4</li>\r\n      </ol>\r\n      <li>Number 4</li>\r\n    </ol>\r\n    <p>End<br>\r\n    </p>\r\n  </body>\r\n</html>"
        }
      }
      """

  Scenario: HTML message with public key attached to Internal
    When the account "[user:user]" has public key attachment "enabled"
    And SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      From: <[user:user]@[domain]>
      To: <[user:to]@[domain]>
      Subject: HTML text internal with public key attached
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/html; charset=utf-8

      <!DOCTYPE html>
      <html>
        <head>

          <meta http-equiv="content-type" content="text/html; charset=UTF-8">
        </head>
        <body>
          <p>This is body of <b>HTML mail</b> with public key attachment.<br>
          </p>
        </body>
      </html>

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject                                     |
      | [user:user]@[domain] | [user:to]@[domain] | HTML text internal with public key attached |
    And IMAP client "1" eventually sees 1 messages in "Sent"
    When the user logs in with username "[user:to]" and password "password"
    And user "[user:to]" connects and authenticates IMAP client "2"
    And user "[user:to]" finishes syncing
    And it succeeds
    Then IMAP client "2" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "[user:user]@[domain]",
        "to": "[user:to]@[domain]",
        "subject": "HTML text internal with public key attached",
        "content": {
          "content-type": "multipart/mixed",
          "sections":[
            {
              "content-type": "text/html",
              "transfer-encoding": "text/html",
              "content-type-charset": "utf-8",
              "transfer-encoding": "quoted-printable",
              "body-is": "<!DOCTYPE html>\r\n<html>\r\n  <head>\r\n\r\n    <meta http-equiv=3D\"content-type\" content=3D\"text/html; charset=3DUTF-8=\r\n\">\r\n  </head>\r\n  <body>\r\n    <p>This is body of <b>HTML mail</b> with public key attachment.<br>\r\n    </p>\r\n  </body>\r\n</html>"
            },
            {
              "content-type": "application/pgp-keys",
              "content-disposition": "attachment",

              "transfer-encoding": "base64"
            }
          ]
        }
      }
      """

  Scenario: HTML message with multiple attachments to Internal
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      Content-Type: multipart/mixed; boundary="------------2p04vJsuXgcobQxmsvuPsEB2"
      To: <[user:to]@[domain]>
      From: <[user:user]@[domain]>
      Subject: HTML message with different attachments

      This is a multi-part message in MIME format.
      --------------2p04vJsuXgcobQxmsvuPsEB2
      Content-Type: text/html; charset=UTF-8
      Content-Transfer-Encoding: 7bit

      <!DOCTYPE html>
      <html>
        <head>

          <meta http-equiv="content-type" content="text/html; charset=UTF-8">
        </head>
        <body>
          <p>Hello, this is a <b>HTML message</b> with <i>different
              attachments</i>.<br>
          </p>
        </body>
      </html>
      --------------2p04vJsuXgcobQxmsvuPsEB2
      Content-Type: text/html; charset=UTF-8; name="index.html"
      Content-Disposition: attachment; filename="index.html"
      Content-Transfer-Encoding: base64

      IDwhRE9DVFlQRSBodG1sPg0KPGh0bWw+DQo8aGVhZD4NCjx0aXRsZT5QYWdlIFRpdGxlPC90
      aXRsZT4NCjwvaGVhZD4NCjxib2R5Pg0KDQo8aDE+TXkgRmlyc3QgSGVhZGluZzwvaDE+DQo8
      cD5NeSBmaXJzdCBwYXJhZ3JhcGguPC9wPg0KDQo8L2JvZHk+DQo8L2h0bWw+IA==
      --------------2p04vJsuXgcobQxmsvuPsEB2
      Content-Type: application/vnd.openxmlformats-officedocument.wordprocessingml.document;
        name="test.docx"
      Content-Disposition: attachment; filename="test.docx"
      Content-Transfer-Encoding: base64

      UEsDBBQABgAIAAAAIQDfpNJsWgEAACAFAAATAAgCW0NvbnRlbnRfVHlwZXNdLnhtbCCiBAIo
      oAACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAC0lMtuwjAQRfeV+g+Rt1Vi6KKqKgKLPpYt
      UukHGHsCVv2Sx7z+vhMCUVUBkQpsIiUz994zVsaD0dqabAkRtXcl6xc9loGTXmk3K9nX5C1/
      ZBkm4ZQw3kHJNoBsNLy9GUw2ATAjtcOSzVMKT5yjnIMVWPgAjiqVj1Ykeo0zHoT8FjPg973e
      A5feJXApT7UHGw5eoBILk7LXNX1uSCIYZNlz01hnlUyEYLQUiep86dSflHyXUJBy24NzHfCO
      Ghg/mFBXjgfsdB90NFEryMYipndhqYuvfFRcebmwpCxO2xzg9FWlJbT62i1ELwGRztyaoq1Y
      od2e/ygHpo0BvDxF49sdDymR4BoAO+dOhBVMP69G8cu8E6Si3ImYGrg8RmvdCZFoA6F59s/m
      2NqciqTOcfQBaaPjP8ber2ytzmngADHp039dm0jWZ88H9W2gQB3I5tv7bfgDAAD//wMAUEsD
      BBQABgAIAAAAIQAekRq37wAAAE4CAAALAAgCX3JlbHMvLnJlbHMgogQCKKAAAgAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAArJLBasMwDEDvg/2D0b1R2sEYo04vY9DbGNkHCFtJTBPb2GrX
      /v082NgCXelhR8vS05PQenOcRnXglF3wGpZVDYq9Cdb5XsNb+7x4AJWFvKUxeNZw4gyb5vZm
      /cojSSnKg4tZFYrPGgaR+IiYzcAT5SpE9uWnC2kiKc/UYySzo55xVdf3mH4zoJkx1dZqSFt7
      B6o9Rb6GHbrOGX4KZj+xlzMtkI/C3rJdxFTqk7gyjWop9SwabDAvJZyRYqwKGvC80ep6o7+n
      xYmFLAmhCYkv+3xmXBJa/ueK5hk/Nu8hWbRf4W8bnF1B8wEAAP//AwBQSwMEFAAGAAgAAAAh
      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
      --------------2p04vJsuXgcobQxmsvuPsEB2
      Content-Type: application/pdf; name="test.pdf"
      Content-Disposition: attachment; filename="test.pdf"
      Content-Transfer-Encoding: base64

      JVBERi0xLjUKJeLjz9MKNyAwIG9iago8PAovVHlwZSAvRm9udERlc2NyaXB0b3IKL0ZvbnRO
      YW1lIC9BcmlhbAovRmxhZ3MgMzIKL0l0YWxpY0FuZ2xlIDAKL0FzY2VudCA5MDUKL0Rlc2Nl
      bnQgLTIxMAovQ2FwSGVpZ2h0IDcyOAovQXZnV2lkdGggNDQxCi9NYXhXaWR0aCAyNjY1Ci9G
      b250V2VpZ2h0IDQwMAovWEhlaWdodCAyNTAKL0xlYWRpbmcgMzMKL1N0ZW1WIDQ0Ci9Gb250
      QkJveCBbLTY2NSAtMjEwIDIwMDAgNzI4XQo+PgplbmRvYmoKOCAwIG9iagpbMjc4IDAgMCAw
      IDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAw
      IDAgMCAwIDAgNzIyIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAw
      IDAgMCAwIDAgMCAwIDAgMCA1NTYgNTU2IDUwMCA1NTYgNTU2IDI3OCA1NTYgNTU2IDIyMiAw
      IDUwMCAyMjIgMCA1NTYgNTU2IDU1NiAwIDMzMyA1MDAgMjc4XQplbmRvYmoKNiAwIG9iago8
      PAovVHlwZSAvRm9udAovU3VidHlwZSAvVHJ1ZVR5cGUKL05hbWUgL0YxCi9CYXNlRm9udCAv
      QXJpYWwKL0VuY29kaW5nIC9XaW5BbnNpRW5jb2RpbmcKL0ZvbnREZXNjcmlwdG9yIDcgMCBS
      Ci9GaXJzdENoYXIgMzIKL0xhc3RDaGFyIDExNgovV2lkdGhzIDggMCBSCj4+CmVuZG9iago5
      IDAgb2JqCjw8Ci9UeXBlIC9FeHRHU3RhdGUKL0JNIC9Ob3JtYWwKL2NhIDEKPj4KZW5kb2Jq
      CjEwIDAgb2JqCjw8Ci9UeXBlIC9FeHRHU3RhdGUKL0JNIC9Ob3JtYWwKL0NBIDEKPj4KZW5k
      b2JqCjExIDAgb2JqCjw8Ci9GaWx0ZXIgL0ZsYXRlRGVjb2RlCi9MZW5ndGggMjUwCj4+CnN0
      cmVhbQp4nKWQQUsDMRCF74H8h3dMhGaTuM3uQumh21oUCxUXPIiH2m7XokZt+/9xJruCns3h
      MW/yzQw8ZGtMJtmqvp7DZreb2EG1UU+nmM1rfElhjeVXOQ+LQFpUHsdWiocLRClmjRTZlYNz
      xuZo9lI44iwcCm+sz1HYyoSA5p245X2B7kQ70SVXDm4pxaOq9Vi96FGpWpbt64F81O5Sdeyh
      R7k6aB/UnqtkzyxphNmT9r7vf/LUjoVYP+6bW/7+YDiynLUr6BIxg/1Zmlal6qgHYsPErq9I
      ntm+EdbqJzQ3Uiwogzsp/hOWD9ZU5e+wUkZDNPh7CItVjW9I9VnOCmVuZHN0cmVhbQplbmRv
      YmoKNSAwIG9iago8PAovVHlwZSAvUGFnZQovTWVkaWFCb3ggWzAgMCA2MTIgNzkyXQovUmVz
      b3VyY2VzIDw8Ci9Gb250IDw8Ci9GMSA2IDAgUgo+PgovRXh0R1N0YXRlIDw8Ci9HUzcgOSAw
      IFIKL0dTOCAxMCAwIFIKPj4KL1Byb2NTZXQgWy9QREYgL1RleHQgL0ltYWdlQiAvSW1hZ2VD
      IC9JbWFnZUldCj4+Ci9Db250ZW50cyAxMSAwIFIKL0dyb3VwIDw8Ci9UeXBlIC9Hcm91cAov
      UyAvVHJhbnNwYXJlbmN5Ci9DUyAvRGV2aWNlUkdCCj4+Ci9UYWJzIC9TCi9TdHJ1Y3RQYXJl
      bnRzIDAKL1BhcmVudCAyIDAgUgo+PgplbmRvYmoKMTIgMCBvYmoKPDwKL1MgL1AKL1R5cGUg
      L1N0cnVjdEVsZW0KL0sgWzBdCi9QIDEzIDAgUgovUGcgNSAwIFIKPj4KZW5kb2JqCjEzIDAg
      b2JqCjw8Ci9TIC9QYXJ0Ci9UeXBlIC9TdHJ1Y3RFbGVtCi9LIFsxMiAwIFJdCi9QIDMgMCBS
      Cj4+CmVuZG9iagoxNCAwIG9iago8PAovTnVtcyBbMCBbMTIgMCBSXV0KPj4KZW5kb2JqCjQg
      MCBvYmoKPDwKL0Zvb3Rub3RlIC9Ob3RlCi9FbmRub3RlIC9Ob3RlCi9UZXh0Ym94IC9TZWN0
      Ci9IZWFkZXIgL1NlY3QKL0Zvb3RlciAvU2VjdAovSW5saW5lU2hhcGUgL1NlY3QKL0Fubm90
      YXRpb24gL1NlY3QKL0FydGlmYWN0IC9TZWN0Ci9Xb3JrYm9vayAvRG9jdW1lbnQKL1dvcmtz
      aGVldCAvUGFydAovTWFjcm9zaGVldCAvUGFydAovQ2hhcnRzaGVldCAvUGFydAovRGlhbG9n
      c2hlZXQgL1BhcnQKL1NsaWRlIC9QYXJ0Ci9DaGFydCAvU2VjdAovRGlhZ3JhbSAvRmlndXJl
      Cj4+CmVuZG9iagozIDAgb2JqCjw8Ci9UeXBlIC9TdHJ1Y3RUcmVlUm9vdAovUm9sZU1hcCA0
      IDAgUgovSyBbMTMgMCBSXQovUGFyZW50VHJlZSAxNCAwIFIKL1BhcmVudFRyZWVOZXh0S2V5
      IDEKPj4KZW5kb2JqCjIgMCBvYmoKPDwKL1R5cGUgL1BhZ2VzCi9LaWRzIFs1IDAgUl0KL0Nv
      dW50IDEKPj4KZW5kb2JqCjEgMCBvYmoKPDwKL1R5cGUgL0NhdGFsb2cKL1BhZ2VzIDIgMCBS
      Ci9MYW5nIChlbi1VUykKL1N0cnVjdFRyZWVSb290IDMgMCBSCi9NYXJrSW5mbyA8PAovTWFy
      a2VkIHRydWUKPj4KPj4KZW5kb2JqCjE1IDAgb2JqCjw8Ci9DcmVhdG9yIDxGRUZGMDA0RDAw
      NjkwMDYzMDA3MjAwNkYwMDczMDA2RjAwNjYwMDc0MDBBRTAwMjAwMDU3MDA2RjAwNzIwMDY0
      MDAyMDAwMzIwMDMwMDAzMTAwMzY+Ci9DcmVhdGlvbkRhdGUgKEQ6MjAyMDA4MjAxMjMxMTAr
      MDAnMDAnKQovUHJvZHVjZXIgKHd3dy5pbG92ZXBkZi5jb20pCi9Nb2REYXRlIChEOjIwMjAw
      ODIwMTIzMTEwWikKPj4KZW5kb2JqCnhyZWYKMCAxNgowMDAwMDAwMDAwIDY1NTM1IGYNCjAw
      MDAwMDIwMTQgMDAwMDAgbg0KMDAwMDAwMTk1NyAwMDAwMCBuDQowMDAwMDAxODQ3IDAwMDAw
      IG4NCjAwMDAwMDE1NjQgMDAwMDAgbg0KMDAwMDAwMTA4MyAwMDAwMCBuDQowMDAwMDAwNDc3
      IDAwMDAwIG4NCjAwMDAwMDAwMTUgMDAwMDAgbg0KMDAwMDAwMDI1MiAwMDAwMCBuDQowMDAw
      MDAwNjQ3IDAwMDAwIG4NCjAwMDAwMDA3MDMgMDAwMDAgbg0KMDAwMDAwMDc2MCAwMDAwMCBu
      DQowMDAwMDAxMzgwIDAwMDAwIG4NCjAwMDAwMDE0NTMgMDAwMDAgbg0KMDAwMDAwMTUyMyAw
      MDAwMCBuDQowMDAwMDAyMTI4IDAwMDAwIG4NCnRyYWlsZXIKPDwKL1NpemUgMTYKL1Jvb3Qg
      MSAwIFIKL0luZm8gMTUgMCBSCi9JRCBbPDY2MDhFOTQxN0M1OUExNkEwNjAzMDgxQzY1MTk1
      MzNCPiA8RTU2RENDMTkyRjY1RjAwNzVDN0FDMjE2ODYxQjg1MjA+XQo+PgpzdGFydHhyZWYK
      MjM0NAolJUVPRgo=
      --------------2p04vJsuXgcobQxmsvuPsEB2
      Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet;
        name="test.xlsx"
      Content-Disposition: attachment; filename="test.xlsx"
      Content-Transfer-Encoding: base64

      UEsDBBQABgAIAAAAIQBi7p1oXgEAAJAEAAATAAgCW0NvbnRlbnRfVHlwZXNdLnhtbCCiBAIo
      oAACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACslMtOwzAQRfdI/EPkLUrcskAINe2CxxIq
      UT7AxJPGqmNbnmlp/56J+xBCoRVqN7ESz9x7MvHNaLJubbaCiMa7UgyLgcjAVV4bNy/Fx+wl
      vxcZknJaWe+gFBtAMRlfX41mmwCYcbfDUjRE4UFKrBpoFRY+gOOd2sdWEd/GuQyqWqg5yNvB
      4E5W3hE4yqnTEOPRE9RqaSl7XvPjLUkEiyJ73BZ2XqVQIVhTKWJSuXL6l0u+cyi4M9VgYwLe
      MIaQvQ7dzt8Gu743Hk00GrKpivSqWsaQayu/fFx8er8ojov0UPq6NhVoXy1bnkCBIYLS2ABQ
      a4u0Fq0ybs99xD8Vo0zL8MIg3fsl4RMcxN8bZLqej5BkThgibSzgpceeRE85NyqCfqfIybg4
      wE/tYxx8bqbRB+QERfj/FPYR6brzwEIQycAhJH2H7eDI6Tt77NDlW4Pu8ZbpfzL+BgAA//8D
      AFBLAwQUAAYACAAAACEAtVUwI/QAAABMAgAACwAIAl9yZWxzLy5yZWxzIKIEAiigAAIAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAKySTU/DMAyG70j8h8j31d2QEEJLd0FIuyFUfoBJ3A+1
      jaMkG92/JxwQVBqDA0d/vX78ytvdPI3qyCH24jSsixIUOyO2d62Gl/pxdQcqJnKWRnGs4cQR
      dtX11faZR0p5KHa9jyqruKihS8nfI0bT8USxEM8uVxoJE6UchhY9mYFaxk1Z3mL4rgHVQlPt
      rYawtzeg6pPPm3/XlqbpDT+IOUzs0pkVyHNiZ9mufMhsIfX5GlVTaDlpsGKecjoieV9kbMDz
      RJu/E/18LU6cyFIiNBL4Ms9HxyWg9X9atDTxy515xDcJw6vI8MmCix+o3gEAAP//AwBQSwME
      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
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAfJJRS8MwFIXfBf9DyXubpGNzhrYDlT05EKwovoXkbis2aUii3f69abvVDoaQl9xz7ndP
      LslWB1VHP2Bd1egc0YSgCLRoZKV3OXor1/ESRc5zLXndaMjRERxaFbc3mTBMNBZebGPA+gpc
      FEjaMWFytPfeMIyd2IPiLgkOHcRtYxX34Wp32HDxxXeAU0IWWIHnknuOO2BsRiI6IaUYkebb
      1j1ACgw1KNDeYZpQ/Of1YJW72tArE6eq/NGEN53iTtlSDOLoPrhqNLZtm7SzPkbIT/HH5vm1
      f2pc6W5XAlCRScGEBe4bW2R4egmLq7nzm7DjbQXy4Rj0KzUp+rgDBGQUArAh7ll5nz0+lWtU
      pITOY7KIybykS0bvWEo+u5EX/V2goaBOg/8n3gdcnNKSzlg46XxCPAOG3JefoPgFAAD//wMA
      UEsDBBQABgAIAAAAIQBhSQkQiQEAABEDAAAQAAgBZG9jUHJvcHMvYXBwLnhtbCCiBAEooAAB
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJySQW/bMAyF7wP6Hwzd
      GzndUAyBrGJIV/SwYQGStmdNpmOhsiSIrJHs14+20dTZeuqN5Ht4+kRJ3Rw6X/SQ0cVQieWi
      FAUEG2sX9pV42N1dfhUFkgm18TFAJY6A4kZffFKbHBNkcoAFRwSsREuUVlKibaEzuGA5sNLE
      3BniNu9lbBpn4Tbalw4CyauyvJZwIAg11JfpFCimxFVPHw2tox348HF3TAys1beUvLOG+Jb6
      p7M5Ymyo+H6w4JWci4rptmBfsqOjLpWct2prjYc1B+vGeAQl3wbqHsywtI1xGbXqadWDpZgL
      dH94bVei+G0QBpxK9CY7E4ixBtvUjLVPSFk/xfyMLQChkmyYhmM5985r90UvRwMX58YhYAJh
      4Rxx58gD/mo2JtM7xMs58cgw8U4424FvOnPON16ZT/onex27ZMKRhVP1w4VnfEi7eGsIXtd5
      PlTb1mSo+QVO6z4N1D1vMvshZN2asIf61fO/MDz+4/TD9fJ6UX4u+V1nMyXf/rL+CwAA//8D
      AFBLAQItABQABgAIAAAAIQBi7p1oXgEAAJAEAAATAAAAAAAAAAAAAAAAAAAAAABbQ29udGVu
      dF9UeXBlc10ueG1sUEsBAi0AFAAGAAgAAAAhALVVMCP0AAAATAIAAAsAAAAAAAAAAAAAAAAA
      lwMAAF9yZWxzLy5yZWxzUEsBAi0AFAAGAAgAAAAhAIE+lJfzAAAAugIAABoAAAAAAAAAAAAA
      AAAAvAYAAHhsL19yZWxzL3dvcmtib29rLnhtbC5yZWxzUEsBAi0AFAAGAAgAAAAhAOzE+h3h
      AQAAiAMAAA8AAAAAAAAAAAAAAAAA7wgAAHhsL3dvcmtib29rLnhtbFBLAQItABQABgAIAAAA
      IQDDLoEGoAAAAMsAAAAUAAAAAAAAAAAAAAAAAP0KAAB4bC9zaGFyZWRTdHJpbmdzLnhtbFBL
      AQItABQABgAIAAAAIQB1PplpkwYAAIwaAAATAAAAAAAAAAAAAAAAAM8LAAB4bC90aGVtZS90
      aGVtZTEueG1sUEsBAi0AFAAGAAgAAAAhAJ+I622WAgAABAYAAA0AAAAAAAAAAAAAAAAAkxIA
      AHhsL3N0eWxlcy54bWxQSwECLQAUAAYACAAAACEAZewQ5boBAAAwAwAAGAAAAAAAAAAAAAAA
      AABUFQAAeGwvd29ya3NoZWV0cy9zaGVldDEueG1sUEsBAi0AFAAGAAgAAAAhAJWICOVBAQAA
      UQIAABEAAAAAAAAAAAAAAAAARBcAAGRvY1Byb3BzL2NvcmUueG1sUEsBAi0AFAAGAAgAAAAh
      AGFJCRCJAQAAEQMAABAAAAAAAAAAAAAAAAAAvBkAAGRvY1Byb3BzL2FwcC54bWxQSwUGAAAA
      AAoACgCAAgAAexwAAAAA
      --------------2p04vJsuXgcobQxmsvuPsEB2
      Content-Type: text/xml; charset=UTF-8; name="testxml.xml"
      Content-Disposition: attachment; filename="testxml.xml"
      Content-Transfer-Encoding: base64

      PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPCFET0NUWVBFIHN1aXRl
      IFNZU1RFTSAiaHR0cDovL3Rlc3RuZy5vcmcvdGVzdG5nLTEuMC5kdGQiID4KCjxzdWl0ZSBu
      YW1lPSJBZmZpbGlhdGUgTmV0d29ya3MiPgoKICAgIDx0ZXN0IG5hbWU9IkFmZmlsaWF0ZSBO
      ZXR3b3JrcyIgZW5hYmxlZD0idHJ1ZSI+CiAgICAgICAgPGNsYXNzZXM+CiAgICAgICAgICAg
      IDxjbGFzcyBuYW1lPSJjb20uY2xpY2tvdXQuYXBpdGVzdGluZy5hZmZOZXR3b3Jrcy5Bd2lu
      VUtUZXN0Ii8+CiAgICAgICAgPC9jbGFzc2VzPgogICAgPC90ZXN0PgoKPC9zdWl0ZT4=
      --------------2p04vJsuXgcobQxmsvuPsEB2
      Content-Type: text/plain; charset=UTF-8; name="text file.txt"
      Content-Disposition: attachment; filename="text file.txt"
      Content-Transfer-Encoding: base64

      dGV4dCBmaWxl

      --------------2p04vJsuXgcobQxmsvuPsEB2--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject                                 |
      | [user:user]@[domain] | [user:to]@[domain] | HTML message with different attachments |
    And IMAP client "1" eventually sees 1 messages in "Sent"
    When the user logs in with username "[user:to]" and password "password"
    And user "[user:to]" connects and authenticates IMAP client "2"
    And user "[user:to]" finishes syncing
    And it succeeds
    Then IMAP client "2" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "[user:user]@[domain]",
        "to": "[user:to]@[domain]",
        "subject": "HTML message with different attachments",
        "content": {
          "content-type": "multipart/mixed",
          "sections":[
            {
              "content-type": "text/html",
              "content-type-charset": "utf-8",
              "transfer-encoding": "quoted-printable",
              "body-is": "<!DOCTYPE html>\r\n<html>\r\n  <head>\r\n\r\n    <meta http-equiv=3D\"content-type\" content=3D\"text/html; charset=3DUTF-8=\r\n\">\r\n  </head>\r\n  <body>\r\n    <p>Hello, this is a <b>HTML message</b> with <i>different\r\n        attachments</i>.<br>\r\n    </p>\r\n  </body>\r\n</html>"
            },
            {
              "content-type": "text/html",
              "content-type-name": "index.html",
              "content-disposition": "attachment",
              "content-disposition-filename": "index.html",
              "transfer-encoding": "base64",
              "body-is": "IDwhRE9DVFlQRSBodG1sPg0KPGh0bWw+DQo8aGVhZD4NCjx0aXRsZT5QYWdlIFRpdGxlPC90aXRs\r\nZT4NCjwvaGVhZD4NCjxib2R5Pg0KDQo8aDE+TXkgRmlyc3QgSGVhZGluZzwvaDE+DQo8cD5NeSBm\r\naXJzdCBwYXJhZ3JhcGguPC9wPg0KDQo8L2JvZHk+DQo8L2h0bWw+IA=="
            },
            {
              "content-type": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
              "content-type-name": "test.docx",
              "content-disposition": "attachment",
              "content-disposition-filename": "test.docx",
              "transfer-encoding": "base64",
              "body-is": "UEsDBBQABgAIAAAAIQDfpNJsWgEAACAFAAATAAgCW0NvbnRlbnRfVHlwZXNdLnhtbCCiBAIooAAC\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAC0\r\nlMtuwjAQRfeV+g+Rt1Vi6KKqKgKLPpYtUukHGHsCVv2Sx7z+vhMCUVUBkQpsIiUz994zVsaD0dqa\r\nbAkRtXcl6xc9loGTXmk3K9nX5C1/ZBkm4ZQw3kHJNoBsNLy9GUw2ATAjtcOSzVMKT5yjnIMVWPgA\r\njiqVj1Ykeo0zHoT8FjPg973eA5feJXApT7UHGw5eoBILk7LXNX1uSCIYZNlz01hnlUyEYLQUiep8\r\n6dSflHyXUJBy24NzHfCOGhg/mFBXjgfsdB90NFEryMYipndhqYuvfFRcebmwpCxO2xzg9FWlJbT6\r\n2i1ELwGRztyaoq1Yod2e/ygHpo0BvDxF49sdDymR4BoAO+dOhBVMP69G8cu8E6Si3ImYGrg8Rmvd\r\nCZFoA6F59s/m2NqciqTOcfQBaaPjP8ber2ytzmngADHp039dm0jWZ88H9W2gQB3I5tv7bfgDAAD/\r\n/wMAUEsDBBQABgAIAAAAIQAekRq37wAAAE4CAAALAAgCX3JlbHMvLnJlbHMgogQCKKAAAgAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAArJLBasMw\r\nDEDvg/2D0b1R2sEYo04vY9DbGNkHCFtJTBPb2GrX/v082NgCXelhR8vS05PQenOcRnXglF3wGpZV\r\nDYq9Cdb5XsNb+7x4AJWFvKUxeNZw4gyb5vZm/cojSSnKg4tZFYrPGgaR+IiYzcAT5SpE9uWnC2ki\r\nKc/UYySzo55xVdf3mH4zoJkx1dZqSFt7B6o9Rb6GHbrOGX4KZj+xlzMtkI/C3rJdxFTqk7gyjWop\r\n9SwabDAvJZyRYqwKGvC80ep6o7+nxYmFLAmhCYkv+3xmXBJa/ueK5hk/Nu8hWbRf4W8bnF1B8wEA\r\nAP//AwBQSwMEFAAGAAgAAAAhANZks1H0AAAAMQMAABwACAF3b3JkL19yZWxzL2RvY3VtZW50Lnht\r\nbC5yZWxzIKIEASigAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\n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"
            },
            {
              "content-type": "application/pdf",
              "content-type-name": "test.pdf",
              "content-disposition": "attachment",
              "content-disposition-filename": "test.pdf",
              "transfer-encoding": "base64",
              "body-is": "JVBERi0xLjUKJeLjz9MKNyAwIG9iago8PAovVHlwZSAvRm9udERlc2NyaXB0b3IKL0ZvbnROYW1l\r\nIC9BcmlhbAovRmxhZ3MgMzIKL0l0YWxpY0FuZ2xlIDAKL0FzY2VudCA5MDUKL0Rlc2NlbnQgLTIx\r\nMAovQ2FwSGVpZ2h0IDcyOAovQXZnV2lkdGggNDQxCi9NYXhXaWR0aCAyNjY1Ci9Gb250V2VpZ2h0\r\nIDQwMAovWEhlaWdodCAyNTAKL0xlYWRpbmcgMzMKL1N0ZW1WIDQ0Ci9Gb250QkJveCBbLTY2NSAt\r\nMjEwIDIwMDAgNzI4XQo+PgplbmRvYmoKOCAwIG9iagpbMjc4IDAgMCAwIDAgMCAwIDAgMCAwIDAg\r\nMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgNzIyIDAgMCAw\r\nIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCAwIDAgMCA1NTYg\r\nNTU2IDUwMCA1NTYgNTU2IDI3OCA1NTYgNTU2IDIyMiAwIDUwMCAyMjIgMCA1NTYgNTU2IDU1NiAw\r\nIDMzMyA1MDAgMjc4XQplbmRvYmoKNiAwIG9iago8PAovVHlwZSAvRm9udAovU3VidHlwZSAvVHJ1\r\nZVR5cGUKL05hbWUgL0YxCi9CYXNlRm9udCAvQXJpYWwKL0VuY29kaW5nIC9XaW5BbnNpRW5jb2Rp\r\nbmcKL0ZvbnREZXNjcmlwdG9yIDcgMCBSCi9GaXJzdENoYXIgMzIKL0xhc3RDaGFyIDExNgovV2lk\r\ndGhzIDggMCBSCj4+CmVuZG9iago5IDAgb2JqCjw8Ci9UeXBlIC9FeHRHU3RhdGUKL0JNIC9Ob3Jt\r\nYWwKL2NhIDEKPj4KZW5kb2JqCjEwIDAgb2JqCjw8Ci9UeXBlIC9FeHRHU3RhdGUKL0JNIC9Ob3Jt\r\nYWwKL0NBIDEKPj4KZW5kb2JqCjExIDAgb2JqCjw8Ci9GaWx0ZXIgL0ZsYXRlRGVjb2RlCi9MZW5n\r\ndGggMjUwCj4+CnN0cmVhbQp4nKWQQUsDMRCF74H8h3dMhGaTuM3uQumh21oUCxUXPIiH2m7XokZt\r\n+/9xJruCns3hMW/yzQw8ZGtMJtmqvp7DZreb2EG1UU+nmM1rfElhjeVXOQ+LQFpUHsdWiocLRClm\r\njRTZlYNzxuZo9lI44iwcCm+sz1HYyoSA5p245X2B7kQ70SVXDm4pxaOq9Vi96FGpWpbt64F81O5S\r\ndeyhR7k6aB/UnqtkzyxphNmT9r7vf/LUjoVYP+6bW/7+YDiynLUr6BIxg/1Zmlal6qgHYsPErq9I\r\nntm+EdbqJzQ3Uiwogzsp/hOWD9ZU5e+wUkZDNPh7CItVjW9I9VnOCmVuZHN0cmVhbQplbmRvYmoK\r\nNSAwIG9iago8PAovVHlwZSAvUGFnZQovTWVkaWFCb3ggWzAgMCA2MTIgNzkyXQovUmVzb3VyY2Vz\r\nIDw8Ci9Gb250IDw8Ci9GMSA2IDAgUgo+PgovRXh0R1N0YXRlIDw8Ci9HUzcgOSAwIFIKL0dTOCAx\r\nMCAwIFIKPj4KL1Byb2NTZXQgWy9QREYgL1RleHQgL0ltYWdlQiAvSW1hZ2VDIC9JbWFnZUldCj4+\r\nCi9Db250ZW50cyAxMSAwIFIKL0dyb3VwIDw8Ci9UeXBlIC9Hcm91cAovUyAvVHJhbnNwYXJlbmN5\r\nCi9DUyAvRGV2aWNlUkdCCj4+Ci9UYWJzIC9TCi9TdHJ1Y3RQYXJlbnRzIDAKL1BhcmVudCAyIDAg\r\nUgo+PgplbmRvYmoKMTIgMCBvYmoKPDwKL1MgL1AKL1R5cGUgL1N0cnVjdEVsZW0KL0sgWzBdCi9Q\r\nIDEzIDAgUgovUGcgNSAwIFIKPj4KZW5kb2JqCjEzIDAgb2JqCjw8Ci9TIC9QYXJ0Ci9UeXBlIC9T\r\ndHJ1Y3RFbGVtCi9LIFsxMiAwIFJdCi9QIDMgMCBSCj4+CmVuZG9iagoxNCAwIG9iago8PAovTnVt\r\ncyBbMCBbMTIgMCBSXV0KPj4KZW5kb2JqCjQgMCBvYmoKPDwKL0Zvb3Rub3RlIC9Ob3RlCi9FbmRu\r\nb3RlIC9Ob3RlCi9UZXh0Ym94IC9TZWN0Ci9IZWFkZXIgL1NlY3QKL0Zvb3RlciAvU2VjdAovSW5s\r\naW5lU2hhcGUgL1NlY3QKL0Fubm90YXRpb24gL1NlY3QKL0FydGlmYWN0IC9TZWN0Ci9Xb3JrYm9v\r\nayAvRG9jdW1lbnQKL1dvcmtzaGVldCAvUGFydAovTWFjcm9zaGVldCAvUGFydAovQ2hhcnRzaGVl\r\ndCAvUGFydAovRGlhbG9nc2hlZXQgL1BhcnQKL1NsaWRlIC9QYXJ0Ci9DaGFydCAvU2VjdAovRGlh\r\nZ3JhbSAvRmlndXJlCj4+CmVuZG9iagozIDAgb2JqCjw8Ci9UeXBlIC9TdHJ1Y3RUcmVlUm9vdAov\r\nUm9sZU1hcCA0IDAgUgovSyBbMTMgMCBSXQovUGFyZW50VHJlZSAxNCAwIFIKL1BhcmVudFRyZWVO\r\nZXh0S2V5IDEKPj4KZW5kb2JqCjIgMCBvYmoKPDwKL1R5cGUgL1BhZ2VzCi9LaWRzIFs1IDAgUl0K\r\nL0NvdW50IDEKPj4KZW5kb2JqCjEgMCBvYmoKPDwKL1R5cGUgL0NhdGFsb2cKL1BhZ2VzIDIgMCBS\r\nCi9MYW5nIChlbi1VUykKL1N0cnVjdFRyZWVSb290IDMgMCBSCi9NYXJrSW5mbyA8PAovTWFya2Vk\r\nIHRydWUKPj4KPj4KZW5kb2JqCjE1IDAgb2JqCjw8Ci9DcmVhdG9yIDxGRUZGMDA0RDAwNjkwMDYz\r\nMDA3MjAwNkYwMDczMDA2RjAwNjYwMDc0MDBBRTAwMjAwMDU3MDA2RjAwNzIwMDY0MDAyMDAwMzIw\r\nMDMwMDAzMTAwMzY+Ci9DcmVhdGlvbkRhdGUgKEQ6MjAyMDA4MjAxMjMxMTArMDAnMDAnKQovUHJv\r\nZHVjZXIgKHd3dy5pbG92ZXBkZi5jb20pCi9Nb2REYXRlIChEOjIwMjAwODIwMTIzMTEwWikKPj4K\r\nZW5kb2JqCnhyZWYKMCAxNgowMDAwMDAwMDAwIDY1NTM1IGYNCjAwMDAwMDIwMTQgMDAwMDAgbg0K\r\nMDAwMDAwMTk1NyAwMDAwMCBuDQowMDAwMDAxODQ3IDAwMDAwIG4NCjAwMDAwMDE1NjQgMDAwMDAg\r\nbg0KMDAwMDAwMTA4MyAwMDAwMCBuDQowMDAwMDAwNDc3IDAwMDAwIG4NCjAwMDAwMDAwMTUgMDAw\r\nMDAgbg0KMDAwMDAwMDI1MiAwMDAwMCBuDQowMDAwMDAwNjQ3IDAwMDAwIG4NCjAwMDAwMDA3MDMg\r\nMDAwMDAgbg0KMDAwMDAwMDc2MCAwMDAwMCBuDQowMDAwMDAxMzgwIDAwMDAwIG4NCjAwMDAwMDE0\r\nNTMgMDAwMDAgbg0KMDAwMDAwMTUyMyAwMDAwMCBuDQowMDAwMDAyMTI4IDAwMDAwIG4NCnRyYWls\r\nZXIKPDwKL1NpemUgMTYKL1Jvb3QgMSAwIFIKL0luZm8gMTUgMCBSCi9JRCBbPDY2MDhFOTQxN0M1\r\nOUExNkEwNjAzMDgxQzY1MTk1MzNCPiA8RTU2RENDMTkyRjY1RjAwNzVDN0FDMjE2ODYxQjg1MjA+\r\nXQo+PgpzdGFydHhyZWYKMjM0NAolJUVPRgo="
            },
            {
              "content-type": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
              "content-type-name": "test.xlsx",
              "content-disposition": "attachment",
              "content-disposition-filename": "test.xlsx",
              "transfer-encoding": "base64",
              "body-is": "UEsDBBQABgAIAAAAIQBi7p1oXgEAAJAEAAATAAgCW0NvbnRlbnRfVHlwZXNdLnhtbCCiBAIooAAC\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs\r\nlMtOwzAQRfdI/EPkLUrcskAINe2CxxIqUT7AxJPGqmNbnmlp/56J+xBCoRVqN7ESz9x7MvHNaLJu\r\nbbaCiMa7UgyLgcjAVV4bNy/Fx+wlvxcZknJaWe+gFBtAMRlfX41mmwCYcbfDUjRE4UFKrBpoFRY+\r\ngOOd2sdWEd/GuQyqWqg5yNvB4E5W3hE4yqnTEOPRE9RqaSl7XvPjLUkEiyJ73BZ2XqVQIVhTKWJS\r\nuXL6l0u+cyi4M9VgYwLeMIaQvQ7dzt8Gu743Hk00GrKpivSqWsaQayu/fFx8er8ojov0UPq6NhVo\r\nXy1bnkCBIYLS2ABQa4u0Fq0ybs99xD8Vo0zL8MIg3fsl4RMcxN8bZLqej5BkThgibSzgpceeRE85\r\nNyqCfqfIybg4wE/tYxx8bqbRB+QERfj/FPYR6brzwEIQycAhJH2H7eDI6Tt77NDlW4Pu8ZbpfzL+\r\nBgAA//8DAFBLAwQUAAYACAAAACEAtVUwI/QAAABMAgAACwAIAl9yZWxzLy5yZWxzIKIEAiigAAIA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAKyS\r\nTU/DMAyG70j8h8j31d2QEEJLd0FIuyFUfoBJ3A+1jaMkG92/JxwQVBqDA0d/vX78ytvdPI3qyCH2\r\n4jSsixIUOyO2d62Gl/pxdQcqJnKWRnGs4cQRdtX11faZR0p5KHa9jyqruKihS8nfI0bT8USxEM8u\r\nVxoJE6UchhY9mYFaxk1Z3mL4rgHVQlPtrYawtzeg6pPPm3/XlqbpDT+IOUzs0pkVyHNiZ9mufMhs\r\nIfX5GlVTaDlpsGKecjoieV9kbMDzRJu/E/18LU6cyFIiNBL4Ms9HxyWg9X9atDTxy515xDcJw6vI\r\n8MmCix+o3gEAAP//AwBQSwMEFAAGAAgAAAAhAIE+lJfzAAAAugIAABoACAF4bC9fcmVscy93b3Jr\r\nYm9vay54bWwucmVscyCiBAEooAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\n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"
            },
            {
              "content-type": "text/xml",
              "content-type-name": "testxml.xml",
              "content-disposition": "attachment",
              "content-disposition-filename": "testxml.xml",
              "transfer-encoding": "base64",
              "body-is": "PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPCFET0NUWVBFIHN1aXRlIFNZ\r\nU1RFTSAiaHR0cDovL3Rlc3RuZy5vcmcvdGVzdG5nLTEuMC5kdGQiID4KCjxzdWl0ZSBuYW1lPSJB\r\nZmZpbGlhdGUgTmV0d29ya3MiPgoKICAgIDx0ZXN0IG5hbWU9IkFmZmlsaWF0ZSBOZXR3b3JrcyIg\r\nZW5hYmxlZD0idHJ1ZSI+CiAgICAgICAgPGNsYXNzZXM+CiAgICAgICAgICAgIDxjbGFzcyBuYW1l\r\nPSJjb20uY2xpY2tvdXQuYXBpdGVzdGluZy5hZmZOZXR3b3Jrcy5Bd2luVUtUZXN0Ii8+CiAgICAg\r\nICAgPC9jbGFzc2VzPgogICAgPC90ZXN0PgoKPC9zdWl0ZT4="
            },
            {
              "content-type": "text/plain",
              "content-type-name": "text file.txt",
              "content-disposition": "attachment",
              "content-disposition-filename": "text file.txt",
              "transfer-encoding": "base64",
              "body-is": "dGV4dCBmaWxl"
            }
          ]
        }
      }
      """

  Scenario: HTML message with multiple inline images to Internal
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      Content-Type: multipart/alternative;
        boundary="------------RRg5SZSbY4O8JM8G9ldSpWOd"
      User-Agent: Mozilla Thunderbird
      Content-Language: en-GB
      To: <[user:to]@[domain]>
      From: <[user:user]@[domain]>
      Subject: HTML message with multiple inline images

      This is a multi-part message in MIME format.
      --------------RRg5SZSbY4O8JM8G9ldSpWOd
      Content-Type: text/plain; charset=UTF-8; format=flowed
      Content-Transfer-Encoding: 7bit

      Inline image 1

      Inline image 2

      End


      --------------RRg5SZSbY4O8JM8G9ldSpWOd
      Content-Type: multipart/related;
        boundary="------------pXWj190lQsd0d77xbCjkhoss"

      --------------pXWj190lQsd0d77xbCjkhoss
      Content-Type: text/html; charset=UTF-8
      Content-Transfer-Encoding: 7bit

      <!DOCTYPE html>
      <html>
        <head>

          <meta http-equiv="content-type" content="text/html; charset=UTF-8">
        </head>
        <body>
          <p>Inline image 1</p>
          <p><img src="cid:part1.g6ktVAf2.OzOgqU7w@protonmail.com"
              moz-do-not-send="false"></p>
          <p>Inline image 2</p>
          <p><img src="cid:part2.rAUlK0aY.qNBo3Y1b@protonmail.com"
              moz-do-not-send="false"></p>
          <p>End<br>
          </p>
          <br>
        </body>
      </html>
      --------------pXWj190lQsd0d77xbCjkhoss
      Content-Type: image/png; name="icon_1.png"
      Content-Disposition: inline; filename="icon_1.png"
      Content-Id: <part1.g6ktVAf2.OzOgqU7w@protonmail.com>
      Content-Transfer-Encoding: base64

      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
      --------------pXWj190lQsd0d77xbCjkhoss
      Content-Type: image/webp; name="icon_2.webp"
      Content-Disposition: inline; filename="icon_2.webp"
      Content-Id: <part2.rAUlK0aY.qNBo3Y1b@protonmail.com>
      Content-Transfer-Encoding: base64

      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
      AAADoAQAAQAAAIAAAAAAAAAA

      --------------pXWj190lQsd0d77xbCjkhoss--

      --------------RRg5SZSbY4O8JM8G9ldSpWOd--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject                                  |
      | [user:user]@[domain] | [user:to]@[domain] | HTML message with multiple inline images |
    And IMAP client "1" eventually sees 1 messages in "Sent"
    When the user logs in with username "[user:to]" and password "password"
    And user "[user:to]" connects and authenticates IMAP client "2"
    And user "[user:to]" finishes syncing
    And it succeeds
    Then IMAP client "2" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "[user:user]@[domain]",
        "to": "[user:to]@[domain]",
        "subject": "HTML message with multiple inline images",
        "content": {
          "content-type": "multipart/mixed",
          "sections":[
            {
              "content-type": "multipart/related",
              "sections":[
                {
                  "content-type": "text/html",
                  "content-type-charset": "utf-8",
                  "transfer-encoding": "quoted-printable",
                  "body-is": "<!DOCTYPE html>\r\n<html>\r\n  <head>\r\n\r\n    <meta http-equiv=3D\"content-type\" content=3D\"text/html; charset=3DUTF-8=\r\n\">\r\n  </head>\r\n  <body>\r\n    <p>Inline image 1</p>\r\n    <p><img src=3D\"cid:part1.g6ktVAf2.OzOgqU7w@protonmail.com\"\r\n        moz-do-not-send=3D\"false\"></p>\r\n    <p>Inline image 2</p>\r\n    <p><img src=3D\"cid:part2.rAUlK0aY.qNBo3Y1b@protonmail.com\"\r\n        moz-do-not-send=3D\"false\"></p>\r\n    <p>End<br>\r\n    </p>\r\n    <br>\r\n  </body>\r\n</html>"
                },
                {
                  "content-type": "image/png",
                  "content-type-name": "icon_1.png",
                  "content-disposition": "inline",
                  "content-disposition-filename": "icon_1.png",
                  "transfer-encoding": "base64",
                  "body-is": "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"
                },
                {
                  "content-type": "image/webp",
                  "content-type-name": "icon_2.webp",
                  "content-disposition": "inline",
                  "content-disposition-filename": "icon_2.webp",
                  "transfer-encoding": "base64",
                  "body-is": "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"
                }
              ]
            }
          ]
        }
      }
      """

  Scenario: Replying to a message after enabling attach public key
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      From: <[user:user]@[domain]>
      To: <[user:to]@[domain]>
      Subject: Reply after enabling attach public key
      Content-Type: text/html; charset=UTF-8
      Content-Transfer-Encoding: 8bit
      Message-ID: <something@protonmail.ch>

      <html>
        <head>

          <meta http-equiv="content-type" content="text/html; charset=UTF-8">
        </head>
        <body>
          Subjεέςτ ¶ Ä È asd
        </body>
      </html>
      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject                                |
      | [user:user]@[domain] | [user:to]@[domain] | Reply after enabling attach public key |
    And IMAP client "1" eventually sees 1 messages in "Sent"
    When the account "[user:to]" has public key attachment "enabled"
    And the user logs in with username "[user:to]" and password "password"
    And user "[user:to]" connects and authenticates IMAP client "2"
    And user "[user:to]" connects and authenticates SMTP client "2"
    And user "[user:to]" finishes syncing
    And it succeeds
    Then IMAP client "2" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "[user:user]@[domain]",
        "to": "[user:to]@[domain]",
        "subject": "Reply after enabling attach public key",
        "content": {
          "content-type": "text/html",
          "content-type-charset": "utf-8",
          "transfer-encoding": "quoted-printable",
          "body-is": "<html>\r\n  <head>\r\n\r\n    <meta http-equiv=\"content-type\" content=\"text/html; charset=UTF-8\">\r\n  </head>\r\n  <body>\r\n    Subjεέςτ ¶ Ä È asd\r\n  </body>\r\n</html>"
        }
      }
      """
    When SMTP client "2" sends the following message from "[user:to]@[domain]" to "[user:user]@[domain]":
      """
      From: <[user:to]@[domain]>
      To: <[user:user]@[domain]>
      Subject: RE: Reply after enabling attach public key
      Content-Type: text/html; charset=UTF-8
      Content-Transfer-Encoding: 8bit
      In-Reply-To: <something@protonmail.ch>
      References: <something@protonmail.ch>

      <html>
        <head>

          <meta http-equiv="content-type" content="text/html; charset=UTF-8">
        </head>
        <body>
          <p>This is body of <b>HTML mail</b> with public key attachment.<br>
          </p>
        </body>
      </html>
      """
    Then it succeeds
    Then IMAP client "2" eventually sees 1 messages in "Sent"
    And IMAP client "1" eventually sees the following messages in "Inbox":
      | from               | subject                                    | in-reply-to               | references                | reply-to           |
      | [user:to]@[domain] | RE: Reply after enabling attach public key | <something@protonmail.ch> | <something@protonmail.ch> | [user:to]@[domain] |
    And IMAP client "1" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "[user:to]@[domain]",
        "to": "[user:user]@[domain]",
        "subject": "RE: Reply after enabling attach public key",
        "content": {
          "content-type": "multipart/mixed",
          "sections":[
            {
              "content-type": "text/html",
              "content-type-charset": "utf-8",
              "transfer-encoding": "quoted-printable",
              "body-is": "<html>\r\n  <head>\r\n\r\n    <meta http-equiv=3D\"content-type\" content=3D\"text/html; charset=3DUTF-8=\r\n\">\r\n  </head>\r\n  <body>\r\n    <p>This is body of <b>HTML mail</b> with public key attachment.<br>\r\n    </p>\r\n  </body>\r\n</html>"
            },
            {
              "content-type": "application/pgp-keys",
              "content-disposition": "attachment",
              "transfer-encoding": "base64"
            }
          ]
        }
      }
      """

  Scenario: Forward a HTML message containing content with various HTML elements (i.e. newsletter/advertising emails)
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      Content-Type: multipart/alternative;
        boundary="------------RF04yQzHab6SMOfxvsHjvSeP"
      Subject: Fwd: Learn PDF Manipulation with Python - Our Latest Updated Tutorials!
      References: <something@protonmail.ch>
      To: <[user:to]@[domain]>
      From: <[user:user]@[domain]>
      In-Reply-To: <something@protonmail.ch>
      X-Forwarded-Message-Id: <something@protonmail.ch>

      This is a multi-part message in MIME format.
      --------------RF04yQzHab6SMOfxvsHjvSeP
      Content-Type: text/plain; charset=UTF-8; format=flowed
      Content-Transfer-Encoding: base64

      DQpGb3J3YXJkZWQgbWVzc2FnZSB3aXRoIHZhcmlvdXMgSFRNTCBlbGVtZW50cw0KDQotLS0t
      LS0tLSBGb3J3YXJkZWQgTWVzc2FnZSAtLS0tLS0tLQ0KU3ViamVjdDogCUxlYXJuIFBERiBN
      YW5pcHVsYXRpb24gd2l0aCBQeXRob24gLSBPdXIgTGF0ZXN0IFVwZGF0ZWQgDQpUdXRvcmlh
      bHMhDQpEYXRlOiAJVGh1LCAxOSBPY3QgMjAyMyAxMjowMDo0OCArMDAwMA0KRnJvbTogCUFi
      ZG91IEAgVGhlIFB5dGhvbiBDb2RlIDxhYmRvdUB0aGVweXRob25jb2RlLmNvbT4NClJlcGx5
      LVRvOiAJYWJkb3VAdGhlcHl0aG9uY29kZS5jb20NClRvOiAJZ29yZ2l0ZXN0aW5nM0Bwcm90
      b25tYWlsLmNvbQ0KDQoNCg0KTGVhcm4gUERGIE1hbmlwdWxhdGlvbiB3aXRoIFB5dGhvbiAt
      IE91ciBMYXRlc3QgVXBkYXRlZCBUdXRvcmlhbHMhDQpMZWFybiBob3cgdG8gZXh0cmFjdCB0
      YWJsZXMgZnJvbSBQREYsIGNvbnZlcnQgSFRNTCB0byBQREYsIGFuZCBjb21wcmVzcyANClBE
      RnMgaW4gDQpQeXRob27Nj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCM
      wqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/i
      gIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDN
      j+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzC
      oM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KA
      jMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P
      4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKg
      zY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCM
      wqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/i
      gIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDN
      j+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzC
      oM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KA
      jMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P
      4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKg
      zY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCM
      wqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/i
      gIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDN
      j+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzC
      oM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KA
      jMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P
      4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKg
      zY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCM
      wqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/i
      gIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDN
      j+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzC
      oM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KAjMKgzY/igIzCoM2P4oCMwqDNj+KA
      jMKgzY/igIwgDQoNCg0KDQogIFB5dGhvbiBMb2dvVGhlIFB5dGhvbiBDb2RlDQogIDxodHRw
      czovLzlsY3RqLnIuYS5kLnNlbmRpYm0xLmNvbS9tay9jbC9mL3NoL1NNSzFFOHRIZUZ1Qm9R
      TUM0ajYzMnJrZzhkUmwvVVE0czZkUHNtRUZLPg0KDQoNCiAgRGlzY292ZXIgT3VyIFBERiBN
      YW5pcHVsYXRpb24gVHV0b3JpYWxzIGluIFB5dGhvbg0KDQpIZXkgdGhlcmUsDQoNCkluIHRo
      aXMgbmV3c2xldHRlciwgd2UncmUgc2hhcmluZyBvdXIgbGF0ZXN0IHVwZGF0ZWQgUERGIE1h
      bmlwdWxhdGlvbiANCnR1dG9yaWFsczoNCg0KDQogICAgMS4gSG93IHRvIEV4dHJhY3QgVGFi
      bGVzIGZyb20gUERGIGluIFB5dGhvbg0KDQpJbiB0aGlzIHR1dG9yaWFsLCB5b3Ugd2lsbCBs
      ZWFybiBob3cgdG8gZXh0cmFjdCB0YWJsZXMgZnJvbSBQREYgZmlsZXMgaW4gDQpQeXRob24g
      dXNpbmcgY2FtZWxvdCBhbmQgdGFidWxhIGxpYnJhcmllcyBhbmQgZXhwb3J0IHRoZW0gaW50
      byBzZXZlcmFsIA0KZm9ybWF0cyBzdWNoIGFzIENTViwgZXhjZWwsIFBhbmRhcyBkYXRhZnJh
      bWUgYW5kIEhUTUwuDQoNCkNoZWNrIGl0IG91dDogSG93IHRvIEV4dHJhY3QgVGFibGVzIGZy
      b20gUERGIGluIFB5dGhvbiANCjxodHRwczovLzlsY3RqLnIuYS5kLnNlbmRpYm0xLmNvbS9t
      ay9jbC9mL3NoL1NNSzFFOHRIZUcxM0daQjlGdEZYOGgzUTZ3MjEvR2w3UVlZUVpRU3BwPg0K
      DQoNCiAgICAyLiBIb3cgdG8gQ29udmVydCBIVE1MIHRvIFBERiBpbiBQeXRob24NCg0KTGVh
      cm4gaG93IHlvdSBjYW4gY29udmVydCBIVE1MIHBhZ2VzIHRvIFBERiBmaWxlcyBmcm9tIGFu
      IEhUTUwgZmlsZSwgVVJMIA0Kb3IgZXZlbiBIVE1MIGNvbnRlbnQgc3RyaW5nIHVzaW5nIHdr
      aHRtbHRvcGRmIHRvb2wgYW5kIGl0cyBwZGZraXQgDQp3cmFwcGVyIGluIFB5dGhvbi4NCg0K
      Q2hlY2sgaXQgb3V0OiBIb3cgdG8gQ29udmVydCBIVE1MIHRvIFBERiBpbiBQeXRob24gDQo8
      aHR0cHM6Ly85bGN0ai5yLmEuZC5zZW5kaWJtMS5jb20vbWsvY2wvZi9zaC9TTUsxRTh0SGVH
      N3VpaTA2UjNQMUVXTUE1RWNIL29tZS1yS1B3RFMwQj4NCg0KDQogICAgMy4gSG93IHRvIENv
      bXByZXNzIFBERiBGaWxlcyBpbiBQeXRob24NCg0KQ29tcHJlc3NpbmcgUERGIGFsbG93cyB5
      b3UgdG8gZGVjcmVhc2UgdGhlIGZpbGUgc2l6ZSBhcyBzbWFsbCBhcyANCnBvc3NpYmxlIHdo
      aWxlIG1haW50YWluaW5nIHRoZSBxdWFsaXR5IG9mIHRoZSBtZWRpYSBpbiB0aGF0IFBERiBm
      aWxlLiBBcyANCmEgcmVzdWx0LCBpdCBzaWduaWZpY2FudGx5IGluY3JlYXNlcyBlZmZlY3Rp
      dmVuZXNzIGFuZCBzaGFyZWFiaWxpdHkuDQoNCkluIHRoaXMgdHV0b3JpYWwsIHlvdSB3aWxs
      IGxlYXJuIGhvdyB0byBjb21wcmVzcyBQREYgZmlsZXMgdXNpbmcgdGhlIA0KUERGVHJvbiBs
      aWJyYXJ5IGluIFB5dGhvbi4NCg0KQ2hlY2sgaXQgb3V0OiBIb3cgdG8gQ29tcHJlc3MgUERG
      IEZpbGVzIGluIFB5dGhvbiANCjxodHRwczovLzlsY3RqLnIuYS5kLnNlbmRpYm0xLmNvbS9t
      ay9jbC9mL3NoL1NNSzFFOHRIZUdFbUFxcDNjRFlWS0xldTNYQ1gvaEhoTDZhVjJYOU1NPg0K
      DQrCrQ0KDQpBZGRpdGlvbmFsbHksIGlmIHlvdSBmaW5kIG91ciB0dXRvcmlhbHMgYmVuZWZp
      Y2lhbCwgeW91IG1pZ2h0IHdhbnQgdG8gDQpkZWx2ZSBkZWVwZXIgd2l0aCBvdXIgUHJhY3Rp
      Y2FsIFB5dGhvbiBQREYgUHJvY2Vzc2luZyBlQm9vayANCjxodHRwczovLzlsY3RqLnIuYS5k
      LnNlbmRpYm0xLmNvbS9tay9jbC9mL3NoL1NNSzFFOHRIZUdMZGN6ZTBuTmh6UUF4ZTFwbW4v
      ZFZEMFFrX3ZIc0hwPi4gDQpUaGlzIGVCb29rIGlzIGEgdHJlYXN1cmUgdHJvdmUgZm9yIHRo
      b3NlIGVhZ2VyIHRvIG1hc3RlciB0aGUgYXJ0IG9mIFBERiANCnByb2Nlc3NpbmcgdXNpbmcg
      UHl0aG9uLiBXaXRoIGl0LCB5b3UnbGwgbGVhcm4gdG8gY3JlYXRlLCByZWFkLCB3cml0ZSwg
      DQphbmQgbWFuaXB1bGF0ZSBQREZzLCBkaXZpbmcgaW50byByZWFsLXdvcmxkIHByb2plY3Rz
      IHRoYXQgZGVtb25zdHJhdGUgDQp0aGUgcG93ZXIgb2YgUHl0aG9uIGluIGhhbmRsaW5nIFBE
      RiBvcGVyYXRpb25zIGVmZmljaWVudGx5LiBUbyBzd2VldGVuIA0KdGhlIGRlYWwsIHVzZSB0
      aGUgY29kZSAqU1VCU0NSSUJFUjE1KiBhdCBjaGVja291dCB0byBzbmFnIGEgKjE1JSogZGlz
      Y291bnQhDQoNCkNoZWNrIGl0IG91dDogUHJhY3RpY2FsIFB5dGhvbiBQREYgUHJvY2Vzc2lu
      ZyBlQm9vay4gDQo8aHR0cHM6Ly85bGN0ai5yLmEuZC5zZW5kaWJtMS5jb20vbWsvY2wvZi9z
      aC9TTUsxRTh0SGVHU1Y1OFN4eVhyVFcwR08wOE4zL2VnZFpMMWxmdGlzQz4NCg0KSWYgeW91
      IGhhdmUgYW55IHF1ZXN0aW9ucywgcGxlYXNlIHJlcGx5IHRvIHRoaXMgZW1haWwgYXMgSSBy
      ZXBseSB0byANCmV2ZXJ5IGVtYWlsLCBqdXN0IGdpdmUgbWUgc29tZSB0aW1lIQ0KDQpBbGwg
      dGhlIGJlc3QsDQoNCkFiZG91IEAgVGhlIFB5dGhvbiBDb2RlDQoNCipUaGUgUHl0aG9uIENv
      ZGUqDQoNCkNvbnN0YW50aW5lLCBBbGdlcmlhDQoNClRoaXMgZW1haWwgd2FzIHNlbnQgdG8g
      Z29yZ2l0ZXN0aW5nM0Bwcm90b25tYWlsLmNvbQ0KDQpZb3UndmUgcmVjZWl2ZWQgaXQgYmVj
      YXVzZSB5b3UndmUgc3Vic2NyaWJlZCB0byBvdXIgbmV3c2xldHRlci4NCg0KVmlldyBpbiBi
      cm93c2VyIA0KPGh0dHBzOi8vOWxjdGouci5hLmQuc2VuZGlibTEuY29tL21rL21yL3NoL1NN
      SnowOVNEcmlPSFdQbXAxQ3p2dk84UlZjSTkvbWZmQWtuX3JpSjRxPiANCnwgVW5zdWJzY3Jp
      YmUgDQo8aHR0cHM6Ly85bGN0ai5yLmEuZC5zZW5kaWJtMS5jb20vbWsvdW4vc2gvU01KejA5
      YTB2a2JYc1ZHVW14OWo0QjFZbEp1SC9rYlhGak00bFNxZ2Y+DQoNCg==
      --------------RF04yQzHab6SMOfxvsHjvSeP
      Content-Type: text/html; charset=UTF-8
      Content-Transfer-Encoding: quoted-printable

      <!DOCTYPE html>
      <html>
        <head>

          <meta http-equiv=3D"content-type" content=3D"text/html; charset=3DUTF=
      -8">
        </head>
        <body yahoo=3D"fix" text=3D"#3b3f44" link=3D"#0092ff">
          <p><br>
          </p>
          <div class=3D"moz-forward-container">Forwarded message with various
            HTML elements<br>
            <br>
            -------- Forwarded Message --------
            <table class=3D"moz-email-headers-table" cellspacing=3D"0"
              cellpadding=3D"0" border=3D"0">
              <tbody>
                <tr>
                  <th valign=3D"BASELINE" nowrap=3D"nowrap" align=3D"RIGHT">Sub=
      ject:
                  </th>
                  <td>Learn PDF Manipulation with Python - Our Latest Updated
                    Tutorials!</td>
                </tr>
                <tr>
                  <th valign=3D"BASELINE" nowrap=3D"nowrap" align=3D"RIGHT">Dat=
      e: </th>
                  <td>Thu, 19 Oct 2023 12:00:48 +0000</td>
                </tr>
                <tr>
                  <th valign=3D"BASELINE" nowrap=3D"nowrap" align=3D"RIGHT">Fro=
      m: </th>
                  <td>Abdou @ The Python Code <a class=3D"moz-txt-link-rfc2396E=
      " href=3D"mailto:abdou@thepythoncode.com">&lt;abdou@thepythoncode.com&gt;=
      </a></td>
                </tr>
                <tr>
                  <th valign=3D"BASELINE" nowrap=3D"nowrap" align=3D"RIGHT">Rep=
      ly-To:
                  </th>
                  <td><a class=3D"moz-txt-link-abbreviated" href=3D"mailto:abdo=
      u@thepythoncode.com">abdou@thepythoncode.com</a></td>
                </tr>
                <tr>
                  <th valign=3D"BASELINE" nowrap=3D"nowrap" align=3D"RIGHT">To:=
        </th>
                  <td><a class=3D"moz-txt-link-abbreviated" href=3D"mailto:gorg=
      itesting3@protonmail.com">gorgitesting3@protonmail.com</a></td>
                </tr>
              </tbody>
            </table>
            <br>
            <br>
            <meta http-equiv=3D"Content-Type" content=3D"text/html; charset=3DU=
      TF-8">
            <meta http-equiv=3D"X-UA-Compatible" content=3D"IE=3Dedge">
            <meta name=3D"format-detection" content=3D"telephone=3Dno">
            <meta name=3D"viewport"
              content=3D"width=3Ddevice-width, initial-scale=3D1.0">
            <title>Learn PDF Manipulation with Python - Our Latest Updated
              Tutorials!</title>
            <style type=3D"text/css" emogrify=3D"no">#outlook a{padding:0;}.Ext=
      ernalClass{width:100%;}.ExternalClass, .ExternalClass p, .ExternalClass s=
      pan, .ExternalClass font, .ExternalClass td, .ExternalClass div{line-heig=
      ht:100%;}table td{border-collapse:collapse;mso-line-height-rule:exactly;}=
      =2Eeditable.image{font-size:0 !important;line-height:0 !important;}.nl2go=
      _preheader{display:none !important;mso-hide:all !important;mso-line-heigh=
      t-rule:exactly;visibility:hidden !important;line-height:0px !important;fo=
      nt-size:0px !important;}body{width:100% !important;-webkit-text-size-adju=
      st:100%;-ms-text-size-adjust:100%;margin:0;padding:0;}img{outline:none;te=
      xt-decoration:none;-ms-interpolation-mode:bicubic;}a img{border:none;}tab=
      le{border-collapse:collapse;mso-table-lspace:0pt;mso-table-rspace:0pt;}th=
      {font-weight:normal;text-align:left;}*[class=3D"gmail-fix"]{display:none =
      !important;}</style>
            <style type=3D"text/css" emogrify=3D"no"></style>
            <style type=3D"text/css" emogrify=3D"no"></style>
            <style type=3D"text/css">p, h1, h2, h3, h4, ol, ul{margin:0;}a, a:l=
      ink{color:#0092ff;text-decoration:underline}.nl2go-default-textstyle{colo=
      r:#3b3f44;font-family:arial,helvetica,sans-serif;font-size:16px;line-heig=
      ht:1.5;word-break:break-word}.default-button{color:#ffffff;font-family:ar=
      ial,helvetica,sans-serif;font-size:16px;font-style:normal;font-weight:bol=
      d;line-height:1.15;text-decoration:none;word-break:break-word}.default-he=
      ading1{color:#1F2D3D;font-family:arial,helvetica,sans-serif;font-size:36p=
      x;word-break:break-word}.default-heading2{color:#1F2D3D;font-family:arial=
      ,helvetica,sans-serif;font-size:32px;word-break:break-word}.default-headi=
      ng3{color:#1F2D3D;font-family:arial,helvetica,sans-serif;font-size:24px;w=
      ord-break:break-word}.default-heading4{color:#1F2D3D;font-family:arial,he=
      lvetica,sans-serif;font-size:18px;word-break:break-word}a[x-apple-data-de=
      tectors]{color:inherit !important;text-decoration:inherit !important;font=
      -size:inherit !important;font-family:inherit !important;font-weight:inher=
      it !important;line-height:inherit !important;}.no-show-for-you{border:non=
      e;display:none;float:none;font-size:0;height:0;line-height:0;max-height:0=
      ;mso-hide:all;overflow:hidden;table-layout:fixed;visibility:hidden;width:=
      0;}</style><!--[if mso]><xml> <o:OfficeDocumentSettings> <o:AllowPNG/> <o=
      :PixelsPerInch>96</o:PixelsPerInch> </o:OfficeDocumentSettings> </xml><![=
      endif]-->
            <style type=3D"text/css">a:link{color:#0092ff;text-decoration:under=
      line}</style>
            <table style=3D" mso-hide:all;display:none" cellspacing=3D"0"
              cellpadding=3D"0" border=3D"0">
              <tbody>
                <tr>
                  <td>Learn how to extract tables from PDF, convert HTML to
                    PDF, and compress PDFs in
      Python=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=
      =E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=
      =A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=
      =80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=
      =CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=
      =8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=
      =8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=
      =C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=
      =E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=
      =A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=
      =80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=
      =CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=
      =8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=
      =8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=
      =C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=
      =E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=
      =A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=
      =80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=
      =CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=
      =8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=
      =8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=
      =C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=
      =E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=
      =A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=
      =80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=
      =CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=
      =8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=
      =8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=
      =C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=
      =E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=
      =A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=
      =80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=
      =CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=
      =8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=
      =8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=
      =C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=
      =E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=
      =A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=
      =80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=
      =CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=
      =8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=
      =8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=
      =C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=
      =E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=
      =A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=
      =80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=
      =CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=
      =8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=
      =8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=
      =C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=
      =E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=
      =A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=
      =80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=
      =CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=
      =8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=
      =8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=C2=A0=CD=8F=E2=80=8C=
      =C2=A0</td>
                </tr>
              </tbody>
            </table>
            <!--[if mso]> <img width=3D"1" height=3D"1" src=3D"https://9lctj.r.=
      a.d.sendibm1.com/mk/op/sh/SMJz09hnzmooEakAYhJWCxug11WP/z0Gs0c1_y3H7" /> <=
      ![endif]-->
            <!--[if !mso]> <!----> <img style=3D"display:none"
      src=3D"https://9lctj.r.a.d.sendibm1.com/mk/op/sh/SMJz09hnzmooEakAYhJWCxug=
      11WP/z0Gs0c1_y3H7"
              moz-do-not-send=3D"true">
            <!-- <![endif]-->
            <table role=3D"presentation" class=3D"nl2go-body-table"
              style=3D"width:100%" width=3D"100%" cellspacing=3D"0" cellpadding=
      =3D"0"
              border=3D"0">
              <tbody>
                <tr>
                  <td>
                    <table role=3D"presentation" class=3D"r0-o"
                      style=3D"table-layout:fixed;width:600px" width=3D"600"
                      cellspacing=3D"0" cellpadding=3D"0" border=3D"0"
                      align=3D"center">
                      <tbody>
                        <tr>
                          <td valign=3D"top">
                            <table role=3D"presentation" class=3D"r2-o"
                              style=3D"table-layout:fixed;width:100%"
                              width=3D"100%" cellspacing=3D"0" cellpadding=3D"0=
      "
                              border=3D"0" align=3D"center">
                              <tbody>
                                <tr>
                                  <td class=3D"r3-i"
      style=3D"background-color:#ffffff;padding-bottom:20px;padding-top:20px">
                                    <table role=3D"presentation" width=3D"100%"=

                                      cellspacing=3D"0" cellpadding=3D"0"
                                      border=3D"0">
                                      <tbody>
                                        <tr>
                                          <th class=3D"r4-c"
                                            style=3D"font-weight:normal"
                                            width=3D"100%" valign=3D"top">
                                            <table role=3D"presentation"
                                              class=3D"r5-o"
      style=3D"table-layout:fixed;width:100%" width=3D"100%" cellspacing=3D"0"
                                              cellpadding=3D"0" border=3D"0">
                                              <tbody>
                                                <tr>
                                                  <td class=3D"r6-i"
      style=3D"padding-left:15px;padding-right:15px" valign=3D"top">
                                                    <table role=3D"presentation=
      "
                                                      width=3D"100%"
                                                      cellspacing=3D"0"
                                                      cellpadding=3D"0"
                                                      border=3D"0">
                                                      <tbody>
                                                        <tr>
                                                          <td class=3D"r7-c">
                                                            <table
      role=3D"presentation" class=3D"r5-o" style=3D"table-layout:fixed;width:57=
      0px"
                                                              width=3D"570"
                                                              cellspacing=3D"0"=

                                                              cellpadding=3D"0"=

                                                              border=3D"0">
                                                              <tbody>
                                                                <tr>
                                                                <td
      class=3D"r8-i nl2go-default-textstyle"
      style=3D"color:#3b3f44;font-family:arial,helvetica,sans-serif;font-size:1=
      6px;line-height:1.5;word-break:break-word;padding-bottom:15px;padding-top=
      :15px">
                                                                <h1
      style=3D"margin:0;font-family:Arial;color:#000000;margin-top:0;font-weigh=
      t:400"><img
      alt=3D"Python Logo"
      src=3D"https://9lctj.img.a.d.sendibm1.com/im/sh/2fMMxBUOJlpV.png?u=3D7xwQ=
      LFBtniwQn1MAnaHPuN5TE0tZRUj"
      title=3D"Python Logo" class=3D"CToWUd" data-bit=3D"iit"
      style=3D"display:block;width:50px;height:50px;float:left;margin-right:5px=
      ;padding-top:5px"
                                                                sib_img_id=3D"0=
      "
      moz-do-not-send=3D"true" width=3D"40" height=3D"50"><a
      href=3D"https://9lctj.r.a.d.sendibm1.com/mk/cl/f/sh/SMK1E8tHeFuBoQMC4j632=
      rkg8dRl/UQ4s6dPsmEFK"
      target=3D"_blank" style=3D"text-decoration:none;color:#2f89fc"
      sib_link_id=3D"0" templating=3D"n" moz-do-not-send=3D"true">The Python Co=
      de</a></h1>
                                                                </td>
                                                                </tr>
                                                              </tbody>
                                                            </table>
                                                          </td>
                                                        </tr>
                                                      </tbody>
                                                    </table>
                                                  </td>
                                                </tr>
                                              </tbody>
                                            </table>
                                          </th>
                                        </tr>
                                      </tbody>
                                    </table>
                                  </td>
                                </tr>
                              </tbody>
                            </table>
                            <table role=3D"presentation" class=3D"r2-o"
                              style=3D"table-layout:fixed;width:100%"
                              width=3D"100%" cellspacing=3D"0" cellpadding=3D"0=
      "
                              border=3D"0" align=3D"center">
                              <tbody>
                                <tr>
                                  <td class=3D"r9-i"
      style=3D"background-color:#ffffff;padding-bottom:20px;padding-top:20px">
                                    <table role=3D"presentation" width=3D"100%"=

                                      cellspacing=3D"0" cellpadding=3D"0"
                                      border=3D"0">
                                      <tbody>
                                        <tr>
                                          <th class=3D"r4-c"
                                            style=3D"font-weight:normal"
                                            width=3D"100%" valign=3D"top">
                                            <table role=3D"presentation"
                                              class=3D"r5-o"
      style=3D"table-layout:fixed;width:100%" width=3D"100%" cellspacing=3D"0"
                                              cellpadding=3D"0" border=3D"0">
                                              <tbody>
                                                <tr>
                                                  <td class=3D"r6-i"
                                                    valign=3D"top">
                                                    <table role=3D"presentation=
      "
                                                      width=3D"100%"
                                                      cellspacing=3D"0"
                                                      cellpadding=3D"0"
                                                      border=3D"0">
                                                      <tbody>
                                                        <tr>
                                                          <td class=3D"r10-c"
                                                            align=3D"left">
                                                            <table
      role=3D"presentation" class=3D"r11-o" style=3D"table-layout:fixed;width:1=
      00%"
                                                              width=3D"100%"
                                                              cellspacing=3D"0"=

                                                              cellpadding=3D"0"=

                                                              border=3D"0">
                                                              <tbody>
                                                                <tr>
                                                                <td
      class=3D"r12-i nl2go-default-textstyle"
      style=3D"color:#3b3f44;font-family:arial,helvetica,sans-serif;font-size:1=
      6px;line-height:1.5;word-break:break-word;padding-top:15px;text-align:lef=
      t"
                                                                valign=3D"top"
                                                                align=3D"left">=

                                                                <div>
                                                                <h1
      class=3D"default-heading1"
      style=3D"margin:0;color:#1f2d3d;font-family:arial,helvetica,sans-serif;fo=
      nt-size:36px;word-break:break-word"><span
      style=3D"font-size:28px">Discover Our PDF Manipulation Tutorials in Pytho=
      n</span></h1>
                                                                </div>
                                                                </td>
                                                                </tr>
                                                              </tbody>
                                                            </table>
                                                          </td>
                                                        </tr>
                                                        <tr>
                                                          <td class=3D"r10-c"
                                                            align=3D"left">
                                                            <table
      role=3D"presentation" class=3D"r11-o" style=3D"table-layout:fixed;width:1=
      00%"
                                                              width=3D"100%"
                                                              cellspacing=3D"0"=

                                                              cellpadding=3D"0"=

                                                              border=3D"0">
                                                              <tbody>
                                                                <tr>
                                                                <td
      class=3D"r12-i nl2go-default-textstyle"
      style=3D"color:#3b3f44;font-family:arial,helvetica,sans-serif;font-size:1=
      6px;line-height:1.5;word-break:break-word;padding-top:15px;text-align:lef=
      t"
                                                                valign=3D"top"
                                                                align=3D"left">=

                                                                <div>
                                                                <p
      style=3D"margin:0"><span style=3D"font-size:16px">Hey there,</span></p>
                                                                <p
      style=3D"margin:0">=C2=A0</p>
                                                                <p
      style=3D"margin:0"><span style=3D"font-size:16px">In this newsletter, we'=
      re
                                                                sharing our
                                                                latest updated
                                                                PDF
                                                                Manipulation
                                                                tutorials:</spa=
      n></p>
                                                                <h2
      class=3D"default-heading2"
      style=3D"margin:0;color:#1f2d3d;font-family:arial,helvetica,sans-serif;fo=
      nt-size:32px;word-break:break-word"><span
      style=3D"font-family:Arial;font-size:24px">1. How to Extract Tables from
                                                                PDF in Python</=
      span></h2>
                                                                <p
      style=3D"margin:0">In this tutorial, you will learn how to extract tables=

                                                                from PDF files
                                                                in Python
                                                                using camelot
                                                                and tabula
                                                                libraries and
                                                                export them
                                                                into several
                                                                formats such
                                                                as CSV, excel,
                                                                Pandas
                                                                dataframe and
                                                                HTML.</p>
                                                                <p
      style=3D"margin:0">=C2=A0</p>
                                                                <p
      style=3D"margin:0">Check it out: <a
      href=3D"https://9lctj.r.a.d.sendibm1.com/mk/cl/f/sh/SMK1E8tHeG13GZB9FtFX8=
      h3Q6w21/Gl7QYYQZQSpp"
      title=3D"How to Extract Tables from PDF in Python" target=3D"_blank"
      style=3D"color:#0092ff;text-decoration:underline" sib_link_id=3D"1"
                                                                templating=3D"n=
      "
      moz-do-not-send=3D"true">How to Extract Tables from PDF in Python</a><br>=

                                                                =C2=A0</p>
                                                                <h2
      class=3D"default-heading2"
      style=3D"margin:0;color:#1f2d3d;font-family:arial,helvetica,sans-serif;fo=
      nt-size:32px;word-break:break-word"><span
      style=3D"font-size:24px">2. How to Convert HTML to PDF in Python</span></=
      h2>
                                                                <p
      style=3D"margin:0">Learn how you can convert HTML pages to PDF files from=

                                                                an HTML file,
                                                                URL or even
                                                                HTML content
                                                                string using
                                                                wkhtmltopdf
                                                                tool and its
                                                                pdfkit wrapper
                                                                in Python.</p>
                                                                <p
      style=3D"margin:0">=C2=A0</p>
                                                                <p
      style=3D"margin:0">Check it out: <a
      href=3D"https://9lctj.r.a.d.sendibm1.com/mk/cl/f/sh/SMK1E8tHeG7uii06R3P1E=
      WMA5EcH/ome-rKPwDS0B"
      title=3D"How to Convert HTML to PDF in Python" target=3D"_blank"
      style=3D"color:#0092ff;text-decoration:underline" sib_link_id=3D"2"
                                                                templating=3D"n=
      "
      moz-do-not-send=3D"true">How to Convert HTML to PDF in Python</a></p>
                                                                <p
      style=3D"margin:0">=C2=A0</p>
                                                                <h2
      class=3D"default-heading2"
      style=3D"margin:0;color:#1f2d3d;font-family:arial,helvetica,sans-serif;fo=
      nt-size:32px;word-break:break-word"><span
      style=3D"font-size:24px">3. How to Compress PDF Files in Python</span></h=
      2>
                                                                <p
      style=3D"margin:0">Compressing PDF allows you to decrease the file size a=
      s
                                                                small as
                                                                possible while
                                                                maintaining
                                                                the quality of
                                                                the media in
                                                                that PDF file.
                                                                As a result,
                                                                it
                                                                significantly
                                                                increases
                                                                effectiveness
                                                                and
                                                                shareability.</=
      p>
                                                                <p
      style=3D"margin:0">=C2=A0</p>
                                                                <p
      style=3D"margin:0">In this tutorial, you will learn how to compress PDF
                                                                files using
                                                                the PDFTron
                                                                library in
                                                                Python.</p>
                                                                <p
      style=3D"margin:0">=C2=A0</p>
                                                                <p
      style=3D"margin:0">Check it out: <a
      href=3D"https://9lctj.r.a.d.sendibm1.com/mk/cl/f/sh/SMK1E8tHeGEmAqp3cDYVK=
      Leu3XCX/hHhL6aV2X9MM"
      title=3D"How to Compress PDF Files in Python" target=3D"_blank"
      style=3D"color:#0092ff;text-decoration:underline" sib_link_id=3D"3"
                                                                templating=3D"n=
      "
      moz-do-not-send=3D"true">How to Compress PDF Files in Python</a></p>
                                                                </div>
                                                                </td>
                                                                </tr>
                                                              </tbody>
                                                            </table>
                                                          </td>
                                                        </tr>
                                                        <tr>
                                                          <td class=3D"r13-c"
                                                            align=3D"center">
                                                            <table
      role=3D"presentation" class=3D"r2-o" style=3D"table-layout:fixed" width=3D=
      "600"
                                                              cellspacing=3D"0"=

                                                              cellpadding=3D"0"=

                                                              border=3D"0">
                                                              <tbody>
                                                                <tr>
                                                                <td
                                                                class=3D"r14-i"=

      style=3D"padding-bottom:30px;padding-top:30px;height:2px">
                                                                <table
      role=3D"presentation" width=3D"100%" cellspacing=3D"0" cellpadding=3D"0"
                                                                border=3D"0">
                                                                <tbody>
                                                                <tr>
                                                                <td>
                                                                <table
      role=3D"presentation" valign=3D"" class=3D"r14-i"
      style=3D"border-top-style:solid;background-clip:border-box;border-top-col=
      or:#4A4A4A;border-top-width:2px;font-size:2px;line-height:2px"
                                                                width=3D"100%"
                                                                height=3D"2"
      cellspacing=3D"0" cellpadding=3D"0" border=3D"0">
                                                                <tbody>
                                                                <tr>
                                                                <td
      style=3D"font-size:0px;line-height:0px" height=3D"0">=C2=AD</td>
                                                                </tr>
                                                                </tbody>
                                                                </table>
                                                                </td>
                                                                </tr>
                                                                </tbody>
                                                                </table>
                                                                </td>
                                                                </tr>
                                                              </tbody>
                                                            </table>
                                                          </td>
                                                        </tr>
                                                        <tr>
                                                          <td class=3D"r10-c"
                                                            align=3D"left">
                                                            <table
      role=3D"presentation" class=3D"r11-o" style=3D"table-layout:fixed;width:1=
      00%"
                                                              width=3D"100%"
                                                              cellspacing=3D"0"=

                                                              cellpadding=3D"0"=

                                                              border=3D"0">
                                                              <tbody>
                                                                <tr>
                                                                <td
      class=3D"r15-i nl2go-default-textstyle"
      style=3D"color:#3b3f44;font-family:arial,helvetica,sans-serif;font-size:1=
      6px;line-height:1.5;word-break:break-word;padding-bottom:15px;padding-top=
      :15px;text-align:left"
                                                                valign=3D"top"
                                                                align=3D"left">=

                                                                <div>
                                                                <p
      style=3D"margin:0">Additionally, if you find our tutorials beneficial, yo=
      u
                                                                might want to
                                                                delve deeper
                                                                with our <a
      href=3D"https://9lctj.r.a.d.sendibm1.com/mk/cl/f/sh/SMK1E8tHeGLdcze0nNhzQ=
      Axe1pmn/dVD0Qk_vHsHp"
      title=3D"Practical Python PDF Processing EBook" target=3D"_blank"
      style=3D"color:#0092ff;text-decoration:underline" sib_link_id=3D"4"
                                                                templating=3D"n=
      "
      moz-do-not-send=3D"true">Practical Python PDF Processing eBook</a>. This
                                                                eBook is a
                                                                treasure trove
                                                                for those
                                                                eager to
                                                                master the art
                                                                of PDF
                                                                processing
                                                                using Python.
                                                                With it,
                                                                you'll learn
                                                                to create,
                                                                read, write,
                                                                and manipulate
                                                                PDFs, diving
                                                                into
                                                                real-world
                                                                projects that
                                                                demonstrate
                                                                the power of
                                                                Python in
                                                                handling PDF
                                                                operations
                                                                efficiently.
                                                                To sweeten the
                                                                deal, use the
                                                                code <strong>SU=
      BSCRIBER15</strong>
                                                                at checkout to
                                                                snag a <strong>=
      15%</strong>
                                                                discount!</p>
                                                                <p
      style=3D"margin:0">=C2=A0</p>
                                                                <p
      style=3D"margin:0">Check it out: <a
      href=3D"https://9lctj.r.a.d.sendibm1.com/mk/cl/f/sh/SMK1E8tHeGSV58SxyXrTW=
      0GO08N3/egdZL1lftisC"
      title=3D"Practical Python PDF Processing eBook" target=3D"_blank"
      style=3D"color:#0092ff;text-decoration:underline" sib_link_id=3D"5"
                                                                templating=3D"n=
      "
      moz-do-not-send=3D"true">Practical Python PDF Processing eBook.</a></p>
                                                                <p
      style=3D"margin:0">=C2=A0</p>
                                                                <p
      style=3D"margin:0">If you have any questions, please reply to this email
                                                                as I reply to
                                                                every email,
                                                                just give me
                                                                some time!</p>
                                                                <p
      style=3D"margin:0">=C2=A0</p>
                                                                <p
      style=3D"margin:0">All the best,</p>
                                                                <p
      style=3D"margin:0">Abdou @ The Python Code</p>
                                                                </div>
                                                                </td>
                                                                </tr>
                                                              </tbody>
                                                            </table>
                                                          </td>
                                                        </tr>
                                                      </tbody>
                                                    </table>
                                                  </td>
                                                </tr>
                                              </tbody>
                                            </table>
                                          </th>
                                        </tr>
                                      </tbody>
                                    </table>
                                  </td>
                                </tr>
                              </tbody>
                            </table>
                            <table role=3D"presentation" class=3D"r2-o"
                              style=3D"table-layout:fixed;width:100%"
                              width=3D"100%" cellspacing=3D"0" cellpadding=3D"0=
      "
                              border=3D"0" align=3D"center">
                              <tbody>
                                <tr>
                                  <td class=3D"r16-i"
      style=3D"background-color:#eff2f7;padding-bottom:20px;padding-top:20px">
                                    <table role=3D"presentation" width=3D"100%"=

                                      cellspacing=3D"0" cellpadding=3D"0"
                                      border=3D"0">
                                      <tbody>
                                        <tr>
                                          <th class=3D"r4-c"
                                            style=3D"font-weight:normal"
                                            width=3D"100%" valign=3D"top">
                                            <table role=3D"presentation"
                                              class=3D"r5-o"
      style=3D"table-layout:fixed;width:100%" width=3D"100%" cellspacing=3D"0"
                                              cellpadding=3D"0" border=3D"0">
                                              <tbody>
                                                <tr>
                                                  <td class=3D"r6-i"
      style=3D"padding-left:15px;padding-right:15px" valign=3D"top">
                                                    <table role=3D"presentation=
      "
                                                      width=3D"100%"
                                                      cellspacing=3D"0"
                                                      cellpadding=3D"0"
                                                      border=3D"0">
                                                      <tbody>
                                                        <tr>
                                                          <td class=3D"r10-c"
                                                            align=3D"left">
                                                            <table
      role=3D"presentation" class=3D"r11-o" style=3D"table-layout:fixed;width:1=
      00%"
                                                              width=3D"100%"
                                                              cellspacing=3D"0"=

                                                              cellpadding=3D"0"=

                                                              border=3D"0">
                                                              <tbody>
                                                                <tr>
                                                                <td
      class=3D"r17-i nl2go-default-textstyle"
      style=3D"font-family:arial,helvetica,sans-serif;word-break:break-word;col=
      or:#3b3f44;font-size:18px;line-height:1.5;padding-top:15px;text-align:cen=
      ter"
                                                                valign=3D"top"
                                                                align=3D"center=
      ">
                                                                <div>
                                                                <p
      style=3D"margin:0"><strong>The Python Code</strong></p>
                                                                </div>
                                                                </td>
                                                                </tr>
                                                              </tbody>
                                                            </table>
                                                          </td>
                                                        </tr>
                                                        <tr>
                                                          <td class=3D"r10-c"
                                                            align=3D"left">
                                                            <table
      role=3D"presentation" class=3D"r11-o" style=3D"table-layout:fixed;width:1=
      00%"
                                                              width=3D"100%"
                                                              cellspacing=3D"0"=

                                                              cellpadding=3D"0"=

                                                              border=3D"0">
                                                              <tbody>
                                                                <tr>
                                                                <td
      class=3D"r18-i nl2go-default-textstyle"
      style=3D"font-family:arial,helvetica,sans-serif;word-break:break-word;col=
      or:#3b3f44;font-size:18px;line-height:1.5;text-align:center"
                                                                valign=3D"top"
                                                                align=3D"center=
      ">
                                                                <div>
                                                                <p
      style=3D"margin:0;font-size:14px">Constantine, Algeria</p>
                                                                </div>
                                                                </td>
                                                                </tr>
                                                              </tbody>
                                                            </table>
                                                          </td>
                                                        </tr>
                                                        <tr>
                                                          <td class=3D"r10-c"
                                                            align=3D"left">
                                                            <table
      role=3D"presentation" class=3D"r11-o" style=3D"table-layout:fixed;width:1=
      00%"
                                                              width=3D"100%"
                                                              cellspacing=3D"0"=

                                                              cellpadding=3D"0"=

                                                              border=3D"0">
                                                              <tbody>
                                                                <tr>
                                                                <td
      class=3D"r17-i nl2go-default-textstyle"
      style=3D"font-family:arial,helvetica,sans-serif;word-break:break-word;col=
      or:#3b3f44;font-size:18px;line-height:1.5;padding-top:15px;text-align:cen=
      ter"
                                                                valign=3D"top"
                                                                align=3D"center=
      ">
                                                                <div>
                                                                <p
      style=3D"margin:0;font-size:14px">This email was sent to
                                                                <a class=3D"moz=
      -txt-link-abbreviated" href=3D"mailto:gorgitesting3@protonmail.com">gorgi=
      testing3@protonmail.com</a></p>
                                                                </div>
                                                                </td>
                                                                </tr>
                                                              </tbody>
                                                            </table>
                                                          </td>
                                                        </tr>
                                                        <tr>
                                                          <td class=3D"r10-c"
                                                            align=3D"left">
                                                            <table
      role=3D"presentation" class=3D"r11-o" style=3D"table-layout:fixed;width:1=
      00%"
                                                              width=3D"100%"
                                                              cellspacing=3D"0"=

                                                              cellpadding=3D"0"=

                                                              border=3D"0">
                                                              <tbody>
                                                                <tr>
                                                                <td
      class=3D"r18-i nl2go-default-textstyle"
      style=3D"font-family:arial,helvetica,sans-serif;word-break:break-word;col=
      or:#3b3f44;font-size:18px;line-height:1.5;text-align:center"
                                                                valign=3D"top"
                                                                align=3D"center=
      ">
                                                                <div>
                                                                <p
      style=3D"margin:0;font-size:14px">You've received it because you've
                                                                subscribed to
                                                                our
                                                                newsletter.</p>=

                                                                </div>
                                                                </td>
                                                                </tr>
                                                              </tbody>
                                                            </table>
                                                          </td>
                                                        </tr>
                                                        <tr>
                                                          <td class=3D"r10-c"
                                                            align=3D"left">
                                                            <table
      role=3D"presentation" class=3D"r11-o" style=3D"table-layout:fixed;width:1=
      00%"
                                                              width=3D"100%"
                                                              cellspacing=3D"0"=

                                                              cellpadding=3D"0"=

                                                              border=3D"0">
                                                              <tbody>
                                                                <tr>
                                                                <td
      class=3D"r19-i nl2go-default-textstyle"
      style=3D"font-family:arial,helvetica,sans-serif;word-break:break-word;col=
      or:#3b3f44;font-size:18px;line-height:1.5;padding-bottom:15px;padding-top=
      :15px;text-align:center"
                                                                valign=3D"top"
                                                                align=3D"center=
      ">
                                                                <div>
                                                                <p
      style=3D"margin:0;font-size:14px"><a
      href=3D"https://9lctj.r.a.d.sendibm1.com/mk/mr/sh/SMJz09SDriOHWPmp1CzvvO8=
      RVcI9/mffAkn_riJ4q"
      style=3D"color:#0092ff;text-decoration:underline" moz-do-not-send=3D"true=
      ">View
                                                                in browser</a>
                                                                | <a
      href=3D"https://9lctj.r.a.d.sendibm1.com/mk/un/sh/SMJz09a0vkbXsVGUmx9j4B1=
      YlJuH/kbXFjM4lSqgf"
      style=3D"color:#0092ff;text-decoration:underline" moz-do-not-send=3D"true=
      ">Unsubscribe</a></p>
                                                                </div>
                                                                </td>
                                                                </tr>
                                                              </tbody>
                                                            </table>
                                                          </td>
                                                        </tr>
                                                      </tbody>
                                                    </table>
                                                  </td>
                                                </tr>
                                              </tbody>
                                            </table>
                                          </th>
                                        </tr>
                                      </tbody>
                                    </table>
                                  </td>
                                </tr>
                              </tbody>
                            </table>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </body>
      </html>

      --------------RF04yQzHab6SMOfxvsHjvSeP--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject                                                                 | X-Forwarded-Message-Id  |
      | [user:user]@[domain] | [user:to]@[domain] | Fwd: Learn PDF Manipulation with Python - Our Latest Updated Tutorials! | something@protonmail.ch |
    And IMAP client "1" eventually sees 1 messages in "Sent"
    When the user logs in with username "[user:to]" and password "password"
    And user "[user:to]" connects and authenticates IMAP client "2"
    And user "[user:to]" finishes syncing
    And it succeeds
    Then IMAP client "2" eventually sees the following messages in "Inbox":
      | from                 | to                 | subject                                                                 | X-Forwarded-Message-Id  |
      | [user:user]@[domain] | [user:to]@[domain] | Fwd: Learn PDF Manipulation with Python - Our Latest Updated Tutorials! | something@protonmail.ch |
    Then IMAP client "2" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "[user:user]@[domain]",
        "to": "[user:to]@[domain]",
        "subject": "Fwd: Learn PDF Manipulation with Python - Our Latest Updated Tutorials!",
        "content":{
          "content-type": "text/html",
          "content-type-charset": "utf-8",
          "transfer-encoding": "quoted-printable",
          "body-is": "<!DOCTYPE html>\r\n<html>\r\n  <head>\r\n\r\n    <meta http-equiv=\"content-type\" content=\"text/html; charset=UTF-8\">\r\n  </head>\r\n  <body yahoo=\"fix\" text=\"#3b3f44\" link=\"#0092ff\">\r\n    <p><br>\r\n    </p>\r\n    <div class=\"moz-forward-container\">Forwarded message with various\r\n      HTML elements<br>\r\n      <br>\r\n      -------- Forwarded Message --------\r\n      <table class=\"moz-email-headers-table\" cellspacing=\"0\"\r\n        cellpadding=\"0\" border=\"0\">\r\n        <tbody>\r\n          <tr>\r\n            <th valign=\"BASELINE\" nowrap=\"nowrap\" align=\"RIGHT\">Subject:\r\n            </th>\r\n            <td>Learn PDF Manipulation with Python - Our Latest Updated\r\n              Tutorials!</td>\r\n          </tr>\r\n          <tr>\r\n            <th valign=\"BASELINE\" nowrap=\"nowrap\" align=\"RIGHT\">Date: </th>\r\n            <td>Thu, 19 Oct 2023 12:00:48 +0000</td>\r\n          </tr>\r\n          <tr>\r\n            <th valign=\"BASELINE\" nowrap=\"nowrap\" align=\"RIGHT\">From: </th>\r\n            <td>Abdou @ The Python Code <a class=\"moz-txt-link-rfc2396E\" href=\"mailto:abdou@thepythoncode.com\">&lt;abdou@thepythoncode.com&gt;</a></td>\r\n          </tr>\r\n          <tr>\r\n            <th valign=\"BASELINE\" nowrap=\"nowrap\" align=\"RIGHT\">Reply-To:\r\n            </th>\r\n            <td><a class=\"moz-txt-link-abbreviated\" href=\"mailto:abdou@thepythoncode.com\">abdou@thepythoncode.com</a></td>\r\n          </tr>\r\n          <tr>\r\n            <th valign=\"BASELINE\" nowrap=\"nowrap\" align=\"RIGHT\">To:  </th>\r\n            <td><a class=\"moz-txt-link-abbreviated\" href=\"mailto:gorgitesting3@protonmail.com\">gorgitesting3@protonmail.com</a></td>\r\n          </tr>\r\n        </tbody>\r\n      </table>\r\n      <br>\r\n      <br>\r\n      <meta http-equiv=\"Content-Type\" content=\"text/html; charset=UTF-8\">\r\n      <meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\">\r\n      <meta name=\"format-detection\" content=\"telephone=no\">\r\n      <meta name=\"viewport\"\r\n        content=\"width=device-width, initial-scale=1.0\">\r\n      <title>Learn PDF Manipulation with Python - Our Latest Updated\r\n        Tutorials!</title>\r\n      <style type=\"text/css\" emogrify=\"no\">#outlook a{padding:0;}.ExternalClass{width:100%;}.ExternalClass, .ExternalClass p, .ExternalClass span, .ExternalClass font, .ExternalClass td, .ExternalClass div{line-height:100%;}table td{border-collapse:collapse;mso-line-height-rule:exactly;}.editable.image{font-size:0 !important;line-height:0 !important;}.nl2go_preheader{display:none !important;mso-hide:all !important;mso-line-height-rule:exactly;visibility:hidden !important;line-height:0px !important;font-size:0px !important;}body{width:100% !important;-webkit-text-size-adjust:100%;-ms-text-size-adjust:100%;margin:0;padding:0;}img{outline:none;text-decoration:none;-ms-interpolation-mode:bicubic;}a img{border:none;}table{border-collapse:collapse;mso-table-lspace:0pt;mso-table-rspace:0pt;}th{font-weight:normal;text-align:left;}*[class=\"gmail-fix\"]{display:none !important;}</style>\r\n      <style type=\"text/css\" emogrify=\"no\"></style>\r\n      <style type=\"text/css\" emogrify=\"no\"></style>\r\n      <style type=\"text/css\">p, h1, h2, h3, h4, ol, ul{margin:0;}a, a:link{color:#0092ff;text-decoration:underline}.nl2go-default-textstyle{color:#3b3f44;font-family:arial,helvetica,sans-serif;font-size:16px;line-height:1.5;word-break:break-word}.default-button{color:#ffffff;font-family:arial,helvetica,sans-serif;font-size:16px;font-style:normal;font-weight:bold;line-height:1.15;text-decoration:none;word-break:break-word}.default-heading1{color:#1F2D3D;font-family:arial,helvetica,sans-serif;font-size:36px;word-break:break-word}.default-heading2{color:#1F2D3D;font-family:arial,helvetica,sans-serif;font-size:32px;word-break:break-word}.default-heading3{color:#1F2D3D;font-family:arial,helvetica,sans-serif;font-size:24px;word-break:break-word}.default-heading4{color:#1F2D3D;font-family:arial,helvetica,sans-serif;font-size:18px;word-break:break-word}a[x-apple-data-detectors]{color:inherit !important;text-decoration:inherit !important;font-size:inherit !important;font-family:inherit !important;font-weight:inherit !important;line-height:inherit !important;}.no-show-for-you{border:none;display:none;float:none;font-size:0;height:0;line-height:0;max-height:0;mso-hide:all;overflow:hidden;table-layout:fixed;visibility:hidden;width:0;}</style><!--[if mso]><xml> <o:OfficeDocumentSettings> <o:AllowPNG/> <o:PixelsPerInch>96</o:PixelsPerInch> </o:OfficeDocumentSettings> </xml><![endif]-->\r\n      <style type=\"text/css\">a:link{color:#0092ff;text-decoration:underline}</style>\r\n      <table style=\" mso-hide:all;display:none\" cellspacing=\"0\"\r\n        cellpadding=\"0\" border=\"0\">\r\n        <tbody>\r\n          <tr>\r\n            <td>Learn how to extract tables from PDF, convert HTML to\r\n              PDF, and compress PDFs in\r\nPython͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0͏\u200c\u00a0</td>\r\n          </tr>\r\n        </tbody>\r\n      </table>\r\n      <!--[if mso]> <img width=\"1\" height=\"1\" src=\"https://9lctj.r.a.d.sendibm1.com/mk/op/sh/SMJz09hnzmooEakAYhJWCxug11WP/z0Gs0c1_y3H7\" /> <![endif]-->\r\n      <!--[if !mso]> <!----> <img style=\"display:none\"\r\nsrc=\"https://9lctj.r.a.d.sendibm1.com/mk/op/sh/SMJz09hnzmooEakAYhJWCxug11WP/z0Gs0c1_y3H7\"\r\n        moz-do-not-send=\"true\">\r\n      <!-- <![endif]-->\r\n      <table role=\"presentation\" class=\"nl2go-body-table\"\r\n        style=\"width:100%\" width=\"100%\" cellspacing=\"0\" cellpadding=\"0\"\r\n        border=\"0\">\r\n        <tbody>\r\n          <tr>\r\n            <td>\r\n              <table role=\"presentation\" class=\"r0-o\"\r\n                style=\"table-layout:fixed;width:600px\" width=\"600\"\r\n                cellspacing=\"0\" cellpadding=\"0\" border=\"0\"\r\n                align=\"center\">\r\n                <tbody>\r\n                  <tr>\r\n                    <td valign=\"top\">\r\n                      <table role=\"presentation\" class=\"r2-o\"\r\n                        style=\"table-layout:fixed;width:100%\"\r\n                        width=\"100%\" cellspacing=\"0\" cellpadding=\"0\"\r\n                        border=\"0\" align=\"center\">\r\n                        <tbody>\r\n                          <tr>\r\n                            <td class=\"r3-i\"\r\nstyle=\"background-color:#ffffff;padding-bottom:20px;padding-top:20px\">\r\n                              <table role=\"presentation\" width=\"100%\"\r\n                                cellspacing=\"0\" cellpadding=\"0\"\r\n                                border=\"0\">\r\n                                <tbody>\r\n                                  <tr>\r\n                                    <th class=\"r4-c\"\r\n                                      style=\"font-weight:normal\"\r\n                                      width=\"100%\" valign=\"top\">\r\n                                      <table role=\"presentation\"\r\n                                        class=\"r5-o\"\r\nstyle=\"table-layout:fixed;width:100%\" width=\"100%\" cellspacing=\"0\"\r\n                                        cellpadding=\"0\" border=\"0\">\r\n                                        <tbody>\r\n                                          <tr>\r\n                                            <td class=\"r6-i\"\r\nstyle=\"padding-left:15px;padding-right:15px\" valign=\"top\">\r\n                                              <table role=\"presentation\"\r\n                                                width=\"100%\"\r\n                                                cellspacing=\"0\"\r\n                                                cellpadding=\"0\"\r\n                                                border=\"0\">\r\n                                                <tbody>\r\n                                                  <tr>\r\n                                                    <td class=\"r7-c\">\r\n                                                      <table\r\nrole=\"presentation\" class=\"r5-o\" style=\"table-layout:fixed;width:570px\"\r\n                                                        width=\"570\"\r\n                                                        cellspacing=\"0\"\r\n                                                        cellpadding=\"0\"\r\n                                                        border=\"0\">\r\n                                                        <tbody>\r\n                                                          <tr>\r\n                                                          <td\r\nclass=\"r8-i nl2go-default-textstyle\"\r\nstyle=\"color:#3b3f44;font-family:arial,helvetica,sans-serif;font-size:16px;line-height:1.5;word-break:break-word;padding-bottom:15px;padding-top:15px\">\r\n                                                          <h1\r\nstyle=\"margin:0;font-family:Arial;color:#000000;margin-top:0;font-weight:400\"><img\r\nalt=\"Python Logo\"\r\nsrc=\"https://9lctj.img.a.d.sendibm1.com/im/sh/2fMMxBUOJlpV.png?u=7xwQLFBtniwQn1MAnaHPuN5TE0tZRUj\"\r\ntitle=\"Python Logo\" class=\"CToWUd\" data-bit=\"iit\"\r\nstyle=\"display:block;width:50px;height:50px;float:left;margin-right:5px;padding-top:5px\"\r\n                                                          sib_img_id=\"0\"\r\nmoz-do-not-send=\"true\" width=\"40\" height=\"50\"><a\r\nhref=\"https://9lctj.r.a.d.sendibm1.com/mk/cl/f/sh/SMK1E8tHeFuBoQMC4j632rkg8dRl/UQ4s6dPsmEFK\"\r\ntarget=\"_blank\" style=\"text-decoration:none;color:#2f89fc\"\r\nsib_link_id=\"0\" templating=\"n\" moz-do-not-send=\"true\">The Python Code</a></h1>\r\n                                                          </td>\r\n                                                          </tr>\r\n                                                        </tbody>\r\n                                                      </table>\r\n                                                    </td>\r\n                                                  </tr>\r\n                                                </tbody>\r\n                                              </table>\r\n                                            </td>\r\n                                          </tr>\r\n                                        </tbody>\r\n                                      </table>\r\n                                    </th>\r\n                                  </tr>\r\n                                </tbody>\r\n                              </table>\r\n                            </td>\r\n                          </tr>\r\n                        </tbody>\r\n                      </table>\r\n                      <table role=\"presentation\" class=\"r2-o\"\r\n                        style=\"table-layout:fixed;width:100%\"\r\n                        width=\"100%\" cellspacing=\"0\" cellpadding=\"0\"\r\n                        border=\"0\" align=\"center\">\r\n                        <tbody>\r\n                          <tr>\r\n                            <td class=\"r9-i\"\r\nstyle=\"background-color:#ffffff;padding-bottom:20px;padding-top:20px\">\r\n                              <table role=\"presentation\" width=\"100%\"\r\n                                cellspacing=\"0\" cellpadding=\"0\"\r\n                                border=\"0\">\r\n                                <tbody>\r\n                                  <tr>\r\n                                    <th class=\"r4-c\"\r\n                                      style=\"font-weight:normal\"\r\n                                      width=\"100%\" valign=\"top\">\r\n                                      <table role=\"presentation\"\r\n                                        class=\"r5-o\"\r\nstyle=\"table-layout:fixed;width:100%\" width=\"100%\" cellspacing=\"0\"\r\n                                        cellpadding=\"0\" border=\"0\">\r\n                                        <tbody>\r\n                                          <tr>\r\n                                            <td class=\"r6-i\"\r\n                                              valign=\"top\">\r\n                                              <table role=\"presentation\"\r\n                                                width=\"100%\"\r\n                                                cellspacing=\"0\"\r\n                                                cellpadding=\"0\"\r\n                                                border=\"0\">\r\n                                                <tbody>\r\n                                                  <tr>\r\n                                                    <td class=\"r10-c\"\r\n                                                      align=\"left\">\r\n                                                      <table\r\nrole=\"presentation\" class=\"r11-o\" style=\"table-layout:fixed;width:100%\"\r\n                                                        width=\"100%\"\r\n                                                        cellspacing=\"0\"\r\n                                                        cellpadding=\"0\"\r\n                                                        border=\"0\">\r\n                                                        <tbody>\r\n                                                          <tr>\r\n                                                          <td\r\nclass=\"r12-i nl2go-default-textstyle\"\r\nstyle=\"color:#3b3f44;font-family:arial,helvetica,sans-serif;font-size:16px;line-height:1.5;word-break:break-word;padding-top:15px;text-align:left\"\r\n                                                          valign=\"top\"\r\n                                                          align=\"left\">\r\n                                                          <div>\r\n                                                          <h1\r\nclass=\"default-heading1\"\r\nstyle=\"margin:0;color:#1f2d3d;font-family:arial,helvetica,sans-serif;font-size:36px;word-break:break-word\"><span\r\nstyle=\"font-size:28px\">Discover Our PDF Manipulation Tutorials in Python</span></h1>\r\n                                                          </div>\r\n                                                          </td>\r\n                                                          </tr>\r\n                                                        </tbody>\r\n                                                      </table>\r\n                                                    </td>\r\n                                                  </tr>\r\n                                                  <tr>\r\n                                                    <td class=\"r10-c\"\r\n                                                      align=\"left\">\r\n                                                      <table\r\nrole=\"presentation\" class=\"r11-o\" style=\"table-layout:fixed;width:100%\"\r\n                                                        width=\"100%\"\r\n                                                        cellspacing=\"0\"\r\n                                                        cellpadding=\"0\"\r\n                                                        border=\"0\">\r\n                                                        <tbody>\r\n                                                          <tr>\r\n                                                          <td\r\nclass=\"r12-i nl2go-default-textstyle\"\r\nstyle=\"color:#3b3f44;font-family:arial,helvetica,sans-serif;font-size:16px;line-height:1.5;word-break:break-word;padding-top:15px;text-align:left\"\r\n                                                          valign=\"top\"\r\n                                                          align=\"left\">\r\n                                                          <div>\r\n                                                          <p\r\nstyle=\"margin:0\"><span style=\"font-size:16px\">Hey there,</span></p>\r\n                                                          <p\r\nstyle=\"margin:0\">\u00a0</p>\r\n                                                          <p\r\nstyle=\"margin:0\"><span style=\"font-size:16px\">In this newsletter, we're\r\n                                                          sharing our\r\n                                                          latest updated\r\n                                                          PDF\r\n                                                          Manipulation\r\n                                                          tutorials:</span></p>\r\n                                                          <h2\r\nclass=\"default-heading2\"\r\nstyle=\"margin:0;color:#1f2d3d;font-family:arial,helvetica,sans-serif;font-size:32px;word-break:break-word\"><span\r\nstyle=\"font-family:Arial;font-size:24px\">1. How to Extract Tables from\r\n                                                          PDF in Python</span></h2>\r\n                                                          <p\r\nstyle=\"margin:0\">In this tutorial, you will learn how to extract tables\r\n                                                          from PDF files\r\n                                                          in Python\r\n                                                          using camelot\r\n                                                          and tabula\r\n                                                          libraries and\r\n                                                          export them\r\n                                                          into several\r\n                                                          formats such\r\n                                                          as CSV, excel,\r\n                                                          Pandas\r\n                                                          dataframe and\r\n                                                          HTML.</p>\r\n                                                          <p\r\nstyle=\"margin:0\">\u00a0</p>\r\n                                                          <p\r\nstyle=\"margin:0\">Check it out: <a\r\nhref=\"https://9lctj.r.a.d.sendibm1.com/mk/cl/f/sh/SMK1E8tHeG13GZB9FtFX8h3Q6w21/Gl7QYYQZQSpp\"\r\ntitle=\"How to Extract Tables from PDF in Python\" target=\"_blank\"\r\nstyle=\"color:#0092ff;text-decoration:underline\" sib_link_id=\"1\"\r\n                                                          templating=\"n\"\r\nmoz-do-not-send=\"true\">How to Extract Tables from PDF in Python</a><br>\r\n                                                          \u00a0</p>\r\n                                                          <h2\r\nclass=\"default-heading2\"\r\nstyle=\"margin:0;color:#1f2d3d;font-family:arial,helvetica,sans-serif;font-size:32px;word-break:break-word\"><span\r\nstyle=\"font-size:24px\">2. How to Convert HTML to PDF in Python</span></h2>\r\n                                                          <p\r\nstyle=\"margin:0\">Learn how you can convert HTML pages to PDF files from\r\n                                                          an HTML file,\r\n                                                          URL or even\r\n                                                          HTML content\r\n                                                          string using\r\n                                                          wkhtmltopdf\r\n                                                          tool and its\r\n                                                          pdfkit wrapper\r\n                                                          in Python.</p>\r\n                                                          <p\r\nstyle=\"margin:0\">\u00a0</p>\r\n                                                          <p\r\nstyle=\"margin:0\">Check it out: <a\r\nhref=\"https://9lctj.r.a.d.sendibm1.com/mk/cl/f/sh/SMK1E8tHeG7uii06R3P1EWMA5EcH/ome-rKPwDS0B\"\r\ntitle=\"How to Convert HTML to PDF in Python\" target=\"_blank\"\r\nstyle=\"color:#0092ff;text-decoration:underline\" sib_link_id=\"2\"\r\n                                                          templating=\"n\"\r\nmoz-do-not-send=\"true\">How to Convert HTML to PDF in Python</a></p>\r\n                                                          <p\r\nstyle=\"margin:0\">\u00a0</p>\r\n                                                          <h2\r\nclass=\"default-heading2\"\r\nstyle=\"margin:0;color:#1f2d3d;font-family:arial,helvetica,sans-serif;font-size:32px;word-break:break-word\"><span\r\nstyle=\"font-size:24px\">3. How to Compress PDF Files in Python</span></h2>\r\n                                                          <p\r\nstyle=\"margin:0\">Compressing PDF allows you to decrease the file size as\r\n                                                          small as\r\n                                                          possible while\r\n                                                          maintaining\r\n                                                          the quality of\r\n                                                          the media in\r\n                                                          that PDF file.\r\n                                                          As a result,\r\n                                                          it\r\n                                                          significantly\r\n                                                          increases\r\n                                                          effectiveness\r\n                                                          and\r\n                                                          shareability.</p>\r\n                                                          <p\r\nstyle=\"margin:0\">\u00a0</p>\r\n                                                          <p\r\nstyle=\"margin:0\">In this tutorial, you will learn how to compress PDF\r\n                                                          files using\r\n                                                          the PDFTron\r\n                                                          library in\r\n                                                          Python.</p>\r\n                                                          <p\r\nstyle=\"margin:0\">\u00a0</p>\r\n                                                          <p\r\nstyle=\"margin:0\">Check it out: <a\r\nhref=\"https://9lctj.r.a.d.sendibm1.com/mk/cl/f/sh/SMK1E8tHeGEmAqp3cDYVKLeu3XCX/hHhL6aV2X9MM\"\r\ntitle=\"How to Compress PDF Files in Python\" target=\"_blank\"\r\nstyle=\"color:#0092ff;text-decoration:underline\" sib_link_id=\"3\"\r\n                                                          templating=\"n\"\r\nmoz-do-not-send=\"true\">How to Compress PDF Files in Python</a></p>\r\n                                                          </div>\r\n                                                          </td>\r\n                                                          </tr>\r\n                                                        </tbody>\r\n                                                      </table>\r\n                                                    </td>\r\n                                                  </tr>\r\n                                                  <tr>\r\n                                                    <td class=\"r13-c\"\r\n                                                      align=\"center\">\r\n                                                      <table\r\nrole=\"presentation\" class=\"r2-o\" style=\"table-layout:fixed\" width=\"600\"\r\n                                                        cellspacing=\"0\"\r\n                                                        cellpadding=\"0\"\r\n                                                        border=\"0\">\r\n                                                        <tbody>\r\n                                                          <tr>\r\n                                                          <td\r\n                                                          class=\"r14-i\"\r\nstyle=\"padding-bottom:30px;padding-top:30px;height:2px\">\r\n                                                          <table\r\nrole=\"presentation\" width=\"100%\" cellspacing=\"0\" cellpadding=\"0\"\r\n                                                          border=\"0\">\r\n                                                          <tbody>\r\n                                                          <tr>\r\n                                                          <td>\r\n                                                          <table\r\nrole=\"presentation\" valign=\"\" class=\"r14-i\"\r\nstyle=\"border-top-style:solid;background-clip:border-box;border-top-color:#4A4A4A;border-top-width:2px;font-size:2px;line-height:2px\"\r\n                                                          width=\"100%\"\r\n                                                          height=\"2\"\r\ncellspacing=\"0\" cellpadding=\"0\" border=\"0\">\r\n                                                          <tbody>\r\n                                                          <tr>\r\n                                                          <td\r\nstyle=\"font-size:0px;line-height:0px\" height=\"0\">\u00ad</td>\r\n                                                          </tr>\r\n                                                          </tbody>\r\n                                                          </table>\r\n                                                          </td>\r\n                                                          </tr>\r\n                                                          </tbody>\r\n                                                          </table>\r\n                                                          </td>\r\n                                                          </tr>\r\n                                                        </tbody>\r\n                                                      </table>\r\n                                                    </td>\r\n                                                  </tr>\r\n                                                  <tr>\r\n                                                    <td class=\"r10-c\"\r\n                                                      align=\"left\">\r\n                                                      <table\r\nrole=\"presentation\" class=\"r11-o\" style=\"table-layout:fixed;width:100%\"\r\n                                                        width=\"100%\"\r\n                                                        cellspacing=\"0\"\r\n                                                        cellpadding=\"0\"\r\n                                                        border=\"0\">\r\n                                                        <tbody>\r\n                                                          <tr>\r\n                                                          <td\r\nclass=\"r15-i nl2go-default-textstyle\"\r\nstyle=\"color:#3b3f44;font-family:arial,helvetica,sans-serif;font-size:16px;line-height:1.5;word-break:break-word;padding-bottom:15px;padding-top:15px;text-align:left\"\r\n                                                          valign=\"top\"\r\n                                                          align=\"left\">\r\n                                                          <div>\r\n                                                          <p\r\nstyle=\"margin:0\">Additionally, if you find our tutorials beneficial, you\r\n                                                          might want to\r\n                                                          delve deeper\r\n                                                          with our <a\r\nhref=\"https://9lctj.r.a.d.sendibm1.com/mk/cl/f/sh/SMK1E8tHeGLdcze0nNhzQAxe1pmn/dVD0Qk_vHsHp\"\r\ntitle=\"Practical Python PDF Processing EBook\" target=\"_blank\"\r\nstyle=\"color:#0092ff;text-decoration:underline\" sib_link_id=\"4\"\r\n                                                          templating=\"n\"\r\nmoz-do-not-send=\"true\">Practical Python PDF Processing eBook</a>. This\r\n                                                          eBook is a\r\n                                                          treasure trove\r\n                                                          for those\r\n                                                          eager to\r\n                                                          master the art\r\n                                                          of PDF\r\n                                                          processing\r\n                                                          using Python.\r\n                                                          With it,\r\n                                                          you'll learn\r\n                                                          to create,\r\n                                                          read, write,\r\n                                                          and manipulate\r\n                                                          PDFs, diving\r\n                                                          into\r\n                                                          real-world\r\n                                                          projects that\r\n                                                          demonstrate\r\n                                                          the power of\r\n                                                          Python in\r\n                                                          handling PDF\r\n                                                          operations\r\n                                                          efficiently.\r\n                                                          To sweeten the\r\n                                                          deal, use the\r\n                                                          code <strong>SUBSCRIBER15</strong>\r\n                                                          at checkout to\r\n                                                          snag a <strong>15%</strong>\r\n                                                          discount!</p>\r\n                                                          <p\r\nstyle=\"margin:0\">\u00a0</p>\r\n                                                          <p\r\nstyle=\"margin:0\">Check it out: <a\r\nhref=\"https://9lctj.r.a.d.sendibm1.com/mk/cl/f/sh/SMK1E8tHeGSV58SxyXrTW0GO08N3/egdZL1lftisC\"\r\ntitle=\"Practical Python PDF Processing eBook\" target=\"_blank\"\r\nstyle=\"color:#0092ff;text-decoration:underline\" sib_link_id=\"5\"\r\n                                                          templating=\"n\"\r\nmoz-do-not-send=\"true\">Practical Python PDF Processing eBook.</a></p>\r\n                                                          <p\r\nstyle=\"margin:0\">\u00a0</p>\r\n                                                          <p\r\nstyle=\"margin:0\">If you have any questions, please reply to this email\r\n                                                          as I reply to\r\n                                                          every email,\r\n                                                          just give me\r\n                                                          some time!</p>\r\n                                                          <p\r\nstyle=\"margin:0\">\u00a0</p>\r\n                                                          <p\r\nstyle=\"margin:0\">All the best,</p>\r\n                                                          <p\r\nstyle=\"margin:0\">Abdou @ The Python Code</p>\r\n                                                          </div>\r\n                                                          </td>\r\n                                                          </tr>\r\n                                                        </tbody>\r\n                                                      </table>\r\n                                                    </td>\r\n                                                  </tr>\r\n                                                </tbody>\r\n                                              </table>\r\n                                            </td>\r\n                                          </tr>\r\n                                        </tbody>\r\n                                      </table>\r\n                                    </th>\r\n                                  </tr>\r\n                                </tbody>\r\n                              </table>\r\n                            </td>\r\n                          </tr>\r\n                        </tbody>\r\n                      </table>\r\n                      <table role=\"presentation\" class=\"r2-o\"\r\n                        style=\"table-layout:fixed;width:100%\"\r\n                        width=\"100%\" cellspacing=\"0\" cellpadding=\"0\"\r\n                        border=\"0\" align=\"center\">\r\n                        <tbody>\r\n                          <tr>\r\n                            <td class=\"r16-i\"\r\nstyle=\"background-color:#eff2f7;padding-bottom:20px;padding-top:20px\">\r\n                              <table role=\"presentation\" width=\"100%\"\r\n                                cellspacing=\"0\" cellpadding=\"0\"\r\n                                border=\"0\">\r\n                                <tbody>\r\n                                  <tr>\r\n                                    <th class=\"r4-c\"\r\n                                      style=\"font-weight:normal\"\r\n                                      width=\"100%\" valign=\"top\">\r\n                                      <table role=\"presentation\"\r\n                                        class=\"r5-o\"\r\nstyle=\"table-layout:fixed;width:100%\" width=\"100%\" cellspacing=\"0\"\r\n                                        cellpadding=\"0\" border=\"0\">\r\n                                        <tbody>\r\n                                          <tr>\r\n                                            <td class=\"r6-i\"\r\nstyle=\"padding-left:15px;padding-right:15px\" valign=\"top\">\r\n                                              <table role=\"presentation\"\r\n                                                width=\"100%\"\r\n                                                cellspacing=\"0\"\r\n                                                cellpadding=\"0\"\r\n                                                border=\"0\">\r\n                                                <tbody>\r\n                                                  <tr>\r\n                                                    <td class=\"r10-c\"\r\n                                                      align=\"left\">\r\n                                                      <table\r\nrole=\"presentation\" class=\"r11-o\" style=\"table-layout:fixed;width:100%\"\r\n                                                        width=\"100%\"\r\n                                                        cellspacing=\"0\"\r\n                                                        cellpadding=\"0\"\r\n                                                        border=\"0\">\r\n                                                        <tbody>\r\n                                                          <tr>\r\n                                                          <td\r\nclass=\"r17-i nl2go-default-textstyle\"\r\nstyle=\"font-family:arial,helvetica,sans-serif;word-break:break-word;color:#3b3f44;font-size:18px;line-height:1.5;padding-top:15px;text-align:center\"\r\n                                                          valign=\"top\"\r\n                                                          align=\"center\">\r\n                                                          <div>\r\n                                                          <p\r\nstyle=\"margin:0\"><strong>The Python Code</strong></p>\r\n                                                          </div>\r\n                                                          </td>\r\n                                                          </tr>\r\n                                                        </tbody>\r\n                                                      </table>\r\n                                                    </td>\r\n                                                  </tr>\r\n                                                  <tr>\r\n                                                    <td class=\"r10-c\"\r\n                                                      align=\"left\">\r\n                                                      <table\r\nrole=\"presentation\" class=\"r11-o\" style=\"table-layout:fixed;width:100%\"\r\n                                                        width=\"100%\"\r\n                                                        cellspacing=\"0\"\r\n                                                        cellpadding=\"0\"\r\n                                                        border=\"0\">\r\n                                                        <tbody>\r\n                                                          <tr>\r\n                                                          <td\r\nclass=\"r18-i nl2go-default-textstyle\"\r\nstyle=\"font-family:arial,helvetica,sans-serif;word-break:break-word;color:#3b3f44;font-size:18px;line-height:1.5;text-align:center\"\r\n                                                          valign=\"top\"\r\n                                                          align=\"center\">\r\n                                                          <div>\r\n                                                          <p\r\nstyle=\"margin:0;font-size:14px\">Constantine, Algeria</p>\r\n                                                          </div>\r\n                                                          </td>\r\n                                                          </tr>\r\n                                                        </tbody>\r\n                                                      </table>\r\n                                                    </td>\r\n                                                  </tr>\r\n                                                  <tr>\r\n                                                    <td class=\"r10-c\"\r\n                                                      align=\"left\">\r\n                                                      <table\r\nrole=\"presentation\" class=\"r11-o\" style=\"table-layout:fixed;width:100%\"\r\n                                                        width=\"100%\"\r\n                                                        cellspacing=\"0\"\r\n                                                        cellpadding=\"0\"\r\n                                                        border=\"0\">\r\n                                                        <tbody>\r\n                                                          <tr>\r\n                                                          <td\r\nclass=\"r17-i nl2go-default-textstyle\"\r\nstyle=\"font-family:arial,helvetica,sans-serif;word-break:break-word;color:#3b3f44;font-size:18px;line-height:1.5;padding-top:15px;text-align:center\"\r\n                                                          valign=\"top\"\r\n                                                          align=\"center\">\r\n                                                          <div>\r\n                                                          <p\r\nstyle=\"margin:0;font-size:14px\">This email was sent to\r\n                                                          <a class=\"moz-txt-link-abbreviated\" href=\"mailto:gorgitesting3@protonmail.com\">gorgitesting3@protonmail.com</a></p>\r\n                                                          </div>\r\n                                                          </td>\r\n                                                          </tr>\r\n                                                        </tbody>\r\n                                                      </table>\r\n                                                    </td>\r\n                                                  </tr>\r\n                                                  <tr>\r\n                                                    <td class=\"r10-c\"\r\n                                                      align=\"left\">\r\n                                                      <table\r\nrole=\"presentation\" class=\"r11-o\" style=\"table-layout:fixed;width:100%\"\r\n                                                        width=\"100%\"\r\n                                                        cellspacing=\"0\"\r\n                                                        cellpadding=\"0\"\r\n                                                        border=\"0\">\r\n                                                        <tbody>\r\n                                                          <tr>\r\n                                                          <td\r\nclass=\"r18-i nl2go-default-textstyle\"\r\nstyle=\"font-family:arial,helvetica,sans-serif;word-break:break-word;color:#3b3f44;font-size:18px;line-height:1.5;text-align:center\"\r\n                                                          valign=\"top\"\r\n                                                          align=\"center\">\r\n                                                          <div>\r\n                                                          <p\r\nstyle=\"margin:0;font-size:14px\">You've received it because you've\r\n                                                          subscribed to\r\n                                                          our\r\n                                                          newsletter.</p>\r\n                                                          </div>\r\n                                                          </td>\r\n                                                          </tr>\r\n                                                        </tbody>\r\n                                                      </table>\r\n                                                    </td>\r\n                                                  </tr>\r\n                                                  <tr>\r\n                                                    <td class=\"r10-c\"\r\n                                                      align=\"left\">\r\n                                                      <table\r\nrole=\"presentation\" class=\"r11-o\" style=\"table-layout:fixed;width:100%\"\r\n                                                        width=\"100%\"\r\n                                                        cellspacing=\"0\"\r\n                                                        cellpadding=\"0\"\r\n                                                        border=\"0\">\r\n                                                        <tbody>\r\n                                                          <tr>\r\n                                                          <td\r\nclass=\"r19-i nl2go-default-textstyle\"\r\nstyle=\"font-family:arial,helvetica,sans-serif;word-break:break-word;color:#3b3f44;font-size:18px;line-height:1.5;padding-bottom:15px;padding-top:15px;text-align:center\"\r\n                                                          valign=\"top\"\r\n                                                          align=\"center\">\r\n                                                          <div>\r\n                                                          <p\r\nstyle=\"margin:0;font-size:14px\"><a\r\nhref=\"https://9lctj.r.a.d.sendibm1.com/mk/mr/sh/SMJz09SDriOHWPmp1CzvvO8RVcI9/mffAkn_riJ4q\"\r\nstyle=\"color:#0092ff;text-decoration:underline\" moz-do-not-send=\"true\">View\r\n                                                          in browser</a>\r\n                                                          | <a\r\nhref=\"https://9lctj.r.a.d.sendibm1.com/mk/un/sh/SMJz09a0vkbXsVGUmx9j4B1YlJuH/kbXFjM4lSqgf\"\r\nstyle=\"color:#0092ff;text-decoration:underline\" moz-do-not-send=\"true\">Unsubscribe</a></p>\r\n                                                          </div>\r\n                                                          </td>\r\n                                                          </tr>\r\n                                                        </tbody>\r\n                                                      </table>\r\n                                                    </td>\r\n                                                  </tr>\r\n                                                </tbody>\r\n                                              </table>\r\n                                            </td>\r\n                                          </tr>\r\n                                        </tbody>\r\n                                      </table>\r\n                                    </th>\r\n                                  </tr>\r\n                                </tbody>\r\n                              </table>\r\n                            </td>\r\n                          </tr>\r\n                        </tbody>\r\n                      </table>\r\n                    </td>\r\n                  </tr>\r\n                </tbody>\r\n              </table>\r\n            </td>\r\n          </tr>\r\n        </tbody>\r\n      </table>\r\n    </div>\r\n  </body>\r\n</html>"
        }
      }
      """

  Scenario: HTML message with inline HTML and HTML attachment encoded in UTF-8
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      Content-Type: multipart/mixed; boundary="------------2p04vJsuXgcobQxmsvuPsEB2"
      To: <[user:to]@[domain]>
      From: <[user:user]@[domain]>
      Subject: HTML message with inline HTML and HTML attachment encoded in UTF-8

      This is a multi-part message in MIME format.
      --------------2p04vJsuXgcobQxmsvuPsEB2
      Content-Type: text/html; charset=UTF-8
      Content-Transfer-Encoding: 7bit

      <!DOCTYPE html>
      <html>
        <head>

          <meta http-equiv="content-type" content="text/html; charset=UTF-8">
        </head>
        <body>
          <p>Hello, this is a <b>HTML message</b> with <i>HTML attachment</i>.<br>
          </p>
        </body>
      </html>
      --------------2p04vJsuXgcobQxmsvuPsEB2
      Content-Type: text/html; charset=UTF-8; name="index.html"
      Content-Disposition: attachment; filename="index.html"
      Content-Transfer-Encoding: base64

      PCFET0NUWVBFIGh0bWw+
      --------------2p04vJsuXgcobQxmsvuPsEB2--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject                                                            |
      | [user:user]@[domain] | [user:to]@[domain] | HTML message with inline HTML and HTML attachment encoded in UTF-8 |
    And IMAP client "1" eventually sees 1 messages in "Sent"
    When the user logs in with username "[user:to]" and password "password"
    And user "[user:to]" connects and authenticates IMAP client "2"
    And user "[user:to]" finishes syncing
    And it succeeds
    Then IMAP client "2" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "[user:user]@[domain]",
        "to": "[user:to]@[domain]",
        "subject": "HTML message with inline HTML and HTML attachment encoded in UTF-8",
        "content": {
          "content-type": "multipart/mixed",
          "sections":[
            {
              "content-type": "text/html",
              "content-type-charset": "utf-8",
              "transfer-encoding": "quoted-printable",
              "body-is": "<!DOCTYPE html>\r\n<html>\r\n  <head>\r\n\r\n    <meta http-equiv=3D\"content-type\" content=3D\"text/html; charset=3DUTF-8=\r\n\">\r\n  </head>\r\n  <body>\r\n    <p>Hello, this is a <b>HTML message</b> with <i>HTML attachment</i>.<br=\r\n>\r\n    </p>\r\n  </body>\r\n</html>"
            },
            {
              "content-type": "text/html",
              "content-type-charset": "UTF-8",
              "content-type-name": "index.html",
              "content-disposition": "attachment",
              "content-disposition-filename": "index.html",
              "transfer-encoding": "base64",
              "body-is": "PCFET0NUWVBFIGh0bWw+"
            }
          ]
        }
      }
      """

  Scenario: HTML msg with inline HTML and HTML attachment not encoded in UTF-8
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      Content-Type: multipart/mixed; boundary="------------2p04vJsuXgcobQxmsvuPsEB2"
      To: <[user:to]@[domain]>
      From: <[user:user]@[domain]>
      Subject: HTML msg with inline HTML and HTML attachment not encoded in UTF-8

      This is a multi-part message in MIME format.
      --------------2p04vJsuXgcobQxmsvuPsEB2
      Content-Type: text/html; charset=UTF-7
      Content-Transfer-Encoding: 7bit

      <!DOCTYPE html><meta http-equiv="content-type" content="text/html; charset=UTF-7">
      --------------2p04vJsuXgcobQxmsvuPsEB2
      Content-Type: text/html; charset=UTF-7; name="index.html"
      Content-Disposition: attachment; filename="index.html"
      Content-Transfer-Encoding: base64

      PCFET0NUWVBFIGh0bWw+CjxtZXRhIGh0dHAtZXF1aXY9ImNvbnRlbnQtdHlwZSIgY29udGVu
      dD0idGV4dC9odG1sOyBjaGFyc2V0PVVURi03Ij4=

      --------------2p04vJsuXgcobQxmsvuPsEB2-- 
      
      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject                                            |
      | [user:user]@[domain] | [user:to]@[domain] | HTML msg with inline HTML and HTML attachment not encoded in UTF-8 |
    And IMAP client "1" eventually sees 1 messages in "Sent"
    When the user logs in with username "[user:to]" and password "password"
    And user "[user:to]" connects and authenticates IMAP client "2"
    And user "[user:to]" finishes syncing
    And it succeeds
    Then IMAP client "2" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "[user:user]@[domain]",
        "to": "[user:to]@[domain]",
        "subject": "HTML msg with inline HTML and HTML attachment not encoded in UTF-8",
        "content": {
          "content-type": "multipart/mixed",
          "sections":[
            {
              "content-type": "text/html",
              "content-type-charset": "utf-8",
              "transfer-encoding": "quoted-printable",
              "body-is": "<!DOCTYPE html><html><head><meta http-equiv=3D\"content-type\" content=3D\"tex=\r\nt/html; charset=3DUTF-8\"/></head><body></body></html>"
            },
            {
              "content-type": "text/html",
              "content-type-charset": "UTF-8",
              "content-type-name": "index.html",
              "content-disposition": "attachment",
              "content-disposition-filename": "index.html",
              "transfer-encoding": "base64",
              "body-is": "PCFET0NUWVBFIGh0bWw+PGh0bWw+PGhlYWQ+PG1ldGEgaHR0cC1lcXVpdj0iY29udGVudC10eXBl\r\nIiBjb250ZW50PSJ0ZXh0L2h0bWw7IGNoYXJzZXQ9VVRGLTgiLz48L2hlYWQ+PGJvZHk+PC9ib2R5\r\nPjwvaHRtbD4="
            }
          ]
        }
      }
      """

  Scenario: HTML message and attachment not encoded in UTF-8 and without meta charset
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      Content-Type: multipart/mixed; boundary="------------2p04vJsuXgcobQxmsvuPsEB2"
      To: <[user:to]@[domain]>
      From: <[user:user]@[domain]>
      Subject: HTML message and attachment not encoded in UTF-8 and without meta charset

      This is a multi-part message in MIME format.
      --------------2p04vJsuXgcobQxmsvuPsEB2
      Content-Type: text/html; charset=UTF-7
      Content-Transfer-Encoding: 7bit

      <!DOCTYPE html>
      
      --------------2p04vJsuXgcobQxmsvuPsEB2
      Content-Type: text/html; charset=UTF-7; name="index.html"
      Content-Disposition: attachment; filename="index.html"
      Content-Transfer-Encoding: base64

      PCFET0NUWVBFIGh0bWw+
      --------------2p04vJsuXgcobQxmsvuPsEB2--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject                                                               |
      | [user:user]@[domain] | [user:to]@[domain] | HTML message and attachment not encoded in UTF-8 and without meta charset |
    And IMAP client "1" eventually sees 1 messages in "Sent"
    When the user logs in with username "[user:to]" and password "password"
    And user "[user:to]" connects and authenticates IMAP client "2"
    And user "[user:to]" finishes syncing
    And it succeeds
    Then IMAP client "2" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "[user:user]@[domain]",
        "to": "[user:to]@[domain]",
        "subject": "HTML message and attachment not encoded in UTF-8 and without meta charset",
        "content": {
          "content-type": "multipart/mixed",
          "sections":[
            {
              "content-type": "text/html",
              "content-type-charset": "utf-8",
              "transfer-encoding": "quoted-printable",
              "body-is": "<!DOCTYPE html>"
            },
            {
              "content-type": "text/html",
              "content-type-charset": "UTF-8",
              "content-type-name": "index.html",
              "content-disposition": "attachment",
              "content-disposition-filename": "index.html",
              "transfer-encoding": "base64",
              "body-is": "PCFET0NUWVBFIGh0bWw+"
            }
          ]
        }
      }
      """
    
