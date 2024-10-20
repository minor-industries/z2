from pprint import pprint

import indoor_bike_data


def main():
    s = bytes.fromhex("74 03 2a 08 1c 00 3b 00  00 3b 00 1c 00 01 00 00")
    bd = indoor_bike_data.parse_indoor_bike_data(s)
    pprint(bd._asdict())


if __name__ == '__main__':
    main()
