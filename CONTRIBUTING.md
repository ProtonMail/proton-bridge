# Contributing guidelines

The following document describes how to contribute to the project. In this context, contribution does not only mean code contribution but also reporting issues, requesting new features, or just asking for help.

## Reporting issues

In case you experience issues while using the application, our request is to contact Proton customer support directly.

The benefits of using Proton customer support are

- Available 24/7/365.
- Provides priority support based on subscription type.
- Will escalate the issue to the developers every time it becomes too technical or they do not know the answer to a question.
- Easier to detect systematic issues by connecting similar reports.
- Possible to quickly derive frequency of an issue.
- Can assist you to transfer sensitive information safely to us.

To speed up the communication with customer support, consider the following:

- Whenever is possible, use the in-app bug report feature. It provides an application specific guide compared to using the generic report form on web.
- Whenever is possible, proactively attach logs to your report. Reporting an issue from the application can help you in that.
- Check whether your system is officially supported by Proton, including the source of the installer. We cannot provide help when the application is packaged by a third party or when the application is used on systems that we do not prepare to support.
- If your report is a feature request, see the Feature request section. In case it is an issue related to application security, see the Security vulnerabilities section.

In the past, we used GitHub issue tracker for more technical issues in parallel to Proton customer support, but we run into limitations with this approach:

- Monitoring GitHub issue tracker took development time as it was managed by the development team.
- It made issue frequency tracking challenging because we did not have a single point of entry for issues.
- Users were confused what technical issue means, and used the GitHub issue tracker for feature requests, or non-technical discussions.
- Users sometimes shared sensitive data through the GitHub issue tracker.

For the above reasons, we do not use GitHub issue tracker anymore but ask you to contact our customer support in case you run into a problem.

### Security vulnerabilities

Proton runs a bug bounty program for security vulnerabilities. They differ from normal bug reports in the following ways:

- These reports go directly to our security team.
- They expect deeper explanation of the issue.
- Depending on the finding, they may be financially rewarded.

More information about the program can be found [here](https://proton.me/security/bug-bounty).

## Feature requests

What someone considers as a bug is sometimes a feature, and sometimes, a missing feature is considered as a bug. Instead of reporting feature requests as bugs, we setup a UserVoice page to allow our users to share their preferences. UserVoice also makes it possible to vote on other feature requests, making the community preference public.

Our product team frequently monitors UserVoice, and the features listed there are taken into account in our planning.

Examples for UserVoice requests:

- Extending the officially supported environments (e.g., operating systems, clients, or computer architectures).
- Requesting new features.
- Integration with non-Proton services.

UserVoice is available [here](https://protonmail.uservoice.com/).

## Asking for help

The best ways to get answer for generic questions or to get help with setting up the system is to interact with our active community on [Reddit](https://reddit.com/r/ProtonMail/) or to contact customer support.

## Code contribution

We are grateful if you can contribute directly with code. In that case there is nothing else to do than to open a pull request.

The following is worthwhile noting

- The project is primarily developed on an internal repository, and the one on GitHub is only a mirror of it. For that reason, the merge request will not be merged on GitHub but added to the project internally. We are keeping the original author in the change set to respect the contribution.
- The application is used on numerous platforms and by many third party clients. To have higher chance your change to be accepted, consider all supported dependencies.
- Give detailed description of the issue, preferably with test steps to reproduce the original issue, and to verify the fix. It is even better if you also extend the automated tests.

### Contribution policy

By making a contribution to this project:

1. You assign any and all copyright related to the contribution to Proton AG;
2. You certify that the contribution was created in whole by you;
3. You understand and agree that this project and the contribution are public and that a record of the contribution (including all personal information you submit with it) is maintained indefinitely and may be redistributed with this project or the open source license(s) involved.
