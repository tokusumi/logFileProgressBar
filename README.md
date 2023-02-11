# logFileProgressBar
Logging progress bar into file

## Setup

Install Go.

## Example - tqdm

### Setup

Install Python. Then,

```bash
$ python3 -m venv env
$ source env/bin/activate
$ python3 -m pip install tqdm
```

### Usage

```bash
python3 example/prog.py 2>&1 | go run monitor.go out.log
```

See output file (`out.log`)
