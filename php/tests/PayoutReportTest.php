<?php declare(strict_types=1);

use PHPUnit\Framework\TestCase;

use Mikael\Task1\PayoutReport;
use Mikael\Task1\PayoutEntry;

function exampleReport1(): PayoutReport {
    return new PayoutReport(
        new PayoutEntry(
            \DateTimeImmutable::createFromFormat("Y-m-d", "2025-05-03"),
            "Alice",
            200,
        ),
        new PayoutEntry(
            \DateTimeImmutable::createFromFormat("Y-m-d", "2025-05-04"),
            "Bob",
            300,
        ),
        new PayoutEntry(
            \DateTimeImmutable::createFromFormat("Y-m-d", "2025-05-04"),
            "Cheyenne",
            500,
        ),
        new PayoutEntry(
            \DateTimeImmutable::createFromFormat("Y-m-d", "2025-05-05"),
            "Douglas",
            700,
        ),
    );
}

function exampleReport2(): PayoutReport {
    return new PayoutReport(
        new PayoutEntry(
            \DateTimeImmutable::createFromFormat("Y-m-d", "2025-05-03"),
            "Alice",
            200,
        ),
        new PayoutEntry(
            \DateTimeImmutable::createFromFormat("Y-m-d", "2025-05-04"),
            "Bob",
            300,
        ),
    );
}

class PayoutReportTest extends TestCase {
    public function testFilterSum(): void {
        $this->assertEquals(exampleReport1()->filterSum(fn(PayoutEntry $entry) => true), 1700);
        $this->assertEquals(exampleReport2()->filterSum(fn(PayoutEntry $entry) => true), 500);
    }
}
