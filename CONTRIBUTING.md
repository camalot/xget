# Contributing to Workspace Tasks

Thank you for your interest in contributing to Workspace Tasks! Whether you're reporting a bug, requesting a feature, improving documentation, or submitting code changes, your contributions help make this extension better for everyone.

This guide will help you get started with contributing to the project.

## 📑 Table of Contents

- [Ways to Contribute](#ways-to-contribute)
- [Creating Good Issues](#creating-good-issues)
  - [Look for an Existing Issue](#look-for-an-existing-issue)
  - [Writing Good Bug Reports](#writing-good-bug-reports)
  - [Writing Good Feature Requests](#writing-good-feature-requests)
- [Contributing Code](#contributing-code)
  - [Development Setup](#development-setup)
  - [Project Structure](#project-structure)
  - [Code Style Guidelines](#code-style-guidelines)
  - [Testing Your Changes](#testing-your-changes)
  - [Submitting a Pull Request](#submitting-a-pull-request)
- [Contributing Documentation](#contributing-documentation)
- [Code of Conduct](#code-of-conduct)

## Ways to Contribute

There are many ways you can contribute to Workspace Tasks:

### 🐛 Report Bugs

Found a bug? Help us fix it by creating a detailed bug report. See [Writing Good Bug Reports](#writing-good-bug-reports).

### ✨ Request Features

Have an idea for a new feature or enhancement? We'd love to hear it! See [Writing Good Feature Requests](#writing-good-feature-requests).

### 📝 Improve Documentation

Help make the documentation clearer, fix typos, or add examples. Documentation improvements are always welcome!

### 💻 Submit Code Changes

Fix bugs, implement features, or improve performance by submitting pull requests. See [Contributing Code](#contributing-code).

### 💬 Help Others

Answer questions in [GitHub Discussions](https://github.com/camalot/vscode-workspace-tasks/discussions) or help troubleshoot issues.

### ⭐ Spread the Word

- Star the repository on GitHub
- Leave a review on the [Visual Studio Code Marketplace](https://marketplace.visualstudio.com/items?itemName=darthminos.workspace-tasks)
- Share the extension with colleagues and friends

## Creating Good Issues

### Look for an Existing Issue

Before creating a new issue, please search [existing issues](https://github.com/camalot/vscode-workspace-tasks/issues) to see if the problem or request has already been reported.

- Scan through [open issues](https://github.com/camalot/vscode-workspace-tasks/issues)
- Check the [most popular feature requests](https://github.com/camalot/vscode-workspace-tasks/issues?q=is%3Aopen+is%3Aissue+label%3Aenhancement+sort%3Areactions-%2B1-desc)

If you find an existing issue that matches yours:

- Add a 👍 reaction to show your support
- Add relevant comments with additional context or information
- Avoid "+1" comments—use reactions instead

### Writing Good Bug Reports

When reporting a bug, please include:

**Required Information:**

- **xget Version** - `xget --version`
- **Command Executed** - The exact command you ran, e.g., `xget zyedidia/micro --tag nightly`
- **Operating System** - Windows, macOS, or Linux (including version)
- **Reproducible Steps** - Clear numbered steps to reproduce the issue
  1. Open workspace with...
  2. Click on...
  3. See error...
- **Expected Behavior** - What you expected to happen
- **Actual Behavior** - What actually happened

**Helpful Additions:**

- **Screenshots or Recordings** - Visual evidence of the issue
- **Error Messages** - Copy and paste any error messages or stack traces
- **Configuration** - Include relevant configuration files

### Writing Good Feature Requests

When requesting a feature, please describe:

- **Problem Statement** - What problem would this feature solve?
- **Proposed Solution** - How should this feature work?
- **Use Case** - How would you use this feature?
- **Alternatives Considered** - Other solutions you've tried
- **Examples** - Screenshots or mockups if applicable
- **Impact** - Who would benefit from this feature?

## Contributing Code

### Development Setup

#### Prerequisites

- **Go** - Version 1.27.0 or later for building xget
- **Git** - For cloning the repository
- **Visual Studio Code** - Latest version recommended

#### Initial Setup

1. **Fork the Repository**

   Click the "Fork" button on the [GitHub repository](https://github.com/camalot/xget)

2. **Clone Your Fork**

   ```bash
   git clone https://github.com/YOUR-USERNAME/xget.git
   cd xget
   ```

3. **Add Upstream Remote**

   ```bash
   git remote add upstream https://github.com/camalot/xget.git
   ```

4. **Open in Visual Studio Code**

   ```bash
   code .
   ```

#### Git Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```text
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, no logic change)
- `refactor:` - Code refactoring
- `perf:` - Performance improvements
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks, dependency updates

### Testing Your Changes

Before submitting a pull request, test your changes thoroughly:

1. **Run Unit Tests**

   ```bash
   go test ./...
   ```

### Submitting a Pull Request

Once your changes are ready:

1. **Update from Upstream**

   ```bash
   git fetch upstream
   git rebase upstream/develop
   ```

2. **Create a Feature Branch**

   ```bash
   git checkout -b feature/your-feature-name
   ```

   Or for bug fixes:

   ```bash
   git checkout -b fix/issue-description
   ```

3. **Commit Your Changes**

   ```bash
   git add .
   git commit -m "feat: add support for new task type"
   ```

4. **Push to Your Fork**

   ```bash
   git push origin feature/your-feature-name
   ```

5. **Open a Pull Request**
   - Go to your fork on GitHub
   - Click "Compare & pull request"
   - Select `camalot/xget` `develop` branch as the base
   - Fill out the PR template with details about your changes

**Pull Request Guidelines:**

- **Clear Title** - Summarize the change in the PR title
- **Description** - Explain what you changed and why
- **Related Issues** - Reference any related issues (e.g., "Fixes #123")
- **Testing** - Describe how you tested your changes
- **Screenshots** - Include screenshots for UI changes
- **Documentation** - Update README.md or other docs if needed
- **Changelog** - Add entry to CHANGELOG.md following existing format

## Contributing Documentation

Documentation improvements are valuable contributions! You can help by:

- **Fixing Typos and Grammar** - Submit PRs for corrections
- **Clarifying Instructions** - Make setup or usage instructions clearer
- **Adding Examples** - Provide real-world usage examples
- **Updating Screenshots** - Replace outdated images
- **Expanding Sections** - Add more detail to existing documentation
- **Creating Guides** - Write tutorials or how-to guides

Documentation files to consider:

- `README.md` - Main user documentation
- `SUPPORT.md` - Support and help resources
- `CHANGELOG.md` - Release notes and version history
- Code comments and Documentation in source files

## Code of Conduct

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for details on our code of conduct, which outlines expected behavior and how to report issues.

### Reporting Issues

If you experience or witness unacceptable behavior, please report it by:

- Opening a private issue on GitHub
- Contacting the project maintainer directly

All reports will be handled with discretion and confidentiality.

## Questions?

If you have questions about contributing:

- 💬 Ask in [GitHub Discussions](https://github.com/camalot/xget/discussions)
- 🐛 Check [existing issues](https://github.com/camalot/xget/issues)
- 📚 Review the [README](README.md) and [SUPPORT](SUPPORT.md) documents
