<?php
declare(strict_types=1);
namespace PhpIntel\Symbol;

use PhpIntel\Symbol;
use PhpIntel\Protocol\Location;

class ConstantSymbol extends Symbol
{
    /**
     * @var string
     */
    public $value;

    public function __construct(
        Location $location,
        string $name,
        string $type,
        string $value
    ) {
        parent::__construct($location, $name, [$type]);

        $this->value = $value;
    }
}