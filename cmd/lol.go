package cmd

import (
	"bufio"
	"fmt"
	"log"
	"lol_stats/internal/api"
	"lol_stats/internal/model"
	"lol_stats/internal/persistence"
	"lol_stats/internal/printer"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// const gamename = "FREE PALESTINE"
// const tag = "tox"

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		load, err := cmd.Flags().GetBool("load")
		if err != nil {
			log.Fatal("Error parsing `load` flag")
		}

		if err != nil {
			log.Fatal(err)
		}

		apiKey := LoadApiKey()

		if load {

			path, err := persistence.GetConfigPath(persistence.Account)

			if err != nil {
				panic(err)
			}

			if _, err := os.Stat(path); err != nil {
				configSetup(apiKey)
			}

			config, err := persistence.LoadConfig()

			if err != nil {
				log.Fatal(err)
			}

			performances := persistence.QueryPerformances(config, apiKey)

			persistence.SaveGames(performances)
		}

		performances, err := persistence.LoadGames()

		if err != nil {
			log.Fatal("Error parsing `game` flag")
		}

		game, err := cmd.Flags().GetInt("game")

		if err != nil {
			log.Fatal(err)
		}

		if game > 1 {
			printer.PrintParticipantStats(performances[game])
		} else {
			printer.PrintPerformanceChart(performances)
		}

	},
}

func configSetup(apiKey string) {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("There is no account file, press Y/N to proceed and make one or to terminate")

	ans, err := reader.ReadString('\n')

	if err != nil {
		panic(err)
	}

	if strings.ToUpper(ans) == "N" {
		os.Exit(0)
	}

	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')

	username = strings.TrimSpace(username)

	if err != nil {
		panic(err)
	}

	fmt.Print("Tagline: ")
	tagline, err := reader.ReadString('\n')

	if err != nil {
		panic(err)
	}

	tagline = strings.TrimSpace(tagline)

	account, err := api.QueryAccount(username, tagline, apiKey)

	if err != nil {
		log.Fatal("error querying account")
	}

	config := model.Config{PUUID: account.PUUID, Username: account.GameName}

	persistence.SaveConfig(config)
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().BoolP("load", "l", false, "Show past 20 games")
	statsCmd.Flags().IntP("game", "g", 0, "Show indexed game")
}

func LoadApiKey() string {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Unable to load env file")

	}

	apiKey := os.Getenv("API_KEY")

	return apiKey
}
