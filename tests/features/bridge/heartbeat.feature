Feature: Send Telemetry Heartbeat
  Background:
    Given there exists an account with username "[user:user1]" and password "password"
    And there exists an account with username "[user:user2]" and password "password"
    And there exists an account with username "[user:user3]" and password "password"
    Then it succeeds
    When bridge starts
    Then it succeeds


  Scenario: Send at first start - one user default settings
    Then bridge telemetry feature is enabled
    And bridge needs to send heartbeat
    When the user logs in with username "[user:user1]" and password "password"
    And user "[user:user1]" finishes syncing
    Then bridge eventually sends the following heartbeat:
    """
      {
        "MeasurementGroup": "bridge.any.usage",
        "Event": "bridge_heartbeat_new",
        "Values": {
          "NumberConnectedAccounts": 1,
          "rolloutPercentage": 1
        },
        "Dimensions": {
          "isAutoUpdateEnabled": "true",
          "isAutoStartEnabled": "true",
          "isBetaEnabled": "false",
          "isDohEnabled": "false",
          "usesSplitMode": "false",
          "useAllMail": "true",
          "useDefaultImapPort": "true",
          "useDefaultSmtpPort": "true",
          "useDefaultCacheLocation": "true",
          "useDefaultKeychain": "true",
          "isContactedByAppleNotes": "false",
          "imapConnectionMode": "starttls",
          "smtpConnectionMode": "starttls",
          "prevVersion": "0.0.0",
          "bridgePlanGroup": "unknown"
        }
      }
      """
    And bridge do not need to send heartbeat


  Scenario: Send at first start - one user modified settings
    Then bridge telemetry feature is enabled
    And bridge needs to send heartbeat
    When the user has disabled automatic updates
    And the user has disabled automatic start
    And the user has enabled alternative routing
    And the user hides All Mail
    And the user set IMAP mode to SSL
    And the user set SMTP mode to SSL
    And the user changes the IMAP port to 42695
    And the user changes the SMTP port to 56942
    And the user changes the gluon path
    And the user changes the default keychain application
    When the user logs in with username "[user:user1]" and password "password"
    And user "[user:user1]" finishes syncing
    Then bridge eventually sends the following heartbeat:
      """
      {
        "MeasurementGroup": "bridge.any.usage",
        "Event": "bridge_heartbeat_new",
        "Values": {
          "NumberConnectedAccounts": 1,
          "rolloutPercentage": 1
        },
        "Dimensions": {
          "isAutoUpdateEnabled": "false",
          "isAutoStartEnabled": "false",
          "isBetaEnabled": "false",
          "isDohEnabled": "true",
          "usesSplitMode": "false",
          "useAllMail": "false",
          "useDefaultImapPort": "false",
          "useDefaultSmtpPort": "false",
          "useDefaultCacheLocation": "false",
          "useDefaultKeychain": "false",
          "isContactedByAppleNotes": "false",
          "imapConnectionMode": "ssl",
          "smtpConnectionMode": "ssl",
          "prevVersion": "0.0.0",
          "bridgePlanGroup": "unknown"
        }
      }
      """
    And bridge do not need to send heartbeat


  Scenario: Send at first start - one user telemetry disabled
    Then bridge telemetry feature is enabled
    And bridge needs to send heartbeat
    When the user disables telemetry in bridge settings
    And the user logs in with username "[user:user1]" and password "password"
    And user "[user:user1]" finishes syncing
    And bridge needs to send heartbeat
    Then the user sets the address mode of user "[user:user1]" to "split"
    And the user enables telemetry in bridge settings
    Then bridge eventually sends the following heartbeat:
      """
      {
        "MeasurementGroup": "bridge.any.usage",
        "Event": "bridge_heartbeat_new",
        "Values": {
          "NumberConnectedAccounts": 1,
          "rolloutPercentage": 1
        },
        "Dimensions": {
          "isAutoUpdateEnabled": "true",
          "isAutoStartEnabled": "true",
          "isBetaEnabled": "false",
          "isDohEnabled": "false",
          "usesSplitMode": "true",
          "useAllMail": "true",
          "useDefaultImapPort": "true",
          "useDefaultSmtpPort": "true",
          "useDefaultCacheLocation": "true",
          "useDefaultKeychain": "true",
          "isContactedByAppleNotes": "false",
          "imapConnectionMode": "starttls",
          "smtpConnectionMode": "starttls",
          "prevVersion": "0.0.0",
          "bridgePlanGroup": "unknown"
        }
      }
      """
    And bridge do not need to send heartbeat


  Scenario: Multiple-users on Bridge reported correctly
    Then bridge telemetry feature is enabled
    When the user logs in with username "[user:user1]" and password "password"
    Then it succeeds
    When the user logs in with username "[user:user2]" and password "password"
    Then it succeeds
    When the user logs in with username "[user:user3]" and password "password"
    Then it succeeds
    When bridge needs to explicitly send heartbeat
    Then bridge eventually sends the following heartbeat:
      """
      {
        "MeasurementGroup": "bridge.any.usage",
        "Event": "bridge_heartbeat_new",
        "Values": {
          "NumberConnectedAccounts": 3,
          "rolloutPercentage": 1
        },
        "Dimensions": {
          "isAutoUpdateEnabled": "true",
          "isAutoStartEnabled": "true",
          "isBetaEnabled": "false",
          "isDohEnabled": "false",
          "usesSplitMode": "false",
          "useAllMail": "true",
          "useDefaultImapPort": "true",
          "useDefaultSmtpPort": "true",
          "useDefaultCacheLocation": "true",
          "useDefaultKeychain": "true",
          "isContactedByAppleNotes": "false",
          "imapConnectionMode": "starttls",
          "smtpConnectionMode": "starttls",
          "prevVersion": "0.0.0",
          "bridgePlanGroup": "unknown"
        }
      }
      """
    And bridge do not need to send heartbeat


  Scenario: Send heartbeat explicitly - apple notes tried to connect
    Then bridge telemetry feature is enabled
    When the user logs in with username "[user:user1]" and password "password"
    Then it succeeds
    When user "[user:user1]" connects IMAP client "1"
    And  IMAP client "1" announces its ID with name "Mac OS X Notes" and version "14.5"
    When bridge needs to explicitly send heartbeat
    Then bridge eventually sends the following heartbeat:
      """
      {
        "MeasurementGroup": "bridge.any.usage",
        "Event": "bridge_heartbeat_new",
        "Values": {
          "NumberConnectedAccounts": 1,
          "rolloutPercentage": 1
        },
        "Dimensions": {
          "isAutoUpdateEnabled": "true",
          "isAutoStartEnabled": "true",
          "isBetaEnabled": "false",
          "isDohEnabled": "false",
          "usesSplitMode": "false",
          "useAllMail": "true",
          "useDefaultImapPort": "true",
          "useDefaultSmtpPort": "true",
          "useDefaultCacheLocation": "true",
          "useDefaultKeychain": "true",
          "isContactedByAppleNotes": "true",
          "imapConnectionMode": "starttls",
          "smtpConnectionMode": "starttls",
          "prevVersion": "0.0.0",
          "bridgePlanGroup": "unknown"
        }
      }
      """
    And bridge do not need to send heartbeat


  Scenario: GroupMeasurement rejected by API
    Given heartbeat is not whitelisted
    Then bridge telemetry feature is enabled
    And bridge needs to send heartbeat
    When the user logs in with username "[user:user1]" and password "password"
    And user "[user:user1]" finishes syncing
    Then bridge needs to send heartbeat
