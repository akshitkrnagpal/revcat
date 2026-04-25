package subscribers

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var vcBalanceCmd = &cobra.Command{
	Use:   "vc-balance <user_id>",
	Short: "Show a customer's virtual currency balances",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		bals, err := client.ListCustomerVCBalances(ctx, args[0])
		if err != nil {
			return err
		}
		return output.JSON(bals)
	},
}

func init() {
	Cmd.AddCommand(vcBalanceCmd)
}

var (
	vcTxFile     string
	vcTxCurrency string
	vcTxAmount   float64
	vcTxReason   string
)

var vcTxCmd = &cobra.Command{
	Use:   "vc-tx <user_id>",
	Short: "Post a virtual currency transaction (credit or debit)",
	Long: `Post a virtual currency transaction. Pass --file <body.json> for the
full v2 shape, or use the shortcut flags --currency / --amount / --reason
for the common case.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var body map[string]any
		if vcTxFile != "" {
			b, err := cliutil.LoadJSON(vcTxFile)
			if err != nil {
				return err
			}
			body = b
		} else if vcTxCurrency != "" && vcTxAmount != 0 {
			body = map[string]any{
				"virtual_currency_id": vcTxCurrency,
				"amount":              vcTxAmount,
			}
			if vcTxReason != "" {
				body["reason"] = vcTxReason
			}
		} else {
			return errors.New("pass --file or --currency + --amount")
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		out, err := client.CreateCustomerVCTransaction(ctx, args[0], body)
		if err != nil {
			return err
		}
		output.Success("transaction posted")
		if output.IsJSON() {
			return output.JSON(out)
		}
		return nil
	},
}

func init() {
	vcTxCmd.Flags().StringVar(&vcTxFile, "file", "", "JSON body")
	vcTxCmd.Flags().StringVar(&vcTxCurrency, "currency", "", "Virtual currency id or lookup_key")
	vcTxCmd.Flags().Float64Var(&vcTxAmount, "amount", 0, "Amount (positive credits, negative debits)")
	vcTxCmd.Flags().StringVar(&vcTxReason, "reason", "", "Audit reason")
	Cmd.AddCommand(vcTxCmd)
}

var (
	vcSetFile string
)

var vcSetBalanceCmd = &cobra.Command{
	Use:   "vc-set-balance <user_id>",
	Short: "Directly set a virtual currency balance",
	Long: `Override a customer's virtual currency balance. Pass --file <body.json>
with the v2 shape (typically virtual_currency_id + balance).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := cliutil.LoadJSON(vcSetFile)
		if err != nil {
			return err
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		out, err := client.UpdateCustomerVCBalance(ctx, args[0], body)
		if err != nil {
			return err
		}
		output.Success("balance updated")
		if output.IsJSON() {
			return output.JSON(out)
		}
		return nil
	},
}

func init() {
	vcSetBalanceCmd.Flags().StringVarP(&vcSetFile, "file", "f", "", "JSON body (required)")
	_ = vcSetBalanceCmd.MarkFlagRequired("file")
	Cmd.AddCommand(vcSetBalanceCmd)
}
