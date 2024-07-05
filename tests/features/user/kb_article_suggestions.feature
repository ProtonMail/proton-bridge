@regression
Feature: The user reports a problem
  Background:
    Given bridge starts
    And it succeeds
  
  Scenario: The user wants to report a problem
    Then the description "Setup Outlook Windows" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "Manual update" provides the following KB suggestions: 
      """
      [
        {
          "url": "https://proton.me/support/bridge-automatic-update",
          "title": "Automatic Update and Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-manual-update",
          "title": "How to manually update Proton Mail Bridge"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "How to update" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-automatic-update",
          "title": "Automatic Update and Bridge"
        },
        {
          "url": "https://proton.me/support/update-required",
          "title": "Update required"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-manual-update",
           "title": "How to manually update Proton Mail Bridge"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "automatic update" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-automatic-update",
          "title": "Automatic Update and Bridge"
        },
        {
          "url": "https://proton.me/support/automatically-start-bridge",
          "title": "Automatically start Bridge"
        },
        {
          "url": "https://proton.me/support/update-required",
           "title": "Update required"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "login on Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/automatically-start-bridge",
          "title": "Automatically start Bridge"
        },
        {
          "url": "https://proton.me/support/bridge-linux-login-error",
          "title": "How to fix Proton Bridge login errors"
        },
        {
          "url": "https://proton.me/support/why-you-need-bridge",
           "title": "Why you need Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "start Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/automatically-start-bridge",
          "title": "Automatically start Bridge"
        },
        {
          "url": "https://proton.me/support/why-you-need-bridge",
          "title": "Why you need Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "restart Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/automatically-start-bridge",
          "title": "Automatically start Bridge"
        },
        {
          "url": "https://proton.me/support/bridge-automatic-update",
          "title": "Automatic Update and Bridge"
        },
        {
          "url": "https://proton.me/support/update-required",
          "title": "Update required"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "message encryption" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/messages-encrypted-via-bridge",
          "title": "Are my messages encrypted via Proton Mail Bridge?"
        },
        {
          "url": "https://proton.me/support/sending-pgp-emails-bridge",
          "title": "Sending PGP emails in Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "pgp encryption" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/sending-pgp-emails-bridge",
          "title": "Sending PGP emails in Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/messages-encrypted-via-bridge",
          "title": "Are my messages encrypted via Proton Mail Bridge?"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "gpg encryption" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/sending-pgp-emails-bridge",
          "title": "Sending PGP emails in Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/messages-encrypted-via-bridge",
          "title": "Are my messages encrypted via Proton Mail Bridge?"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "privacy and security in Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/messages-encrypted-via-bridge",
          "title": "Are my messages encrypted via Proton Mail Bridge?"
        },
        {
          "url": "https://proton.me/support/why-you-need-bridge",
          "title": "Why you need Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "labels in Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/labels-in-bridge",
          "title": "Labels in Bridge"
        },
        {
          "url": "https://proton.me/support/why-you-need-bridge",
          "title": "Why you need Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "folders in Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/labels-in-bridge",
          "title": "Labels in Bridge"
        },
        {
          "url": "https://proton.me/support/why-you-need-bridge",
          "title": "Why you need Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "directories in Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/labels-in-bridge",
          "title": "Labels in Bridge"
        },
        {
          "url": "https://proton.me/support/why-you-need-bridge",
          "title": "Why you need Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "connection issue with Thunderbird" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        },
        {
          "url": "https://proton.me/support/thunderbird-connection-server-timed-error",
          "title": "Thunderbird: 'Connection to server timed out' error"
        },
        {
          "url": "https://proton.me/support/clients-supported-bridge",
          "title": "Email clients supported by Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "connection issue with Outlook" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        },
        {
          "url": "https://proton.me/support/thunderbird-connection-server-timed-error",
          "title": "Thunderbird: 'Connection to server timed out' error"
        },
        {
          "url": "https://proton.me/support/clients-supported-bridge",
          "title": "Email clients supported by Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "connection issue with Apple Mail" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        },
        {
          "url": "https://proton.me/support/thunderbird-connection-server-timed-error",
          "title": "Thunderbird: 'Connection to server timed out' error"
        },
        {
          "url": "https://proton.me/support/clients-supported-bridge",
          "title": "Email clients supported by Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "combined mode" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/difference-combined-addresses-mode-split-addresses-mode",
          "title": "Difference between combined addresses mode and split addresses mode"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "split mode" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/difference-combined-addresses-mode-split-addresses-mode",
          "title": "Difference between combined addresses mode and split addresses mode"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Connection to server timed out" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/thunderbird-connection-server-timed-error",
          "title": "Thunderbird: 'Connection to server timed out' error"
        },
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        },
        {
          "url": "https://proton.me/support/bridge-linux-login-error",
          "title": "How to fix Proton Bridge login errors"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "update required" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/update-required",
          "title": "Update required"
        },
        {
          "url": "https://proton.me/support/bridge-automatic-update",
          "title": "Automatic Update and Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-manual-update",
          "title": "How to manually update Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "port already occupied error" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "Port already occupied error"
        },
        {
          "url": "https://proton.me/support/invalid-password-error-setting-email-client",
          "title": "Invalid password error while setting up email client"
        },
        {
          "url": "https://proton.me/support/bridge-linux-login-error",
          "title": "How to fix Proton Bridge login errors"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "1143 port" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "Port already occupied error"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "1025 port" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "Port already occupied error"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "IMAP server" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "Port already occupied error"
        },
        {
          "url": "https://proton.me/support/imap-smtp-and-pop3-setup",
          "title": "IMAP, SMTP, and POP3 setup"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "SMTP server" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "Port already occupied error"
        },
        {
          "url": "https://proton.me/support/imap-smtp-and-pop3-setup",
          "title": "IMAP, SMTP, and POP3 setup"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Canary supported by Proton Mail Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "Port already occupied error"
        },
        {
          "url": "https://proton.me/support/clients-supported-bridge",
          "title": "Email clients supported by Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/why-you-need-bridge",
          "title": "Why you need Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Eudora" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/clients-supported-bridge",
          "title": "Email clients supported by Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "configure Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/imap-smtp-and-pop3-setup",
          "title": "IMAP, SMTP, and POP3 setup"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/why-you-need-bridge",
          "title": "Why you need Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Install Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/install-bridge-linux-rpm-file",
          "title": "Installing Proton Mail Bridge for Linux using an RPM file"
        },
        {
          "url": "https://proton.me/support/why-you-need-bridge",
          "title": "Why you need Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Bridge on Linux" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/bridge-for-linux",
          "title": "Proton Mail Bridge for Linux"
        },
        {
          "url": "https://proton.me/support/operating-systems-supported-bridge",
          "title": "System requirements for Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "requirements for Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/operating-systems-supported-bridge",
          "title": "System requirements for Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/why-you-need-bridge",
          "title": "Why you need Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Configure email client" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/why-you-need-bridge",
          "title": "Why you need Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Bridge application" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/why-you-need-bridge",
          "title": "Why you need Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "invalid password error" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/invalid-password-error-setting-email-client",
          "title": "Invalid password error while setting up email client"
        },
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "Port already occupied error"
        },
        {
          "url": "https://proton.me/support/bridge-linux-login-error",
          "title": "How to fix Proton Bridge login errors"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Setup guide for Outlook" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2016",
          "title": "Proton Mail Bridge Microsoft Outlook 2016 for Windows setup guide"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Proton Mail Bridge Microsoft Outlook configuration" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2016",
          "title": "Proton Mail Bridge Microsoft Outlook 2016 for Windows setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook",
          "title": "Proton Mail Bridge new Outlook for macOS setup guide"
        }
      ]
      """ 

   Scenario: The user wants to report a problem
    Then the description "Proton Mail Bridge Microsoft Outlook 2016" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        },
        {
          "url": "https://proton.me/support/clients-supported-bridge",
          "title": "Email clients supported by Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        }
      ]
      """ 
      
   Scenario: The user wants to report a problem
    Then the description "Proton Mail Bridge Apple Mail setup guide" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-apple-mail",
          "title": "Proton Mail Bridge Apple Mail setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2016",
          "title": "Proton Mail Bridge Microsoft Outlook 2016 for Windows setup guide"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Apple Mail setup configuration" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-apple-mail",
          "title": "Proton Mail Bridge Apple Mail setup guide"
        },
        {
        "url": "https://proton.me/support/protonmail-bridge-clients-windows-thunderbird",
        "title": "Proton Mail Bridge Thunderbird setup guide for Windows, macOS, and Linux"
        },
        {
          "url": "https://proton.me/support/imap-smtp-and-pop3-setup",
          "title": "IMAP, SMTP, and POP3 setup"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Outlook for macOS configuration" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        },
        {
          "url": "https://proton.me/support/proton-mail-bridge-new-outlook-for-windows-set-up-guide",
          "title": "Proton Mail Bridge New Outlook for Windows set up guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook#may-17",
          "title": "Important notice regarding the New Outlook for Mac and issues you might face"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "Thunderbird configuration" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-thunderbird",
          "title": "Proton Mail Bridge Thunderbird setup guide for Windows, macOS, and Linux"
        },
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        },
        {
          "url": "https://proton.me/support/thunderbird-connection-server-timed-error",
          "title": "Thunderbird: 'Connection to server timed out' error"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Outlook 2016 for macOS" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-outlook-2016",
          "title": "Proton Mail Bridge Microsoft Outlook 2016 for macOS setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Outlook 2019 for macOS" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook 2019 for macOS setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Proton Mail Bridge Microsoft Outlook for Windows 2013 setup guide" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "macOS and Outlook 11" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-outlook-2016",
          "title": "Proton Mail Bridge Microsoft Outlook 2016 for macOS setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook 2019 for macOS setup guide"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "use PKGBUILD file for Linux" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/install-bridge-linux-pkgbuild-file",
          "title": "Installing Proton Mail Bridge for Linux using a PKGBUILD file"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/bridge-for-linux",
          "title": "Proton Mail Bridge for Linux"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "use DEB file for Linux" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-for-linux",
          "title": "Proton Mail Bridge for Linux"
        },
        {
          "url": "https://proton.me/support/installing-bridge-linux-deb-file",
          "title": "Installing Proton Mail Bridge for Linux using a DEB file"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "verify Bridge package for Linux" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/verifying-bridge-package",
          "title": "Verifying the Proton Mail Bridge package for Linux"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/bridge-for-linux",
          "title": "Proton Mail Bridge for Linux"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Bridge CLI guide" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-cli-guide",
          "title": "Bridge CLI (command line interface) guide"
        },
        {
          "url": "https://proton.me/support/why-you-need-bridge",
          "title": "Why you need Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Install Bridge for Linux using an RPM file" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/install-bridge-linux-rpm-file",
          "title": "Installing Proton Mail Bridge for Linux using an RPM file"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/bridge-for-linux",
          "title": "Proton Mail Bridge for Linux"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "login errors on Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-linux-login-error",
          "title": "How to fix Proton Bridge login errors"
        },
        {
          "url": "https://proton.me/support/automatically-start-bridge",
          "title": "Automatically start Bridge"
        },
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "Port already occupied error"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Missing system tray icon in Linux" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-linux-tray-icon",
          "title": "How to fix a missing system tray icon in Linux"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/bridge-for-linux",
          "title": "Proton Mail Bridge for Linux"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "how to receive notification" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-linux-tray-icon",
          "title": "How to fix a missing system tray icon in Linux"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Why you need Proton Mail Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/why-you-need-bridge",
          "title": "Why you need Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Bridge Manual update" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-automatic-update",
          "title": "Automatic Update and Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-manual-update",
          "title": "How to manually update Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/update-required",
          "title": "Update required"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Warning when installing Bridge on macOS" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/macos-certificate-warning",
          "title": "Warning when installing Proton Mail Bridge on macOS"
        },
        {
          "url": "https://proton.me/support/operating-systems-supported-bridge",
          "title": "System requirements for Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Certificate warning" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/macos-certificate-warning",
          "title": "Warning when installing Proton Mail Bridge on macOS"
        },
        {
          "url": "https://proton.me/support/apple-mail-certificate",
          "title": "Why you need to install a certificate for Apple Mail with Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Installing a certificate for Apple Mail on macOS" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/macos-certificate-warning",
          "title": "Warning when installing Proton Mail Bridge on macOS"
        },
        {
          "url": "https://proton.me/support/apple-mail-certificate",
          "title": "Why you need to install a certificate for Apple Mail with Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "xoxoxo" provides the following KB suggestions:
      """
      []
      """ 

