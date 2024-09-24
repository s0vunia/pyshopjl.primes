package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var (
	outputFile string
	timeout    int
	ranges     []string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "find_primes",
	Short: "Find prime numbers in given ranges",
	Long:  `A console utility to find prime numbers in specified ranges and output them to a file.`,
	Run:   run,
}

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "file", "f", "", "Output file name (required)")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "t", 0, "Timeout in seconds (required)")
	rootCmd.Flags().StringArrayVarP(&ranges, "range", "r", []string{}, "Number range in format start:end (required)")

	for _, flag := range []string{"file", "timeout", "range"} {
		if err := rootCmd.MarkFlagRequired(flag); err != nil {
			fmt.Printf("Error marking flag '%s' as required: %v\n", flag, err)
			os.Exit(1)
		}
	}
}

func run(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	resultChan := make(chan int, 100)
	errorChan := make(chan error, 1)
	processingDone := make(chan struct{})
	wg1 := &sync.WaitGroup{}

	wg1.Add(1)
	go func() {
		defer wg1.Done()
		if err := writeResults(ctx, resultChan, outputFile); err != nil {
			select {
			case errorChan <- fmt.Errorf("error writing results: %w", err):
			default:
			}
		}
	}()

	go func() {
		wg1.Wait()
		processingDone <- struct{}{}
		close(processingDone)
	}()

	wg2 := &sync.WaitGroup{}
	for _, r := range ranges {
		r := r
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			processRange(ctx, r, resultChan)
		}()
	}

	go func() {
		wg2.Wait()
		close(resultChan)
	}()

	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			fmt.Println("Operation timed out")
		} else {
			fmt.Println("Operation was cancelled")
		}
	case err := <-errorChan:
		fmt.Printf("Error occurred: %v\n", err)
	case <-processingDone:
		fmt.Println("All results processed and written")
	}
}

func processRange(ctx context.Context, r string, resultChan chan<- int) {
	start, end, err := parseRange(r)
	if err != nil {
		fmt.Printf("Error parsing range %s: %v\n", r, err)
		return
	}

	for num := start; num <= end; num++ {
		select {
		case <-ctx.Done():
			return
		default:
			if isPrime(num) {
				resultChan <- num
			}
		}
	}
}

func parseRange(r string) (int, int, error) {
	parts := strings.Split(r, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format")
	}

	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start number")
	}

	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end number")
	}

	if start > end {
		return 0, 0, fmt.Errorf("start number must be less than or equal to end number")
	}

	return start, end, nil
}

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	for i := 2; i <= int(math.Sqrt(float64(n))); i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func writeResults(ctx context.Context, resultChan <-chan int, outputFile string) (err error) {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer func() {
		closeErr := file.Close()
		err = errors.Join(err, closeErr)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case prime, ok := <-resultChan:
			if !ok {
				return nil
			}
			if _, writeErr := fmt.Fprintln(file, prime); writeErr != nil {
				return fmt.Errorf("error writing to file: %w", writeErr)
			}
		}
	}
}
