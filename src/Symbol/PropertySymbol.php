<?php
declare(strict_types=1);
namespace PhpIntel\Symbol;

use PhpIntel\Symbol;
use PhpIntel\Protocol\Location;

class PropertySymbol extends Symbol
{
    /**
     * @var int
     */
    public $modifier;

    /**
     * @var string
     */
    public $scope;

    public function __construct(
        Location $location, string $name, array $types, int $modifier, string $scope
    ) {
        parent::__construct($location, $name, $types);

        $this->modifier = $modifier;
        $this->scope = $scope;
    }
}