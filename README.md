# AutoComplete CLI

This is a simple command-line interface (CLI) text editor that provides autocomplete suggestions based on word frequency using a Trie data structure.

## Features
- Real-time autocomplete suggestions based on the words from `words.txt`
- Suggestions sorted by word frequency
- TAB key to cycle through suggestions
- ENTER key to select suggestion
- Backspace support
- Custom blinking autocomplete recommendation
- Ability to dynamically insert new words into the Trie
- Graceful exit on `Ctrl+C` or `ESC`

## How It Works
1. The application reads the `words.txt` file at startup.
2. Words are inserted into the Trie structure in a case-sensitive manner.
3. As the user types, the current word is extracted and matched against the Trie.
4. If suggestions are found, they are displayed with a blinking effect.
5. The user can navigate suggestions with the `TAB` key and select them with `ENTER`.
6. Typed words are automatically added to the Trie on space (`SPACE`) keypress.

## Installation

### Prerequisites
- Go 1.18 or higher

### Steps
1. Clone the repository:
   ```bash
   git clone https://github.com/username/autocomplete-cli.git
   cd autocomplete-cli
   go mod tidy
   ```
2. Add your custom words to `words.txt` in the root directory (if needed).
3. Run the application:
   ```bash
   go run main.go
   ```

## Usage
- Start typing any word.
- Wait for 1 second to see autocomplete suggestions (if any).
- Use `TAB` to navigate suggestions.
- Press `ENTER` to select a suggestion.
- Press `Ctrl+C` or `ESC` to exit the application.
