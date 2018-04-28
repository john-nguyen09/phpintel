<?php
declare(strict_types=1);
namespace PhpIntel\Symbol;

use PhpIntel\PhpDocument;
use PhpIntel\Symbol;
use PhpIntel\Protocol\Location;

class InterfaceSymbol extends Symbol
{
    /**
     * @var string[]
     */
    public $parents;

    public function __construct(
        Location $location, string $name, array $parents
    ) {
        parent::__construct($location, $name, []);

        $this->parents = $parents;
    }

    public function resolveToFqn(PhpDocument $doc)
    {
        $this->appendNamespaceToName($doc);
    }
}