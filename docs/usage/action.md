---
title: ❓ FAQ
nav_order: 4
layout: default
parent: 🧭 Usage
---

<!-- markdownlint-disable MD022 MD025 -->
# GitHub Action
{: .no_toc }

For GitHub Actions workflows, use [xget-action](https://github.com/camalot/xget-action)
to install `xget` (with binary caching) and run it in a single step:

``` yaml
- name: Install a tool with xget
  uses: camalot/xget-action@v1
  with:
    package: junegunn/fzf
```

Install xget only (without running it) by omitting the `package` input:

``` yaml
- name: Setup xget
  uses: camalot/xget-action@v1

- name: Install multiple tools
  shell: bash
  run: |
    xget some-org/some-app --asset '~\.tar\.gz$'
    xget some-org2/another-app --asset '~\.tar\.gz$'
```

See the [xget-action README](https://github.com/camalot/xget-action#readme)
for the full list of inputs/outputs and more examples.
