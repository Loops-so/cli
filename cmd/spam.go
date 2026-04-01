package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var spamCmd = &cobra.Command{
	Use:    "spam",
	Short:  "Send spam",
	Long:   "Send spam",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(`==========================================
| ,dP""8a "888888b,  d8b    "888b  ,888" |
| 88b   "  888  d88 dPY8b    88Y8b,8888  |
| ` + "`" + `"Y8888a 888ad8P'dPaaY8b   88 Y88P888  |
| a,   Y88 888    dP    Y8b  88  YP 888  |
| ` + "`" + `"8ad8P'a888a  a88a;*a888aa88a   a888a |
|                ;*;;;;*;;;*;;;*,,       |
|        _,---''6ooc,*;;;*;;;*;;*d;,     |
|     .-'      666o6o6o6oc,*;;*;dHH;     |
|   .' nhhn,.  6666o66oo6o6o6cMMMMMM` + "`" + `.   |
|  / nhhhhhhhn,66666666666o6oo6MMMMMW,\  |
| .,nhhhhhhhhnhY666666666666666MMMMWHP", |
| |nhhhhhhhnhMFjj,boY6666666666MMMWWHP | |
| ` + "``" + `hhhhhnhWFjjjjjbbbbbboY6666MMMWWHPf ' |
|  \ ` + "`" + `mYHMFjjjjjjjjbbbbbbbbbboYHHPP"` + "`" + ` /  |
|   ` + "`" + `. ""ijjjjjjjjjmbbbbbbbbbbbbbo, ,'   |
|     ` + "`" + `-._` + "`" + `iijjjjmMF` + "`" + `"bbbbb<=========.   |
|         ` + "`" + `---..._______...|<[Hormel |   |
|                          ` + "`" + `========='   |
==========================================`)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(spamCmd)
}
