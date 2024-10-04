package main

import "fmt"

// Filter function to determine whether the window needs to change
func shouldRedraw(prevIndex, currentIndex, windowStart, windowSize int) bool {
	windowEnd := windowStart + windowSize

	// If the current index is out of the window bounds
	if currentIndex < windowStart || currentIndex >= windowEnd {
		return true
	}

	// No redraw needed if the current index is within the window bounds
	return false
}

// Function to print only items within the window
func printItems(items []string, currentIndex, windowStart, windowSize int) {
	windowEnd := windowStart + windowSize
	if windowEnd > len(items) {
		windowEnd = len(items)
	}

	fmt.Println("Items in the window:")
	for i := windowStart; i < windowEnd; i++ {
		prefix := "  "
		if i == currentIndex {
			prefix = "> "
		}
		fmt.Printf("%s%s\n", prefix, items[i])
	}
}

func main() {
	items := []string{"Item 1", "Item 2", "Item 3", "Item 4", "Item 5", "Item 6", "Item 7", "Item 8", "Item 9", "Item 10"}
	windowSize := 4

	// Cursor positions
	currentIndex := 0
	previousIndex := -1
	windowStart := 0

	// Initial window print
	printItems(items, currentIndex, windowStart, windowSize)

	// Function to simulate cursor movement
	moveCursor := func(newIndex int) {
		previousIndex = currentIndex
		currentIndex = newIndex

		// Check if window needs to be redrawn
		if shouldRedraw(previousIndex, currentIndex, windowStart, windowSize) {
			// Adjust window start position based on new cursor position
			if currentIndex < windowStart {
				windowStart = currentIndex
			} else if currentIndex >= windowStart+windowSize {
				windowStart = currentIndex - windowSize + 1
			}
			printItems(items, currentIndex, windowStart, windowSize)
		}
	}

	// Simulate some cursor movements
	fmt.Println("\nMove cursor to index 3:")
	moveCursor(3) // Moves cursor to Item 4 (no redraw needed)

	fmt.Println("\nMove cursor to index 5:")
	moveCursor(5) // Moves cursor to Item 6 (this should cause the window to shift)

	fmt.Println("\nMove cursor to index 8:")
	moveCursor(8) // Moves cursor to Item 9 (this should cause the window to shift)

	fmt.Println("\nMove cursor to index 2:")
	moveCursor(2) // Moves cursor back to Item 3 (this should cause the window to shift back)
}
