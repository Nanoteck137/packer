package cli

import "github.com/spf13/cobra"

var packCmd = &cobra.Command{
	Use: "pack",
}

type Info struct {
	Name   string `json:"name"`
	Series string `json:"series"`

	IsManga        bool `json:"isManga"`
	PreferVertical bool `json:"preferVertical"`

	Cover string   `json:"cover"`
	Pages []string `json:"pages"`
}

var packOldManga = &cobra.Command{
	Use: "old-manga",
	Run: func(cmd *cobra.Command, args []string) {
		// NOTE(patrik):
		//   <NAME>.sw
		//     info.json
		//     0.jpg
		//     1.png
		//     cover.png
	},
}

func init() {
	packCmd.AddCommand(packOldManga)
	rootCmd.AddCommand(packCmd)
}
