package sys

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var SizeCmd = &cobra.Command{
	Use:   "size",
	Short: "Lists the size of all files and folders in the current directory.",
	Long: `Displays the disk usage of every entry (including hidden files and folders)
in the current working directory, sorted by size descending.`,
	Annotations: map[string]string{
		"title": "Directory Size",
	},
	Run: func(cmd *cobra.Command, args []string) {
		runSize()
	},
}

type sizeEntry struct {
	name  string
	size  int64
	isDir bool
}

type walkTask struct {
	path          string
	topLevelIndex int
}

func runSize() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting current directory: %v\n", err)
		return
	}

	entries, err := os.ReadDir(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading directory: %v\n", err)
		return
	}

	start := time.Now()

	numWorkers := runtime.NumCPU()
	if numWorkers < 4 {
		numWorkers = 4
	} else if numWorkers > 16 {
		numWorkers = 16
	}

	sizes := make([]int64, len(entries))
	var wg sync.WaitGroup
	tasks := make(chan walkTask, 200000)

	for workerID := 1; workerID <= numWorkers; workerID++ {
		go func(id int) {
			for task := range tasks {
				if isDevMode {
					fmt.Printf("[DEBUG] worker%d: scanning %s\n", id, task.path)
				}
				subEntries, err := os.ReadDir(task.path)
				if err != nil {
					wg.Done()
					continue
				}
				for _, se := range subEntries {
					fullPath := filepath.Join(task.path, se.Name())
					if se.IsDir() {
						wg.Add(1)
						newTask := walkTask{path: fullPath, topLevelIndex: task.topLevelIndex}
						select {
						case tasks <- newTask:
						default:
							go func(t walkTask) {
								tasks <- t
							}(newTask)
						}
					} else {
						info, err := se.Info()
						if err == nil {
							atomic.AddInt64(&sizes[task.topLevelIndex], info.Size())
						}
					}
				}
				wg.Done()
			}
		}(workerID)
	}

	// Initialize the queue with top-level directories
	for i, e := range entries {
		fullPath := filepath.Join(cwd, e.Name())
		if e.IsDir() {
			wg.Add(1)
			tasks <- walkTask{path: fullPath, topLevelIndex: i}
		} else {
			info, err := e.Info()
			if err == nil {
				sizes[i] = info.Size()
			}
		}
	}

	// Wait for all subdirectories to finish processing
	wg.Wait()
	close(tasks)

	// Assemble final sizeEntry slice
	items := make([]sizeEntry, len(entries))
	for i, e := range entries {
		items[i] = sizeEntry{
			name:  e.Name(),
			size:  sizes[i],
			isDir: e.IsDir(),
		}
	}

	elapsed := time.Since(start)

	sort.Slice(items, func(i, j int) bool {
		return items[i].size > items[j].size
	})

	timeStr := ui.YellowStyle().Render(elapsed.Round(time.Millisecond).String())
	fmt.Printf("%s Calculated in %s\n\n", ui.HourglassIcon, timeStr)

	header := ui.BoldStyle.Render(fmt.Sprintf("%-10s  %-6s  %s", "SIZE", "TYPE", "NAME"))
	fmt.Println(header)
	fmt.Println(ui.GrayStyle().Render(strings.Repeat("-", 50)))

	for _, item := range items {
		var kindStr, nameStr, sizeStr string

		// Pad raw strings first, then colorize — avoids ANSI codes breaking %-Ns alignment
		rawSize := fmt.Sprintf("%-10s", formatSize(item.size))
		sizeStr = coloredSize(rawSize, item.size)

		if item.isDir {
			kindStr = ui.CyanStyle().Render(fmt.Sprintf("%-6s", "dir"))
			nameStr = ui.CyanStyle().Bold(true).Render(item.name + "/")
		} else {
			kindStr = ui.GrayStyle().Render(fmt.Sprintf("%-6s", "file"))
			nameStr = item.name
		}
		fmt.Printf("%s  %s  %s\n", sizeStr, kindStr, nameStr)
	}
}

func calcDirSize(path string) int64 {
	var total int64
	filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				total += info.Size()
			}
		}
		// return nil for WalkDir to continue to loop inside this dir
		return nil
	})
	return total
}

func formatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func coloredSize(s string, b int64) string {
	const mb = 1024 * 1024
	var style lipgloss.Style
	switch {
	case b >= 100*mb:
		style = ui.RedStyle()
	case b >= mb:
		style = ui.YellowStyle()
	default:
		style = ui.GreenStyle()
	}
	return style.Render(s)
}
