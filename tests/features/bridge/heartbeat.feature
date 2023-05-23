Feature: Send Telemetry Heartbeat
  Background:
    Given there exists an account with username "[user:user1]" and password "password"
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
        "Event": "bridge_heartbeat",
        "Values": {
          "nb_account": 1
        },
        "Dimensions": {
          "auto_update": "on",
          "auto_start": "on",
          "beta": "off",
          "doh": "off",
          "split_mode": "off",
          "show_all_mail": "on",
          "imap_connection_mode": "starttls",
          "smtp_connection_mode": "starttls",
          "imap_port": "default",
          "smtp_port": "default",
          "cache_location": "default",
          "keychain_pref": "default",
          "prev_version": "0.0.0",
          "rollout": "42"
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
        "Event": "bridge_heartbeat",
        "Values": {
          "nb_account": 1
        },
        "Dimensions": {
          "auto_update": "off",
          "auto_start": "off",
          "beta": "off",
          "doh": "on",
          "split_mode": "off",
          "show_all_mail": "off",
          "imap_connection_mode": "ssl",
          "smtp_connection_mode": "ssl",
          "imap_port": "custom",
          "smtp_port": "custom",
          "cache_location": "custom",
          "keychain_pref": "custom",
          "prev_version": "0.0.0",
          "rollout": "42"
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
        "Event": "bridge_heartbeat",
        "Values": {
          "nb_account": 1
        },
        "Dimensions": {
          "auto_update": "on",
          "auto_start": "on",
          "beta": "off",
          "doh": "off",
          "split_mode": "on",
          "show_all_mail": "on",
          "imap_connection_mode": "starttls",
          "smtp_connection_mode": "starttls",
          "imap_port": "default",
          "smtp_port": "default",
          "cache_location": "default",
          "keychain_pref": "default",
          "prev_version": "0.0.0",
          "rollout": "42"
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
