package input

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

// LineReader handles interactive line input with proper UTF-8 support
type LineReader struct {
	fd      int
	prompt  string
	builder strings.Builder
	width   int // Terminal width
}

// New creates a new LineReader
func New(prompt string) *LineReader {
	fd := int(os.Stdin.Fd())
	width := 80
	if term.IsTerminal(fd) {
		if w, _, err := term.GetSize(fd); err == nil {
			width = w
		}
	}
	return &LineReader{
		fd:     fd,
		prompt: prompt,
		width:  width,
	}
}

// ReadLine reads a line with proper UTF-8 and backspace handling
func (r *LineReader) ReadLine() (string, error) {
	r.builder.Reset()

	// Save terminal state and switch to raw mode
	oldState, err := term.MakeRaw(r.fd)
	if err != nil {
		// Fallback to simple input
		fmt.Print(r.prompt)
		var input string
		fmt.Scanln(&input)
		return input, err
	}
	defer term.Restore(r.fd, oldState)

	// Display initial prompt
	r.displayPrompt()

	buf := make([]byte, 1)

	for {
		// Read one byte
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return "", err
		}
		if n == 0 {
			continue
		}

		b := buf[0]

		switch b {
		case 3: // Ctrl+C
			fmt.Println("^C")
			return "", fmt.Errorf("interrupted")

		case 4: // Ctrl+D
			if r.builder.Len() == 0 {
				fmt.Println()
				return "exit", nil
			}

		case 13, 10: // Enter (CR/LF)
			// Handle CRLF combination
			if b == 13 {
				// Peek ahead for LF
				peek := make([]byte, 1)
				n, _ := os.Stdin.Read(peek)
				if n == 1 && peek[0] != 10 {
					// Not LF, put it back (we can't really put it back, so we'll process it next iteration)
				}
			}
			fmt.Println()
			return strings.TrimSpace(r.builder.String()), nil

		case 127, 8: // Backspace (127) or Ctrl+H (8)
			if r.builder.Len() > 0 {
				// Remove the last UTF-8 rune
				str := r.builder.String()
				runes := []rune(str)
				if len(runes) > 0 {
					// Remove last rune
					runes = runes[:len(runes)-1]
					r.builder.Reset()
					r.builder.WriteString(string(runes))

					// Clear and redisplay
					r.clearLine()
					r.displayPrompt()
					r.displayInput()
				}
			}

		case 27: // ESC - start of escape sequence (arrows, etc.)
			// Read and discard the next two bytes
			tmp := make([]byte, 2)
			os.Stdin.Read(tmp)
			// Could be extended to handle arrow keys

		case 9: // Tab - ignore for now
			// Could implement tab completion

		default:
			// Check if this could be valid UTF-8
			if b >= 32 && b <= 126 {
				// ASCII - print immediately
				r.builder.WriteByte(b)
				fmt.Print(string(b))
			} else if b >= 0x80 {
				// Multi-byte UTF-8 sequence
				// Determine UTF-8 sequence length from first byte
				seq := []byte{b}
				var seqLen int
				switch {
				case b&0xE0 == 0xC0:
					seqLen = 2
				case b&0xF0 == 0xE0:
					seqLen = 3
				case b&0xF8 == 0xF0:
					seqLen = 4
				default:
					seqLen = 1 // Invalid, skip
				}

				for i := 1; i < seqLen; i++ {
					n, err := os.Stdin.Read(buf)
					if err != nil || n == 0 {
						break
					}
					seq = append(seq, buf[0])
				}

				// Validate and decode
				if utf8.FullRune(seq) {
					rune, _ := utf8.DecodeRune(seq)
					if rune != utf8.RuneError {
						r.builder.WriteRune(rune)
						fmt.Print(string(rune))
					}
				}
			}
		}
	}
}

// displayPrompt shows the prompt
func (r *LineReader) displayPrompt() {
	fmt.Print(r.prompt)
}

// displayInput shows the current input buffer
func (r *LineReader) displayInput() {
	fmt.Print(r.builder.String())
}

// clearLine clears the current line
func (r *LineReader) clearLine() {
	// Move to beginning of line, clear to end
	fmt.Print("\r\033[K")
}

// SetPrompt changes the prompt
func (r *LineReader) SetPrompt(prompt string) {
	r.prompt = prompt
}
