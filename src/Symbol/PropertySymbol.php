<?php
declare(strict_types=1);
namespace PhpIntel\Symbol;

use PhpIntel\PhpDocument;
use PhpIntel\Symbol;
use PhpIntel\Protocol\Location;

class PropertySymbol extends Symbol
{
    use SymbolWithType;

    /**
     * @var int
     */
    public $modifier;

    /**
     * @var string
     */
    public $scope;

    /**
     * Fully qualified name
     *
     * @var string[]
     */
    public $types;

    public function __construct(
        Location $location, string $name, array $types, int $modifier, string $scope
    ) {
        parent::__construct($location, $name);

        $this->types = $types;
        $this->modifier = $modifier;
        $this->scope = $scope;
    }

    public function resolveToFqn(PhpDocument $doc)
    {
        $this->scope = $this->aliasToFqn($doc, $alias);
        $this->resolveTypesToFqn($doc);
    }
}