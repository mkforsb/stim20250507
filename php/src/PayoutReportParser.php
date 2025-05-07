<?php declare(strict_types=1);
namespace Mikael\Task1;

interface PayoutReportParser {
    public function parse(mixed $report): PayoutReport;
}
