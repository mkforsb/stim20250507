<?php declare(strict_types=1);
namespace Mikael\Task1\StrictCSVPayoutReportParser;

use Fusonic\CsvReader\CsvReader;
use Mikael\Task1\PayoutEntry;
use Mikael\Task1\PayoutReportParser;
use Mikael\Task1\PayoutReport;

class Parser implements PayoutReportParser {
    public function parse(mixed $report): PayoutReport {
        $reader = new CsvReader($report);
        $entries = [];

        foreach ($reader->readObjects(Record::class) as $item) {
            array_push($entries, new PayoutEntry($item->date, $item->name, $item->amount));
        }

        return new PayoutReport(...$entries);
    }
}
