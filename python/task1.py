import csv
from dataclasses import dataclass
import datetime
import requests
from typing import Callable, Dict, List, Generator
import unittest


@dataclass
class PayoutEntry:
    date: datetime.date
    name: str
    amount: int

    def validate(self) -> "PayoutEntry":
        if self.amount < 0:
            raise ValueError("invalid payout entry: negative amount")

        return self


@dataclass
class PayoutReport:
    entries: List[PayoutEntry]

    def filter(self, f: Callable[[PayoutEntry], bool]) -> "PayoutReport":
        return PayoutReport(list(filter(f, self.entries)))

    def sum(self) -> int:
        return sum([entry.amount for entry in self.entries])


class StrictCSVPayoutReportParser:
    def parse(self, text: str) -> PayoutReport:
        """
        Parse a payout report.

        Throws ValueError on parse failure.
        """
        fieldset = {"date", "name", "amount"}
        reader = csv.DictReader(text.splitlines())

        try:
            first = next(reader)
        except StopIteration:
            raise ValueError("empty input")

        # Note that this actually tolerates permutations of the field order, a feature that
        # is perhaps more surprising/confusing than useful.

        if not set(first.keys()) == fieldset:
            raise ValueError(
                f"field set mismatch: required {fieldset} "
                f"but found {set(first.keys())}"
            )

        def parse_entry(entry: Dict[str, str]) -> PayoutEntry:
            return PayoutEntry(
                date=datetime.date.fromisoformat(entry["date"]),  # may ValueError
                name=entry["name"],
                amount=int(entry["amount"]),  # may ValueError
            )

        return PayoutReport(
            [parse_entry(first).validate()]
            + [parse_entry(entry).validate() for entry in reader]
        )


if __name__ == "__main__":
    import sys

    if len(sys.argv) < 2:
        print(f"usage: {sys.argv[0]} <date>")
        print(f"example: {sys.argv[0]} {datetime.date.today()}")
        exit()

    def reports() -> Generator[PayoutReport]:
        for n in range(1, 101):
            with requests.get(
                f"https://codetest.stim.se/payouts/{n}",
                headers={"User-Agent": "Mozilla/Firefox"},
            ) as fd:
                yield StrictCSVPayoutReportParser().parse(fd.text)

    print(
        sum(
            [
                report.filter(
                    lambda entry: entry.date == datetime.date.fromisoformat(sys.argv[1])
                ).sum()
                for report in reports()
            ]
        )
    )


class Tests(unittest.TestCase):
    def test_strict_csv_payout_report_parser_parse_failure(self):
        with self.assertRaises(ValueError):
            StrictCSVPayoutReportParser().parse("")

        with self.assertRaises(ValueError):
            StrictCSVPayoutReportParser().parse("color,brand,size\nblue,x,large\n")

        with self.assertRaises(ValueError):
            StrictCSVPayoutReportParser().parse(
                "date,name,amount\n2025-05-05,Alice,-2\n"
            )

    def test_strict_csv_payout_report_parser_parse_success(self):
        report = StrictCSVPayoutReportParser().parse(
            "date,name,amount\n"
            "2025-05-05,Johnathon Reichert,1389\n"
            "2025-05-03,Martina Leuschke,814\n"
        )

        self.assertEqual(
            report.entries[0],
            PayoutEntry(
                datetime.date.fromisoformat("2025-05-05"), "Johnathon Reichert", 1389
            ),
        )

        self.assertEqual(
            report.entries[1],
            PayoutEntry(
                datetime.date.fromisoformat("2025-05-03"), "Martina Leuschke", 814
            ),
        )

    def test_payout_report_filter_sum(self):
        report = PayoutReport(
            [
                PayoutEntry(datetime.date.fromisoformat("2004-05-19"), "Alice", 100),
                PayoutEntry(datetime.date.fromisoformat("2004-05-19"), "Bob", 200),
                PayoutEntry(datetime.date.fromisoformat("2007-01-11"), "Charles", 400),
            ]
        )

        self.assertEqual(report.sum(), 700)

        self.assertEqual(
            report.filter(
                lambda entry: entry.date == datetime.date.fromisoformat("2004-05-19")
            ).sum(),
            300,
        )

        self.assertEqual(
            report.filter(
                lambda entry: entry.date == datetime.date.fromisoformat("2007-01-11")
            ).sum(),
            400,
        )
