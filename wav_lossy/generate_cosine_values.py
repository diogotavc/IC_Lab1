import math
from itertools import product
from struct import Struct
from argparse import ArgumentParser
from pathlib import Path

FLOAT_STRUCT = Struct("<d")

def main() -> None:
    parser = ArgumentParser()
    parser.add_argument("chunk_size", type=int)
    parser.add_argument("output_file", type=Path)

    args = parser.parse_args()

    with args.output_file.open("wb") as output_file:
        for f, i in product(range(args.chunk_size), repeat=2):
            output_file.write(FLOAT_STRUCT.pack(math.cos(f * (i + .5) * math.pi / args.chunk_size)))


if __name__ == "__main__":
    main()
