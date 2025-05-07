<?php declare(strict_types=1);
namespace Mikael\Task1;

class PayoutEntry {
    public function __construct(
        public readonly \DateTimeImmutable $date,
        public readonly string $name,
        public readonly int $amount,
    ) {}
}
