@regression
Feature: The user reports a problem
  Background:
    Given bridge starts
    And it succeeds

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
          "url": "https://proton.me/support/invalid-password-error-setting-email-client",
          "title": "Invalid password error while setting up email client"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "New Outlook for Windows" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/proton-mail-bridge-new-outlook-for-windows-set-up-guide",
          "title": "Proton Mail Bridge New Outlook for Windows set up guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook#may-17",
          "title": "Important notice regarding the New Outlook for Mac and issues you might face"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook",
          "title": "Proton Mail Bridge new Outlook for macOS setup guide"
        }
      ]
      """
  
  Scenario: The user wants to report a problem
    Then the description "Setup Outlook Windows" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        },
        {
           "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "I am seeing recovered messages folder in the email client" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/what-is-the-recovered-messages-folder-in-bridge",
          "title": "What is the Recovered Messages folder in Bridge (and your email client)?"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "New Outlook for macOS" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook",
          "title": "Proton Mail Bridge new Outlook for macOS setup guide"
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
    Then the description "How to update Bridge" provides the following KB suggestions:
      """
      [
        {
         "url":"https://proton.me/support/update-required",
         "title":"Update required"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "automatic update" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/update-required",
          "title": "Update required"
        } 
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "restart Bridge" provides the following KB suggestions:
      """
      [
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
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-thunderbird",
          "title": "Proton Mail Bridge Thunderbird setup guide for Windows, macOS, and Linux"
        },
        {
          "url": "https://proton.me/support/how-to-resolve-connection-issues-in-bridge",
          "title": "How to resolve connection issues in Bridge"
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
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        },
        {
          "url": "https://proton.me/support/how-to-resolve-connection-issues-in-bridge",
          "title": "How to resolve connection issues in Bridge"
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
          "url": "https://proton.me/support/protonmail-bridge-clients-apple-mail",
          "title": "Proton Mail Bridge Apple Mail setup guide"
        },
        {
          "url": "https://proton.me/support/how-to-resolve-connection-issues-in-bridge",
          "title": "How to resolve connection issues in Bridge"
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
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Update required" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/update-required",
          "title": "Update required"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "port already occupied error" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "How to fix “IMAP or SMTP port error” on Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/bridge-linux-login-error",
          "title": "How to fix Proton Bridge login errors"
        },
        {
          "url": "https://proton.me/support/invalid-password-error-setting-email-client",
          "title": "Invalid password error while setting up email client"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "1143 port" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "How to fix “IMAP or SMTP port error” on Proton Mail Bridge"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "1025 port" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "How to fix “IMAP or SMTP port error” on Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "IMAP server" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "How to fix “IMAP or SMTP port error” on Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "SMTP server" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "How to fix “IMAP or SMTP port error” on Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Canary supported by Proton Mail Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "How to fix “IMAP or SMTP port error” on Proton Mail Bridge"
        }
      ]
      """ 
  
  Scenario: The user wants to report a problem
    Then the description "install a certificate for Apple Mail" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/apple-mail-certificate",
          "title": "Why you need to install a certificate for Apple Mail with Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/macos-certificate-warning",
          "title": "Warning when installing Proton Mail Bridge on macOS"
        },
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        }
      ]
      """
  
  Scenario: The user wants to report a problem
    Then the description "configure Bridge" provides the following KB suggestions:
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
          "url": "https://proton.me/support/protonmail-bridge-clients-apple-mail",
          "title": "Proton Mail Bridge Apple Mail setup guide"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Install Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/macos-certificate-warning",
          "title": "Warning when installing Proton Mail Bridge on macOS"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Bridge on Linux" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-linux-tray-icon",
          "title": "How to fix a missing system tray icon in Linux"
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
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-apple-mail",
          "title": "Proton Mail Bridge Apple Mail setup guide"
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
          "title": "How to fix “IMAP or SMTP port error” on Proton Mail Bridge"
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
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Proton Mail Bridge Microsoft Outlook configuration" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook#may-17",
          "title": "Important notice regarding the New Outlook for Mac and issues you might face"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook",
          "title": "Proton Mail Bridge new Outlook for macOS setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        }
      ]
      """ 

   Scenario: The user wants to report a problem
    Then the description "Proton Mail Bridge Microsoft Outlook 2016" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/proton-mail-bridge-new-outlook-for-windows-set-up-guide",
          "title": "Proton Mail Bridge New Outlook for Windows set up guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook#may-17",
          "title": "Important notice regarding the New Outlook for Mac and issues you might face"
        },
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        }
      ]
      """ 
      
   Scenario: The user wants to report a problem
    Then the description "Proton Mail Bridge Apple Mail setup guide" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-apple-mail",
          "title": "Proton Mail Bridge Apple Mail setup guide"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Apple Mail setup configuration" provides the following KB suggestions:
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
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-thunderbird",
          "title": "Proton Mail Bridge Thunderbird setup guide for Windows, macOS, and Linux"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Outlook for macOS configuration" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook",
          "title": "Proton Mail Bridge new Outlook for macOS setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook#may-17",
          "title": "Important notice regarding the New Outlook for Mac and issues you might face"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook 2019 for macOS setup guide"
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
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook#may-17",
          "title": "Important notice regarding the New Outlook for Mac and issues you might face"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Outlook 2016 for macOS" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook 2019 for macOS setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook",
          "title": "Proton Mail Bridge new Outlook for macOS setup guide"
        },
        {
          "url": "https://proton.me/support/proton-mail-bridge-new-outlook-for-windows-set-up-guide",
          "title": "Proton Mail Bridge New Outlook for Windows set up guide"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Outlook 2019 for macOS" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook 2019 for macOS setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook",
          "title": "Proton Mail Bridge new Outlook for macOS setup guide"
        },
        {
          "url": "https://proton.me/support/proton-mail-bridge-new-outlook-for-windows-set-up-guide",
          "title": "Proton Mail Bridge New Outlook for Windows set up guide"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "Proton Mail Bridge Microsoft Outlook for Windows 2013 setup guide" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "macOS and Outlook 11" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook 2019 for macOS setup guide"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-new-outlook",
          "title": "Proton Mail Bridge new Outlook for macOS setup guide"
        },
        {
          "url": "https://proton.me/support/proton-mail-bridge-new-outlook-for-windows-set-up-guide",
          "title": "Proton Mail Bridge New Outlook for Windows set up guide"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "use PKGBUILD file for Linux" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-linux-tray-icon",
          "title": "How to fix a missing system tray icon in Linux"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "use DEB file for Linux" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-linux-tray-icon",
          "title": "How to fix a missing system tray icon in Linux"
        }
      ]
      """ 

  Scenario: The user wants to report a problem
    Then the description "verify Bridge package for Linux" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-linux-tray-icon",
          "title": "How to fix a missing system tray icon in Linux"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "Install Bridge for Linux using an RPM file" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/macos-certificate-warning",
          "title": "Warning when installing Proton Mail Bridge on macOS"
        },
        {
          "url": "https://proton.me/support/bridge-linux-tray-icon",
          "title": "How to fix a missing system tray icon in Linux"
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
          "url": "https://proton.me/support/invalid-password-error-setting-email-client",
          "title": "Invalid password error while setting up email client"
        },
        {
          "url": "https://proton.me/support/automatically-start-bridge",
          "title": "Automatically start Bridge"
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
    Then the description "Bridge Manual update" provides the following KB suggestions:
      """
      [
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
          "url": "https://proton.me/support/macos-certificate-warning",
          "title": "Warning when installing Proton Mail Bridge on macOS"
        },
        {
          "url": "https://proton.me/support/apple-mail-certificate",
          "title": "Why you need to install a certificate for Apple Mail with Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-macos-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook 2019 for macOS setup guide"
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
          "url": "https://proton.me/support/apple-mail-certificate",
          "title": "Why you need to install a certificate for Apple Mail with Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/macos-certificate-warning",
          "title": "Warning when installing Proton Mail Bridge on macOS"
        },
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "Bridge is not able to contact the server" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/how-to-resolve-connection-issues-in-bridge",
          "title": "How to resolve connection issues in Bridge"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "lost connection" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/how-to-resolve-connection-issues-in-bridge",
          "title": "How to resolve connection issues in Bridge"
        },
        {
          "url": "https://proton.me/support/bridge-ssl-connection-issue",
          "title": "Proton Mail Bridge connection issues with Thunderbird, Outlook, and Apple Mail"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "internal error" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-internal-error",
          "title": "How to fix an “internal error” warning on Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "How to fix “IMAP or SMTP port error” on Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/bridge-linux-login-error",
          "title": "How to fix Proton Bridge login errors"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "local cache" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-cant-move-cache",
          "title": "How to fix a “can’t move cache” error on Proton Mail Bridge"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "new emails are not arriving" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/how-to-troubleshoot-messages-received-with-a-delay-in-your-email-client",
          "title": "How to troubleshoot messages received with a delay in your email client"
        },
        {
          "url": "https://proton.me/support/not-receiving-messages-email-client",
          "title": "How to troubleshoot not receiving new messages in your email client"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "I cannot find emails in my email client" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/how-to-troubleshoot-messages-received-with-a-delay-in-your-email-client",
          "title": "How to troubleshoot messages received with a delay in your email client"
        },
        {
          "url": "https://proton.me/support/not-receiving-messages-email-client",
          "title": "How to troubleshoot not receiving new messages in your email client"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "emails are not arriving" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/not-receiving-messages-email-client",
          "title": "How to troubleshoot not receiving new messages in your email client"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "invalid return path" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/resolve-invalid-return-path-error",
          "title": "How to resolve the “Invalid Return Path” error when sending messages from your email client"
        },
        {
          "url": "https://proton.me/support/invalid-password-error-setting-email-client",
          "title": "Invalid password error while setting up email client"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "I get This computer only folder in Outlook" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/this-computer-only-folder-outlook",
          "title": "How to troubleshoot “This computer only” folders in Outlook for Windows"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "I cannot access keychain on Proton Mail Bridge" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-cannot-access-keychain",
          "title": "How to fix “cannot access keychain” on Proton Mail Bridge"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "I get an address list has changed warning" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-address-list-has-changed",
          "title": "How to fix “the address list has changed” warning on Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/difference-combined-addresses-mode-split-addresses-mode",
          "title": "Difference between combined addresses mode and split addresses mode"
        },
        {
          "url": "https://proton.me/support/macos-certificate-warning",
          "title": "Warning when installing Proton Mail Bridge on macOS"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "failed to parse message" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/failed-to-parse-message-error",
          "title": "Troubleshooting “failed to parse message” errors"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "I get an IMAP login failed warning" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/bridge-imap-login-failed",
          "title": "How to fix an “IMAP Login failed” warning on Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/automatically-start-bridge",
          "title": "Automatically start Bridge"
        },
        {
          "url": "https://proton.me/support/port-already-occupied-error",
          "title": "How to fix “IMAP or SMTP port error” on Proton Mail Bridge"
        }
      ]
      """

  Scenario: The user wants to report a problem
    Then the description "Checking mail server capabilities notice" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/resolve-checking-server-notice",
          "title": "How to resolve issues with “Checking mail server capabilities” notice in Thunderbird"
        }
      ]
      """
  
  Scenario: The user wants to report a problem
    Then the description "xoxoxo" provides the following KB suggestions:
      """
      []
      """