package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/alecthomas/kong"
	"github.com/goccy/go-yaml"
	"github.com/thedataflows/confedit/internal/config"
	"github.com/thedataflows/confedit/internal/features/dconf"
	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/features/systemd"
	"github.com/thedataflows/confedit/internal/types"
	log "github.com/thedataflows/go-lib-log"
)

// ListCmd lists configured targets
type ListCmd struct {
	Long   bool   `short:"l" help:"Show detailed information about targets"`
	Format string `short:"f" default:"table" enum:"table,json,yaml" help:"Output format (table, json, yaml)"`
}

func (c *ListCmd) Run(ctx *kong.Context, cli *CLI) error {
	log.Infof(PKG_CMD, "Listing configured targets")
	log.Debugf(PKG_CMD, "List command options: %+v; context: %+v", cli, ctx.Args)

	// Initialize loader to get configuration
	loader := config.NewCueConfigLoader(cli.Config, cli.Schema)
	systemConfig, err := loader.LoadConfiguration()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	if len(systemConfig.Targets) == 0 {
		fmt.Println("No targets configured.")
		return nil
	}

	switch c.Format {
	case "json":
		return c.outputJSON(systemConfig.Targets)
	case "yaml":
		return c.outputYAML(systemConfig.Targets)
	case "table":
		return c.outputTable(systemConfig.Targets)
	default:
		return fmt.Errorf("unsupported format: %s", c.Format)
	}
}

func (c *ListCmd) outputJSON(targets []types.AnyTarget) error {
	if c.Long {
		// Output the actual target structs directly
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(targets)
	} else {
		// For simple output, just show name and type
		simple := make([]map[string]string, len(targets))
		for i, target := range targets {
			simple[i] = map[string]string{
				"name": target.GetName(),
				"type": target.GetType(),
			}
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(simple)
	}
}

func (c *ListCmd) outputYAML(targets []types.AnyTarget) error {
	if c.Long {
		// Output the actual target structs directly
		data, err := yaml.Marshal(targets)
		if err != nil {
			return fmt.Errorf("marshal YAML: %w", err)
		}
		fmt.Print(string(data))
		return nil
	} else {
		// For simple output, just show name and type
		simple := make([]map[string]string, len(targets))
		for i, target := range targets {
			simple[i] = map[string]string{
				"name": target.GetName(),
				"type": target.GetType(),
			}
		}
		data, err := yaml.Marshal(simple)
		if err != nil {
			return fmt.Errorf("marshal YAML: %w", err)
		}
		fmt.Print(string(data))
		return nil
	}
}

func (c *ListCmd) outputTable(targets []types.AnyTarget) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	if c.Long {
		_, _ = fmt.Fprintln(w, "NAME\tTYPE\tDETAILS")
		for _, target := range targets {
			details := c.getTargetDetails(target)
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
				target.GetName(),
				target.GetType(),
				details,
			)
		}
	} else {
		_, _ = fmt.Fprintln(w, "NAME\tTYPE")
		for _, target := range targets {
			_, _ = fmt.Fprintf(w, "%s\t%s\n",
				target.GetName(),
				target.GetType(),
			)
		}
	}

	return nil
}

func (c *ListCmd) getTargetDetails(target types.AnyTarget) string {
	switch target.GetType() {
	case types.TYPE_FILE:
		if fileTarget, ok := target.(*file.Target); ok {
			return fmt.Sprintf("path=%s format=%s", fileTarget.Config.Path, fileTarget.Config.Format)
		}
	case types.TYPE_DCONF:
		if dconfTarget, ok := target.(*dconf.Target); ok {
			return fmt.Sprintf("schema=%s", dconfTarget.Config.Schema)
		}
	case types.TYPE_SYSTEMD:
		if systemdTarget, ok := target.(*systemd.Target); ok {
			return fmt.Sprintf("unit=%s section=%s", systemdTarget.Config.Unit, systemdTarget.Config.Section)
		}
	}
	return ""
}
