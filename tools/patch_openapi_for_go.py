#!/usr/bin/env python3
"""
Patch lighter-python/openapi.json into a Go-friendly copy before code
generation.

The Candle schema declares both upper- and lower-case single-letter keys
(C/c, H/h, L/l, O/o, V/v). Python keeps these distinct (attributes are
case-sensitive); Go field names are not, so oapi-codegen would emit two
fields named `C`, `H`, ... which fails to compile. We attach `x-go-name`
extensions so each Go field gets a distinct, descriptive name while the
JSON tag stays byte-identical to the wire format.

Usage:
    python3 tools/patch_openapi_for_go.py            # writes tools/_openapi_go.json
    python3 tools/patch_openapi_for_go.py OUT.json    # custom output path
"""

import json
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
GO_ROOT = os.path.dirname(HERE)
SRC = os.path.normpath(os.path.join(GO_ROOT, "..", "lighter-python", "openapi.json"))

# Mirror the Python field semantics (taken from each property's description).
CANDLE_GO_NAMES = {
    "t": "Timestamp",
    "o": "Open",
    "h": "High",
    "l": "Low",
    "c": "Close",
    "v": "VolumeBase",
    "V": "VolumeQuote",
    "i": "LastTradeId",
    "C": "CloseRaw",
    "H": "HighRaw",
    "L": "LowRaw",
    "O": "OpenRaw",
}


def patch(spec):
    candle = spec.get("components", {}).get("schemas", {}).get("Candle", {})
    props = candle.get("properties", {})
    for key, go_name in CANDLE_GO_NAMES.items():
        if key in props:
            props[key]["x-go-name"] = go_name
    return spec


def main():
    out = sys.argv[1] if len(sys.argv) > 1 else os.path.join(HERE, "_openapi_go.json")
    with open(SRC) as f:
        spec = json.load(f)
    patch(spec)
    with open(out, "w") as f:
        json.dump(spec, f)
    print(f"wrote {out}")


if __name__ == "__main__":
    main()
