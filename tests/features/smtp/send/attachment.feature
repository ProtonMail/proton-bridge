Feature: SMTP sending with attachment
  Background:
    Given there exists an account with username "[user:user1]" and password "password"
    And there exists an account with username "[user:user2]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user1]" and password "password"
    And user "[user:user1]" finishes syncing
    Then it succeeds
    When user "[user:user1]" connects and authenticates SMTP client "1"
    And user "[user:user1]" connects and authenticates IMAP client "1"
    Then it succeeds

  @long-black
  Scenario: Sending with cyrillic PDF attachment
    When SMTP client "1" sends the following message from "[user:user1]@[domain]" to "[user:user2]@[domain]":
      """
      Content-Type: multipart/mixed; boundary="------------bYzsV6z0EdKTbltmCDZgIM15"
      From: Bridge Test <[user:user1]@[domain]>
      To: Internal Bridge <[user:user2]@[domain]>
      Subject: Test with cyrillic attachment

      --------------bYzsV6z0EdKTbltmCDZgIM15
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/plain; charset=utf-8

      Shake that body
      --------------bYzsV6z0EdKTbltmCDZgIM15
      Content-Type: application/pdf;
       name="=?UTF-8?B?0JDQkdCS0JPQlNCD0JXQltCX0IXQmNCI0JrQm9CJ0JzQndCK0J7Qn9Cg?=
       =?UTF-8?B?0KHQotCM0KPQpNCl0KfQj9CX0KgucGRm?="
      Content-Disposition: attachment;
       filename*0*=UTF-8''%D0%90%D0%91%D0%92%D0%93%D0%94%D0%83%D0%95%D0%96%D0%97;
       filename*1*=%D0%85%D0%98%D0%88%D0%9A%D0%9B%D0%89%D0%9C%D0%9D%D0%8A%D0%9E;
       filename*2*=%D0%9F%D0%A0%D0%A1%D0%A2%D0%8C%D0%A3%D0%A4%D0%A5%D0%A7%D0%8F;
       filename*3*=%D0%97%D0%A8%2E%70%64%66
      Content-Transfer-Encoding: base64

      0JDQkdCS0JPQlNCD0JXQltCX0IXQmNCI0JrQm9CJ0JzQndCK0J7Qn9Cg0KHQotCM0KPQpNCl0KfQj9CX0Kg=

      --------------bYzsV6z0EdKTbltmCDZgIM15--

      """
    Then it succeeds
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                  | to                    | subject                       |
      | [user:user1]@[domain] | [user:user2]@[domain] | Test with cyrillic attachment |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "Test with cyrillic attachment",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "[user:user2]@[domain]",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [],
          "MIMEType": "text/plain"
        }
      }
      """
    And the body in the "POST" response to "/mail/v4/attachments" is:
      """
      {
        "Attachment":{
          "Name": "АБВГДЃЕЖЗЅИЈКЛЉМНЊОПРСТЌУФХЧЏЗШ.pdf",
          "MIMEType": "application/pdf",
          "Disposition": "attachment"
          }
      }
      """


  @long-black
  Scenario: Sending with cyrillic docx attachment
    When SMTP client "1" sends the following message from "[user:user1]@[domain]" to "[user:user2]@[domain]":
      """
      Content-Type: multipart/mixed; boundary="------------9xfXriG1c1v5iJlMiIMCaIWP"
      From: Bridge Test <[user:user1]@[domain]>
      To: Internal Bridge <[user:user2]@[domain]>
      Subject: Test with cyrillic attachment

      --------------9xfXriG1c1v5iJlMiIMCaIWP
      Content-Type: text/plain; charset=UTF-8; format=flowed
      Content-Transfer-Encoding: 7bit

      Shake that body
      --------------9xfXriG1c1v5iJlMiIMCaIWP
      Content-Type: application/vnd.openxmlformats-officedocument.wordprocessingml.document;
       name="=?UTF-8?B?0JDQkdCS0JPQlNCD0JXQltCX0IXQmNCI0JrQm9CJ0JzQndCK0J7Qn9Cg?=
       =?UTF-8?B?0KHQotCM0KPQpNCl0KfQj9CX0KguZG9jeA==?="
      Content-Disposition: attachment;
       filename*0*=UTF-8''%D0%90%D0%91%D0%92%D0%93%D0%94%D0%83%D0%95%D0%96%D0%97;
       filename*1*=%D0%85%D0%98%D0%88%D0%9A%D0%9B%D0%89%D0%9C%D0%9D%D0%8A%D0%9E;
       filename*2*=%D0%9F%D0%A0%D0%A1%D0%A2%D0%8C%D0%A3%D0%A4%D0%A5%D0%A7%D0%8F;
       filename*3*=%D0%97%D0%A8%2E%64%6F%63%78
      Content-Transfer-Encoding: base64

      0JDQkdCS0JPQlNCD0JXQltCX0IXQmNCI0JrQm9CJ0JzQndCK0J7Qn9Cg0KHQotCM0KPQpNCl0KfQj9CX0Kg=

      --------------9xfXriG1c1v5iJlMiIMCaIWP--

      """
    Then it succeeds
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                  | to                    | subject                       |
      | [user:user1]@[domain] | [user:user2]@[domain] | Test with cyrillic attachment |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "Test with cyrillic attachment",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "[user:user2]@[domain]",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [],
          "MIMEType": "text/plain"
        }
      }
      """
    And the body in the "POST" response to "/mail/v4/attachments" is:
      """
      {
        "Attachment":{
          "Name": "АБВГДЃЕЖЗЅИЈКЛЉМНЊОПРСТЌУФХЧЏЗШ.docx",
          "MIMEType": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
          "Disposition": "attachment"
          }
      }
      """