# Prime Number Finder CLI

A command-line interface (CLI) application that finds prime numbers within specified ranges and outputs them to a file.

## Features

- Process multiple number ranges concurrently
- Output prime numbers to a specified file
- Set a timeout for the operation
- Efficient prime number calculation

## Requirements

- Go 1.22.6 or later

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/s0vunia/pyshopjl.primes.git
   cd pyshopjl.primes
   ```

2. Build the application:
   ```
   go build -o find_primes
   ```

## Usage

Run the application with the following command:

```
./find_primes --file <output_file> --timeout <seconds> --range <start>:<end> [--range <start>:<end> ...]
```

Arguments:
- `--file` or `-f`: Name of the output file (required)
- `--timeout` or `-t`: Timeout in seconds (required)
- `--range` or `-r`: Number range in format start:end (required, can be specified multiple times)

Example:
```
./find_primes --file primes.txt --timeout 30 --range 1:100 --range 1000:2000
```

This command will find prime numbers in the ranges 1-100 and 1000-2000, with a timeout of 30 seconds, and write the results to `primes.txt`.

## How It Works

1. The application processes each specified range in a separate goroutine.
2. Prime numbers found are sent to a channel.
3. A dedicated goroutine writes the prime numbers from the channel to the specified output file.
4. The operation continues until all ranges are processed, the timeout is reached, or an error occurs.

## Error Handling

- If the timeout is reached before processing is complete, the application will terminate and report a timeout.
- Any errors during file operations or number processing will be reported to the user.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
