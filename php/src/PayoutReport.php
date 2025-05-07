<?php declare(strict_types=1);
namespace Mikael\Task1;

class PayoutReport {
    public readonly array $entries;

    public function __construct(PayoutEntry ...$entries) {
        $this->entries = $entries;
    }
    
    public function filterSum(callable $fn): int {
        $total = 0;

        foreach ($this->entries as $entry) {
            if ($fn($entry)) {
                $total += $entry->amount;
            }
        }

        return $total;
    }
}
