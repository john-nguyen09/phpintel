<?php
declare(strict_types=1);
namespace PhpIntel\Symbol;

use PhpIntel\Symbol;
use PhpIntel\Protocol\Location;

class MethodSymbol extends FunctionSymbol
{
    /**
     * Modifier of the method (public, private, protected, final, abstract, static)
     *
     * @var int
     */
    public $modifier;

    /**
     * @var string
     */
    public $scope;

    public function __construct(
        Location $location, string $name, array $types, $params,
        int $modifier, string $scope
    ) {
        parent::__construct($location, $name, $types, $params);

        $this->modifier = $modifier;
        $this->scope = $scope;
    }
}