<?php declare(strict_types=1);

use PHPUnit\Framework\TestCase;

use Mikael\Task1\StrictCSVPayoutReportParser\Parser;

class StrictCSVPayoutReportParserTest extends TestCase {
    public function testParseFirstReport(): void {
        $fd = fopen("tests/assets/1.csv", "r");
        $report = new Parser()->parse($fd);
        
        $this->assertEquals(count($report->entries), 119925);
    }
}
