package node

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// Reset ANSI sequence
const resetSequence = "\033[0m"

const (
	BLACK_COLOR = iota
	RED_COLOR
	GREEN_COLOR
	YELLOW_COLOR
	BLUE_COLOR
	MAGENTA_COLOR
	CYAN_COLOR
	WHITE_COLOR
)

type EchelonNode struct {
	lock                sync.RWMutex
	done                sync.WaitGroup
	title               string
	titleColor          int
	description         []string
	maxDescriptionLines int
	startTime           time.Time
	endTime             time.Time
	children            []*EchelonNode
}

func StartNewEchelonNode(title string) *EchelonNode {
	result := &EchelonNode{
		title:               title,
		titleColor:          -1,
		description:         make([]string, 0),
		maxDescriptionLines: 5,
		startTime:           time.Now(),
		endTime:             time.Unix(0, 0),
		children:            make([]*EchelonNode, 0),
	}
	result.done.Add(1)
	return result
}

func (node *EchelonNode) UpdateTitle(text string) {
	node.lock.Lock()
	defer node.lock.Unlock()
	node.title = text
}

func (node *EchelonNode) ClearDescription() {
	node.SetDescription(make([]string, 0))
}

func (node *EchelonNode) SetDescription(description []string) {
	node.lock.Lock()
	defer node.lock.Unlock()
	node.description = description
}

func (node *EchelonNode) AppendDescription(text string) {
	node.lock.Lock()
	defer node.lock.Unlock()
	node.description = append(node.description, text)
	linesTotal := len(node.description)
	if linesTotal > node.maxDescriptionLines {
		node.description = node.description[(linesTotal - node.maxDescriptionLines):]
	}
}

func (node *EchelonNode) Draw() []string {
	node.lock.RLock()
	defer node.lock.RUnlock()
	result := []string{node.fancyTitle()}
	if len(node.children) > 0 {
		for _, child := range node.children {
			for _, childDescriptionLine := range child.Draw() {
				result = append(result, "  "+childDescriptionLine)
			}
		}
	} else {
		for _, descriptionLine := range node.description {
			result = append(result, "  "+descriptionLine)
		}
	}

	return result
}

func (node *EchelonNode) fancyTitle() string {
	node.lock.RLock()
	defer node.lock.RUnlock()
	prefix := "[+]"
	if node.IsRunning() {
		prefix = "[-]"
	}
	coloredTitle := fmt.Sprintf("%s%s%s", getColorSequence(node.titleColor), node.title, resetSequence)
	return fmt.Sprintf("%s %s %s", prefix, coloredTitle, formatDuration(node.ExecutionDuration()))
}

func formatDuration(duration time.Duration) string {
	if duration < 10*time.Second {
		return fmt.Sprintf("%.1fs", float64(duration.Milliseconds())/1000)
	}
	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(math.Floor(duration.Seconds())))
	}
	if duration < time.Hour {
		return fmt.Sprintf("%02d:%02d", int(math.Floor(duration.Minutes())), int(math.Floor(duration.Seconds())))
	}
	return fmt.Sprintf(
		"%02d:%02d:%02d",
		int(math.Floor(duration.Hours())),
		int(math.Floor(duration.Minutes())),
		int(math.Floor(duration.Seconds())),
	)
}

func (node *EchelonNode) ExecutionDuration() time.Duration {
	node.lock.RLock()
	defer node.lock.RUnlock()
	if node.IsRunning() {
		return time.Now().Sub(node.startTime)
	} else {
		return node.endTime.Sub(node.startTime)
	}
}

func (node *EchelonNode) IsRunning() bool {
	node.lock.RLock()
	defer node.lock.RUnlock()
	return node.endTime.Before(node.startTime)
}

func (node *EchelonNode) AddNewChild(child *EchelonNode) {
	node.lock.Lock()
	defer node.lock.Unlock()
	node.children = append(node.children, child)
}

func (node *EchelonNode) Complete() {
	node.CompleteWithColor(-1)
}
func (node *EchelonNode) CompleteWithColor(ansiColor int) {
	node.lock.Lock()
	node.endTime = time.Now()
	node.titleColor = ansiColor
	node.lock.Unlock()
	node.done.Done()
}

func (node *EchelonNode) Wait() {
	node.done.Wait()
}

func getColorSequence(code int) string {
	if code < 0 {
		return resetSequence
	}
	return fmt.Sprintf("\033[3%dm", code)
}