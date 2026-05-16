package main

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	projectcleanerapp "terminal.app/project_cleaner_app"
	serverapp "terminal.app/server_app"
)

func main() {
	for {
		mode := ""
		prompt := &survey.Select{
			Message: "What App Your Run?",
			Options: []string{"Project Cache Cleaner", "Server App", "Exit"},
			Default: "Project Cache Cleaner",
		}

		// Chooser ကို ပြသခြင်း
		err := survey.AskOne(prompt, &mode)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		// ရွေးချယ်မှုအပေါ် မူတည်ပြီး အလုပ်လုပ်ခြင်း
		switch mode {
		case "Project Cache Cleaner":
			projectcleanerapp.RunApp()
		case "Server App":
			serverapp.RunApp()
		case "Exit":
			println("Bye!")
			return
			// os.Exit(0)
		}
	}
}

func clearScreen() {
	// Linux/Unix systems တွေအတွက် standard code ပါ
	fmt.Print("\033[H\033[2J")
}
