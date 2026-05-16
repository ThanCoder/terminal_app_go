package projectcleanerapp

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/AlecAivazis/survey/v2"
)

type ProjectInfo struct {
	Type  string
	Path  string
	Size  int64
	Items []string //ဖျက်မယ့် folder များ
}

func RunApp() {
	cwd, _ := os.UserHomeDir()
	scanPath := filepath.Join(cwd, "projects")
	fmt.Printf("🔍 Scanning projects in: %s\n\n", scanPath)

	projects := scanProjects(scanPath)
	fmt.Printf("projects: %v\n", projects)

	if len(projects) == 0 {
		fmt.Println("✨ No cache folders found. Everything is clean!")
		return
	}
	// Tabwriter သုံးပြီး column တွေကို စနစ်တကျ ပြပါမယ်
	displayAndConfirm(projects)

}

// Helper functions (dirSize, exists)
func dirSize(path string) (int64, error) {
	var allSize int64
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			allSize += info.Size()
		}

		return nil
	})
	return allSize, err

}

func existsPath(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Display logic (Tabwriter ပိုင်း) ကို ဒီမှာ ခွဲထုတ်ထားတယ်
func displayAndConfirm(projects []ProjectInfo) {
	if len(projects) == 0 {
		fmt.Println("✨ No cache folders found.")
		return
	}

	// Terminal Screen ကို အကုန်ရှင်းထုတ်ပြီး cursor ကို ထိပ်ဆုံးပြန်တင်မယ်
	fmt.Print("\033[H\033[2J")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "TYPE\tPATH\tCACHE SIZE")
	fmt.Fprintln(w, "----\t----\t----------")

	var totalSize int64
	for _, p := range projects {
		totalSize += p.Size
		fmt.Fprintf(w, "%s\t%s\t%s\n", p.Type, p.Path, formatSize(p.Size))
	}
	w.Flush()

	fmt.Printf("\nTotal Project Found: %d\n", len(projects))
	fmt.Printf("Total Space Reclaimable: %s\n", formatSize(totalSize))

	fmt.Print("\n⚠️ Delete all? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) == "y" {
		deleteOptions(projects)
	}
}

func formatSize(bytes int64) string {
	const unit = 1024

	// Byte အဆင့်ပဲရှိရင် တိုက်ရိုက်ပြန်မယ်
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	// Units သတ်မှတ်ချက်
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}

	size := float64(bytes)
	unitIndex := -1

	// 1024 နဲ့ စားလို့ရသရွေ့ unit တစ်ဆင့်ချင်း တက်သွားမယ်
	for size >= unit && unitIndex < len(units)-1 {
		size /= unit
		unitIndex++
	}

	// ဥပမာ - 1.25 GB, 500.00 MB
	return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}

func deleteOptions(projects []ProjectInfo) {
	mode := ""
	prompt := &survey.Select{
		Message: "Choose deletion method (Use arrow keys):",
		Options: []string{"Normal Delete (One by one)", "Fast Delete (Parallel)", "Cancel"},
		Default: "Normal Delete (One by one)",
	}

	err := survey.AskOne(prompt, &mode)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	switch mode {
	case "Normal Delete (One by one)":
		deleteNormal(projects)
	case "Fast Delete (Parallel)":
		deleteGoRoutine(projects)
	default:
		fmt.Println("❌ Operation cancelled.")
	}

}

// / ***** Scan *******
func scanProjects(root string) []ProjectInfo {
	var projects []ProjectInfo
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		//error && not dir -> ကျော်မယ်
		if err != nil && !info.IsDir() {
			return nil
		}
		// dir skip
		if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == "build" {
			return filepath.SkipDir
		}
		var p ProjectInfo
		isProject := false

		//flutter
		if existsPath(filepath.Join(path, "pubspec.yaml")) {
			// flutter
			p = ProjectInfo{Type: "Flutter", Path: path}
			p.Items = append(p.Items, filepath.Join(path, "build"), filepath.Join(path, ".dart_tool"))
			isProject = true
		} else
		// node js
		if existsPath(filepath.Join(path, "package.json")) {
			p = ProjectInfo{Type: "NodeJs", Path: path}
			p.Items = append(p.Items, filepath.Join(path, "node_modules"))
			isProject = true
		} else if existsPath(filepath.Join(path, "Cargo.toml")) {
			p = ProjectInfo{Type: "Rust", Path: path}
			p.Items = append(p.Items, filepath.Join(path, "target"))
			isProject = true
		}

		if isProject {
			// size ရယူမယ်
			var pSize int64
			for _, item := range p.Items {
				if existsPath(item) {
					size, _ := dirSize(item)
					pSize += size
				}
			}
			if pSize > 0 {
				p.Size = pSize
				projects = append(projects, p)
			}
			// flutter ဖြစ်ပြီးတော့ example folder လည်းရှိရင် သူ့ထဲမှာ ဆက်ရှာမယ်
			if p.Type == "Flutter" && path == filepath.Join(path, "example") {
				return nil //ဆက်ရှာမယ်
			}
			//မဟုတ်ရင် SKIP
			return filepath.SkipDir

		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error Walk Path: %s\n", root)
	}
	return projects
}

/// ***** Delete *******

// delete normal
func deleteNormal(projects []ProjectInfo) {
	for _, p := range projects {
		for _, item := range p.Items {
			os.RemoveAll(item)
		}
	}
	fmt.Println("✅ Cleaned!")
}

// delete Goroutine
func deleteGoRoutine(projects []ProjectInfo) {
	var wg sync.WaitGroup

	fmt.Println("🚀 Parallel cleaning started...")

	for _, p := range projects {
		for _, item := range p.Items {
			// Folder ရှိမှ ဖျက်ဖို့ ထပ်စစ်တာ ပိုသေချာတယ်
			if _, err := os.Stat(item); err == nil {
				wg.Add(1)

				// Goroutine စတင်ခြင်း
				go func(path string) {
					defer wg.Done()

					err := os.RemoveAll(path)
					if err != nil {
						fmt.Printf("❌ Error removing %s: %v\n", path, err)
					} else {
						// လိုအပ်ရင် log ပြလို့ရတယ်၊ ဒါပေမဲ့ output တွေ ရှုပ်ကုန်မှာစိုးရင် ပိတ်ထားနိုင်ပါတယ်
						// fmt.Printf("🗑️ Deleted: %s\n", path)
					}
				}(item)
			}
		}
	}

	wg.Wait() // အားလုံးဖျက်ပြီးမှ နောက်တစ်ဆင့်သွားမယ်
	fmt.Println("✅ All selected caches cleaned successfully!")
}
