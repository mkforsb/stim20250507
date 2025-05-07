<?php declare(strict_types=1);

require __DIR__ . '/vendor/autoload.php';

use \Mikael\Task1\PayoutEntry;
use \Mikael\Task1\StrictCSVPayoutReportParser;

// Apparently fusonic/csv-reader requires streams to support fseek
function seekable_fake_url_stream(string $url) {
    $memfd = fopen("php://memory", "r+");
    $urlfd = fopen($url, "r");
    
    fwrite($memfd, stream_get_contents($urlfd));
    fclose($urlfd);

    rewind($memfd);
    return $memfd;
}

function usage($argv) {
    echo "usage: {$argv[0]} <date> <start> <end>", PHP_EOL;
    exit();
}

function main($argv) {
    $target_date = new \DateTimeImmutable()->format("Y-m-d");

    if (count($argv) != 4) {
        echo "Error: Missing argument(s).", PHP_EOL;
        usage($argv);
    }

    if (false === \DateTimeImmutable::createFromFormat("Y-m-d", $argv[1])) {
        echo "Error: Invalid date. Expected e.g 2025-05-05.", PHP_EOL;
        usage($argv);
    }

    $target_date = $argv[1];
    $start = intval($argv[2]);
    $end = intval($argv[3]);

    if ($start < 1 || $start > 100) {
        echo "Error: Invalid start. Expected integer in range [1, 100].", PHP_EOL;
        usage($argv);
    }

    if ($end < $start || $end > 100) {
        echo "Error: Invalid end. Expected integer in range [$start, 100].", PHP_EOL;
        usage($argv);
    }

    $parser = new StrictCSVPayoutReportParser\Parser();

    $total = 0;

    foreach (range($start, $end) as $n) {
        $report = $parser->parse(
            seekable_fake_url_stream("https://codetest.stim.se/payouts/$n", "r")
        );

        $total += $report->filterSum(
            fn(PayoutEntry $entry) => $entry->date->format("Y-m-d") == $target_date
        );
    }

    echo $total, PHP_EOL;
}

main($argv);

