<?php
declare(strict_types=1);
namespace PhpIntel\Symbol;

use PhpIntel\Protocol\Location;

class ClassConstantSymbol extends ConstantSymbol
{
    /**
     * Visibility modifier
     *
     * @var int
     */
    public $modifier;

    /**
     * The class, interface or trait name
     *
     * @var string
     */
    public $scope;

    public function __construct(
        Location $location, string $name, string $type, $value, int $modifier, string $scope
    ) {
        parent::__construct($location, $name, $type, $value);

        $this->modifier = $modifier;
        $this->scope = $scope;
    }
}