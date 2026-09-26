
### GitHub Action

For GitHub Actions workflows, use [xget-action](https://github.com/camalot/xget-action)
to install `xget` (with binary caching) and run it in a single step:

``` yaml
- name: Install a tool with xget
  uses: camalot/xget-action@v1
  with:
    package: junegunn/fzf
```

You can also use the action to just install `xget` on the GitHub Actions runner without installing any packages. You can then use `xget` in subsequent steps to install other tools as needed.

``` yaml
- uses: camalot/xget-action@v1
- shell: bash
  run: |
    xget install eza-community/eza --to ~/.local/bin
    xget install junegunn/fzf --to ~/.local/bin
    xget install bschaatsbergen/cidr --to ~/.local/bin
```

The `xget-action` uses the `--non-interactive` flag by default to ensure that installations do not prompt for user input, which is suitable for automated CI/CD environments.

See the [xget-action](https://github.com/camalot/xget-action#readme)
for the full list of inputs/outputs and more examples.