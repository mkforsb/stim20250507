<?php declare(strict_types=1);
namespace Mikael\Task1\StrictCSVPayoutReportParser;

use Fusonic\CsvReader\Attributes\TitleMapping;

class Record {
    #[TitleMapping("date")]
    public \DateTimeImmutable $date;

    #[TitleMapping("name")]
    public string $name;

    #[TitleMapping("amount")]
    public int $amount; 
}
