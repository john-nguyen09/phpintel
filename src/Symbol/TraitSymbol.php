<?php
declare(strict_types=1);
namespace PhpIntel\Symbol;

use PhpIntel\Symbol;
use PhpIntel\Protocol\Location;

class TraitSymbol extends Symbol
{
    // TODO: usage of other traits

    public function __construct(Location $location, string $name)
    {
        parent::__construct($location, $name, []);
    }
}