#!/usr/bin/env python3
import subprocess
import sys
import os
import json

def main():
    if len(sys.argv) != 2:
        print("Usage: python3 mcp_client.py '<json-rpc-request>'")
        sys.exit(1)

    binary = os.path.expanduser("~/tokencompress/tokencompress")
    request = sys.argv[1]

    proc = subprocess.Popen(
        [binary, "--mode", "mcp"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )
    out, err = proc.communicate(request + "\n")
    print(out)
    if err:
        print("Error:", err, file=sys.stderr)
    proc.wait()

if __name__ == "__main__":
    main()
