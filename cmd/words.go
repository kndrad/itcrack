/*
Copyright © 2024 Konrad Nowara

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kndrad/itcrack/internal/screenshot"
	"github.com/spf13/cobra"
)

var (
	filename string
	save     bool
	outPath  string
	verbose  bool
)

// wordsCmd represents the words command.
var wordsCmd = &cobra.Command{
	Use:   "words",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("words called.")
		filename = filepath.Clean(filename)

		exit := Exit()
		defer exit()

		// Get all screenshot files
		var files []string

		// Check if filename is a directory, if it is - process many screenshots within it.
		stat, err := os.Stat(filename)
		if err != nil {
			logger.Error("wordsCmd", "err", err)

			return fmt.Errorf("wordsCmd: %w", err)
		}
		if stat.IsDir() {
			// File represents a directory so append each screenshot file to files (with non image removal).
			logger.Info("wordsCmd: processing directory", "filename", filename)

			entries, err := os.ReadDir(filepath.Clean(filename))
			if err != nil {
				logger.Error("wordsCmd", "err", err)

				return fmt.Errorf("wordsCmd: %w", err)
			}
			// Append image files only
			for _, e := range entries {
				if !e.IsDir() && isImageFile(e.Name()) {
					files = append(files, filepath.Join(filename, "/", e.Name()))
				}
			}
			logger.Info("wordsCmd: number of image files in a directory", "len(filename)", len(files))
		} else {
			files = append(files, filename)
		}

		outFile, err := os.OpenFile(filepath.Clean(outPath), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			fmt.Println(err)

			return fmt.Errorf("wordsCmd: %w", err)
		}
		defer outFile.Close()

		// Clear outFile and reset to beginning
		if err := outFile.Truncate(0); err != nil {
			logger.Error("wordsCmd", "err", err)

			return fmt.Errorf("wordsCmd: %w", err)
		}
		if _, err := outFile.Seek(0, 0); err != nil {
			logger.Error("wordsCmd", "err", err)

			return fmt.Errorf("wordsCmd: %w", err)
		}

		// Process each screenshot and write an out file
		for _, name := range files {
			fmt.Println("filename:", name)
			content, err := os.ReadFile(filepath.Clean(name))
			if err != nil {
				logger.Error("wordsCmd", "err", err)

				return fmt.Errorf("wordsCmd: %w", err)
			}
			words, err := screenshot.RecognizeWords(content)
			if err != nil {
				logger.Error("wordsCmd", "err", err)

				return fmt.Errorf("wordsCmd: %w", err)
			}
			if err := screenshot.WriteWords(words, screenshot.NewWordsTextFileWriter(outFile)); err != nil {
				logger.Error("wordsCmd", "err", err)

				return fmt.Errorf("wordsCmd: %w", err)
			}
		}

		return nil
	},
}

// frequencyCmd represents the frequency command.
var frequencyCmd = &cobra.Command{
	Use:   "frequency",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("frequency called")
		filename = filepath.Clean(filename)
		if verbose {
			logger.Info("frequencyCmd", "filename", filename)
		}

		exit := Exit()
		defer exit()

		content, err := os.ReadFile(filename)
		if err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd err: %w", err)
		}

		// Scan each word
		scanner := bufio.NewScanner(bytes.NewReader(content))
		scanner.Split(bufio.ScanWords)

		// Filter out non existing words?
		// Or try to adjust the word to the nearest possible?
		words := make([]string, 0)
		words = append(words, "test")

		for scanner.Scan() {
			word := scanner.Text()
			words = append(words, word)
		}
		if err := scanner.Err(); err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}

		textAnalysis, err := screenshot.NewTextAnalysis("1")
		if err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}
		for _, word := range words {
			textAnalysis.Add(word)
		}

		jsonPath := filepath.Join(
			filepath.Clean(outPath),
			string(filepath.Separator),
			textAnalysis.Name()+".json",
		)
		logger.Info("frequencyCmd opening file", "jsonFilePath", jsonPath)
		jsonFile, err := OpenCleanFile(jsonPath, os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}
		defer jsonFile.Close()

		analysisJSON, err := json.MarshalIndent(textAnalysis, "", " ")
		if err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}
		logger.Info("frequencyCmd writing analysisJSON")
		if _, err := jsonFile.Write(analysisJSON); err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}

		return nil
	},
}

func OpenCleanFile(path string, flag int, perm fs.FileMode) (*os.File, error) {
	path = filepath.Clean(path)

	f, err := os.OpenFile(path, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("OpenCleanFile: %w", err)
	}
	if err := f.Truncate(0); err != nil {
		return nil, fmt.Errorf("OpenCleanFile: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("OpenCleanFile: %w", err)
	}

	return f, nil
}

func init() {
	rootCmd.AddCommand(wordsCmd)

	wordsCmd.PersistentFlags().StringVarP(&filename, "file", "f", "", "Screenshot file to recognize words from")
	if err := wordsCmd.MarkPersistentFlagRequired("file"); err != nil {
		logger.Error("rootcmd", "err", err.Error())
	}
	wordsCmd.Flags().BoolVarP(&save, "save", "s", false, "Save the output")
	wordsCmd.Flags().StringVarP(&outPath, "out", "o", "", "Output path")
	wordsCmd.MarkFlagsRequiredTogether("save", "out")
	wordsCmd.Flags().BoolVarP(&verbose, "verbose", "v", true, "Verbose")

	rootCmd.AddCommand(frequencyCmd)
	frequencyCmd.PersistentFlags().StringVarP(&filename, "file", "f", "", "File to analyze words output frequency from")
	if err := frequencyCmd.MarkPersistentFlagRequired("file"); err != nil {
		logger.Error("frequencyCmd", "err", err.Error())
	}

	frequencyCmd.Flags().StringVarP(&outPath, "out", "o", ".", "Output path")
	frequencyCmd.Flags().BoolVarP(&verbose, "verbose", "v", true, "Verbose")
}

func isImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))

	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}

func Exit(funcs ...func() error) func() error {
	return func() error {
		for _, f := range funcs {
			if err := f(); err != nil {
				return fmt.Errorf("onExit: %w", err)
			}
		}

		fmt.Println("Program is done.")
		os.Exit(1)

		return nil
	}
}
