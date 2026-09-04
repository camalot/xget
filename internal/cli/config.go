package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/camalot/xget/internal/config"
	"github.com/spf13/cobra"
)

// errSilent signals a non-zero exit without printing a message, mirroring
// `git config --get` behavior for unset keys.
var errSilent = errors.New("")

// IsSilent reports whether an error should exit non-zero without any output.
func IsSilent(err error) bool {
	return errors.Is(err, errSilent)
}

type configFlags struct {
	config string
}

func newConfigCommand() *cobra.Command {
	f := &configFlags{}

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Get, set, and edit xget configuration values",
		Long: "Get, set, and edit xget configuration values.\n\n" +
			"Values are read from and written to the config file resolved by the normal\n" +
			"search order, unless --config is given. If no config file exists, a new one is\n" +
			"created at $XDG_CONFIG_HOME/xget/.xget.yml (or ~/.config/xget/.xget.yml).",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringVarP(&f.config, "config", "c", "", "path to the config file to use")
	cmd.AddCommand(
		newConfigGetCommand(f),
		newConfigSetCommand(f),
		newConfigClearCommand(f),
		newConfigPopCommand(f),
		newConfigListCommand(f),
		newConfigPathCommand(f),
		newConfigEditCommand(f),
	)
	return cmd
}

func openDocument(f *configFlags) (*config.Document, error) {
	path, err := config.ResolvePath(f.config)
	if err != nil {
		return nil, err
	}
	return config.LoadDocument(path)
}

func newConfigGetCommand(f *configFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <global|owner/repo> <key>",
		Short: "Print a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := openDocument(f)
			if err != nil {
				return err
			}
			values, err := doc.Get(args[0], args[1])
			if errors.Is(err, config.ErrKeyNotSet) {
				return errSilent
			}
			if err != nil {
				return err
			}
			for _, value := range values {
				cmd.Println(value)
			}
			return nil
		},
	}
}

func newConfigSetCommand(f *configFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "set <global|owner/repo> <key>=<value>",
		Short: "Set a configuration value",
		Long: "Set a configuration value.\n\n" +
			"Scalar keys are replaced. List keys (for example ignore and asset_filters)\n" +
			"append the value; repeat the command to add more entries. Use `xget config pop`\n" +
			"to remove a single entry or `xget config clear` to remove the key.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value, err := config.ParseAssignment(args[1])
			if err != nil {
				return err
			}
			doc, err := openDocument(f)
			if err != nil {
				return err
			}
			if err := doc.Set(args[0], key, value); err != nil {
				return err
			}
			if err := doc.Save(); err != nil {
				return err
			}
			cmd.Printf("%s: %s.%s\n", doc.Path, args[0], key)
			return nil
		},
	}
}

func newConfigClearCommand(f *configFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "clear <global|owner/repo> <key>",
		Short: "Remove a configuration key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := openDocument(f)
			if err != nil {
				return err
			}
			if err := doc.Clear(args[0], args[1]); err != nil {
				if errors.Is(err, config.ErrKeyNotSet) {
					return errSilent
				}
				return err
			}
			if err := doc.Save(); err != nil {
				return err
			}
			cmd.Printf("%s: cleared %s.%s\n", doc.Path, args[0], args[1])
			return nil
		},
	}
}

func newConfigPopCommand(f *configFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "pop <global|owner/repo> <key>=<value>",
		Short: "Remove a single value from a configuration key",
		Long: "Remove a single value from a list-valued configuration key.\n\n" +
			"For scalar keys, the key is removed when its current value matches.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value, err := config.ParseAssignment(args[1])
			if err != nil {
				return err
			}
			doc, err := openDocument(f)
			if err != nil {
				return err
			}
			if err := doc.Pop(args[0], key, value); err != nil {
				if errors.Is(err, config.ErrKeyNotSet) {
					return errSilent
				}
				return err
			}
			if err := doc.Save(); err != nil {
				return err
			}
			cmd.Printf("%s: removed %s from %s.%s\n", doc.Path, value, args[0], key)
			return nil
		},
	}
}

func newConfigListCommand(f *configFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Print every configured value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			doc, err := openDocument(f)
			if err != nil {
				return err
			}
			for _, entry := range doc.Entries() {
				if entry.Key == "" {
					continue
				}
				cmd.Printf("%s.%s=%s\n", entry.Section, entry.Key, entry.Value)
			}
			return nil
		},
	}
}

func newConfigPathCommand(f *configFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the config file path that would be used",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := config.ResolvePath(f.config)
			if err != nil {
				return err
			}
			cmd.Println(path)
			return nil
		},
	}
}

func newConfigEditCommand(f *configFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open the config file in an editor",
		Long: "Open the config file in an editor.\n\n" +
			"The editor is taken from XGET_EDITOR, VISUAL, or EDITOR, falling back to nano\n" +
			"(and notepad on Windows when nano is unavailable).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			doc, err := openDocument(f)
			if err != nil {
				return err
			}
			if !doc.Existed {
				if err := doc.Save(); err != nil {
					return err
				}
			}
			editor, args, err := resolveEditor()
			if err != nil {
				return err
			}
			if err := runEditor(editor, append(args, doc.Path)); err != nil {
				return fmt.Errorf("editor %q failed: %w", editor, err)
			}
			// Re-read so syntax errors surface immediately.
			if _, err := config.Load(doc.Path); err != nil && !os.IsNotExist(err) {
				return err
			}
			return nil
		},
	}
}
